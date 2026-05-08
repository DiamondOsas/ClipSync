package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// EnsureAppInPath checks if the application's directory is in the system PATH.
// If not, it attempts to add it.
func EnsureAppInPath() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	dir := filepath.Dir(exe)

	// Check current process PATH
	pathEnv := os.Getenv("PATH")
	paths := strings.Split(pathEnv, string(os.PathListSeparator))
	for _, p := range paths {
		if strings.EqualFold(filepath.Clean(p), filepath.Clean(dir)) {
			return nil // Already in PATH
		}
	}

	switch runtime.GOOS {
	case "windows":
		return addToPathWindows(dir)
	case "darwin", "linux":
		return addToPathUnix(dir)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func addToPathWindows(dir string) error {
	psCmd := fmt.Sprintf(`
$regPath = "HKCU:\Environment"
$name = "Path"
$currentPath = (Get-ItemProperty -Path $regPath -Name $name -ErrorAction SilentlyContinue).$name
if ($currentPath -eq $null) {
    Set-ItemProperty -Path $regPath -Name $name -Value "%s"
} else {
    $paths = $currentPath -split ";"
    $found = $false
    foreach ($p in $paths) {
        if ($p -eq "%s") {
            $found = $true
            break
        }
    }
    if (-not $found) {
        $newPath = $currentPath
        if (-not $newPath.EndsWith(";")) {
            $newPath += ";"
        }
        $newPath += "%s"
        Set-ItemProperty -Path $regPath -Name $name -Value $newPath
    }
}
`, dir, dir, dir)

	cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to update PATH via PowerShell: %v, output: %s", err, string(out))
	}
	return nil
}

func addToPathUnix(dir string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	filesToUpdate := []string{
		filepath.Join(homeDir, ".bashrc"),
		filepath.Join(homeDir, ".zshrc"),
		filepath.Join(homeDir, ".profile"),
	}

	exportCmd := fmt.Sprintf("\nexport PATH=\"$PATH:%s\"\n", dir)

	added := false
	for _, file := range filesToUpdate {
		if _, err := os.Stat(file); err == nil {
			content, err := os.ReadFile(file)
			if err != nil {
				continue
			}

			contentStr := string(content)
			if !strings.Contains(contentStr, dir) {
				f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0600)
				if err != nil {
					continue
				}
				if _, err := f.WriteString(exportCmd); err == nil {
					added = true
				}
				f.Close()
			} else {
				// We assume it's already added if dir is mentioned
				added = true
			}
		}
	}

	if !added {
		profilePath := filepath.Join(homeDir, ".profile")
		f, err := os.OpenFile(profilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err == nil {
			f.WriteString(exportCmd)
			f.Close()
		}
	}

	return nil
}
