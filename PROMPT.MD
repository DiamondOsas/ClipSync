package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func firewallRuleExists() bool {
	cmd := exec.Command("netsh", "advfirewall", "firewall", "show", "rule", "name=ClipSync-mDNS")
	out, _ := cmd.CombinedOutput()
	return strings.Contains(string(out), "ClipSync-mDNS")
}

func addFirewallRule() error {
	cmd := exec.Command("powershell", "-Command", "Start-Process", "powershell", "-ArgumentList", `
		New-NetFirewallRule -DisplayName "ClipSync-mDNS" -Direction Inbound -Protocol UDP -LocalPort 5353 -Action Allow
	`, "-Verb", "RunAs")
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func ensureFirewall() {
	if firewallRuleExists() {
		return
	}
	fmt.Println("ClipSync needs to open UDP 5353 for device discovery. Allow? (y/N)")
	var ans string
	fmt.Scanln(&ans)
	if strings.ToLower(ans) == "y" {
		if err := addFirewallRule(); err != nil {
			fmt.Println("Could not add rule:", err)
		}
	}
}