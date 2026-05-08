package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"clipsync/internal/core"
	"clipsync/internal/globals"
	"clipsync/internal/network"

	"github.com/mattn/go-isatty"
)

const IPC_PORT = 9998
const asciiArt = `
   ___ _ _      ___                 
  / __| (_)___ / __| _  _ _ _  __   
 | (__| | | _ \ \__ \ || | ' \/ _|  
  \___|_|_| .__/|___/\_, |_||_\__|  
          |_|        |__/           
`

// Run evaluates whether to run as CLI/Daemon or GUI.
// Returns true if execution was handled by CLI and main should exit.
func Run() bool {
	isTerm := isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
	isTermIn := isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())

	hasArgs := len(os.Args) > 1

	// If not running in a terminal and no arguments were passed, fallback to GUI mode.
	if !isTerm && !isTermIn && !hasArgs {
		return false
	}

	// Check if this is the internal daemon flag
	if hasArgs && os.Args[1] == "--daemon" {
		runDaemon()
		return true
	}

	fmt.Println(asciiArt)

	if !hasArgs {
		// If started from terminal with no arguments, start the daemon and exit
		startDaemon()
		return true
	}

	cmd := strings.ToLower(os.Args[1])

	switch cmd {
	case "start", "--start", "-start":
		startDaemon()
	case "list-devices", "--list-devices", "-list-devices":
		listDevices()
	case "connect", "--connect", "-connect":
		ip := ""
		if len(os.Args) > 2 {
			ip = os.Args[2]
		}
		if ip == "" {
			fmt.Println("Please provide an IP. Usage: clipsync connect <ip>")
			return true
		}
		connectToDevice(ip)
	case "stop", "--stop", "-stop":
		stopDaemon()
	case "help", "--help", "-help", "-h":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printHelp()
	}

	return true
}

func printHelp() {
	fmt.Println("Usage: clipsync [command]")
	fmt.Println("\nCommands:")
	fmt.Println("  start          Start the background daemon (default if no args)")
	fmt.Println("  list-devices   List all discovered devices")
	fmt.Println("  connect <ip>   Manually connect to a device by IP")
	fmt.Println("  stop           Stop the background daemon")
	fmt.Println("  help           Show this help menu")
}

func isDaemonRunning() bool {
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/status", IPC_PORT))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func startDaemon() {
	if isDaemonRunning() {
		fmt.Println("[*] ClipSync background daemon is already running.")
		return
	}

	fmt.Println("[*] Starting ClipSync in the background...")
	
	exePath, err := os.Executable()
	if err != nil {
		exePath = os.Args[0]
	}

	cmd := exec.Command(exePath, "--daemon")
	
	// Create detached process on Windows
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000} // CREATE_NO_WINDOW
	}

	err = cmd.Start()
	if err != nil {
		fmt.Printf("[-] Failed to start daemon: %v\n", err)
		return
	}
	fmt.Printf("[+] ClipSync daemon started successfully (PID: %d).\n", cmd.Process.Pid)
}

func runDaemon() {
	// Setup user friendly logging for daemon
	logDir := filepath.Join(os.TempDir(), "clipsync_logs")
	os.MkdirAll(logDir, 0755)
	logFile, err := os.OpenFile(filepath.Join(logDir, "daemon.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	}
	log.Println("=======================================")
	log.Println("Starting ClipSync Daemon")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go startIPCServer(cancel)

	err = core.StartSync(ctx)
	if err != nil && err != context.Canceled {
		log.Fatalf("Daemon exited with error: %v", err)
	}
	log.Println("Daemon gracefully stopped.")
}

func startIPCServer(cancelFunc context.CancelFunc) {
	mux := http.NewServeMux()

	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("/devices", func(w http.ResponseWriter, r *http.Request) {
		globals.ConnDevicesMu.Lock()
		defer globals.ConnDevicesMu.Unlock()
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(globals.ConnDevices)
	})

	mux.HandleFunc("/connect", func(w http.ResponseWriter, r *http.Request) {
		ip := r.URL.Query().Get("ip")
		if ip == "" {
			http.Error(w, "Missing 'ip' parameter", http.StatusBadRequest)
			return
		}
		
		// Attempt connection
		log.Printf("[IPC] Connecting manually to: %s", ip)
		network.Connect(ip)
		
		// Also add it to our IPS list if not already present
		globals.IPSMu.Lock()
		found := false
		for _, existingIP := range globals.IPS {
			if existingIP == ip {
				found = true
				break
			}
		}
		if !found {
			globals.IPS = append(globals.IPS, ip)
		}
		globals.IPSMu.Unlock()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Connection request sent"))
	})

	mux.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Stopping daemon..."))
		
		// Run cancellation in a goroutine so the response can be sent back to CLI
		go func() {
			cancelFunc()
			os.Exit(0)
		}()
	})

	server := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", IPC_PORT),
		Handler: mux,
	}

	log.Printf("Starting IPC server on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("IPC server failed: %v", err)
	}
}

func listDevices() {
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/devices", IPC_PORT))
	if err != nil {
		fmt.Println("[-] Failed to contact daemon. Is it running? (Try 'clipsync start')")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("[-] Failed to get devices (Status: %d)\n", resp.StatusCode)
		return
	}

	var devices []globals.Device
	if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
		fmt.Println("[-] Failed to parse response from daemon.")
		return
	}

	if len(devices) == 0 {
		fmt.Println("[*] No devices discovered yet.")
		return
	}

	fmt.Println("[*] Connected Devices:")
	for i, dev := range devices {
		fmt.Printf("  %d. %s (IP: %s)\n", i+1, dev.Name, dev.Ip)
	}
}

func connectToDevice(ip string) {
	resp, err := http.Post(fmt.Sprintf("http://127.0.0.1:%d/connect?ip=%s", IPC_PORT, ip), "application/json", nil)
	if err != nil {
		fmt.Println("[-] Failed to contact daemon. Is it running? (Try 'clipsync start')")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("[+] Connection request sent to %s.\n", ip)
	} else {
		fmt.Printf("[-] Failed to send connection request (Status: %d).\n", resp.StatusCode)
	}
}

func stopDaemon() {
	resp, err := http.Post(fmt.Sprintf("http://127.0.0.1:%d/stop", IPC_PORT), "application/json", nil)
	if err != nil {
		fmt.Println("[-] Daemon is not running.")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("[+] Daemon stopped successfully.")
	} else {
		fmt.Printf("[-] Failed to stop daemon (Status: %d).\n", resp.StatusCode)
	}
}
