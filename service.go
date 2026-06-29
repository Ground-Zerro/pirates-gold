package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	serviceName = "pirates-gold"
	servicePath = "/etc/systemd/system/pirates-gold.service"
)

func handleService(outDir string) {
	if serviceInstalled() {
		if err := removeService(); err != nil {
			fmt.Fprintf(os.Stderr, "error removing service: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Service pirates-gold removed.")
		return
	}

	if err := installService(outDir); err != nil {
		fmt.Fprintf(os.Stderr, "error installing service: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Service pirates-gold v.%s installed and enabled for autostart.\n\n", version)
	fmt.Printf("  Start   →  systemctl start %s\n", serviceName)
	fmt.Printf("  Stop    →  systemctl stop %s\n", serviceName)
	fmt.Printf("  Status  →  systemctl status %s\n", serviceName)
	fmt.Printf("  Logs    →  journalctl -u %s -f\n", serviceName)
}

func serviceInstalled() bool {
	_, err := os.Stat(servicePath)
	return err == nil
}

func installService(outDir string) error {
	binPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("binary path: %w", err)
	}
	binPath, err = filepath.EvalSymlinks(binPath)
	if err != nil {
		return fmt.Errorf("symlink resolution: %w", err)
	}

	absOut, err := filepath.Abs(outDir)
	if err != nil {
		return fmt.Errorf("absolute -out path: %w", err)
	}

	args := []string{binPath}
	for _, a := range os.Args[1:] {
		if a != "--service" && a != "-service" {
			args = append(args, a)
		}
	}
	if !containsFlag(args, "-out", "--out") {
		args = append(args, "-out", absOut)
	}
	if !containsFlag(args, "-workers", "--workers") {
		args = append(args, "-workers", fmt.Sprintf("%d", defaultWorkers))
	}
	if !containsFlag(args, "-rate", "--rate") {
		args = append(args, "-rate", fmt.Sprintf("%.1f", defaultRate))
	}

	content := fmt.Sprintf(`[Unit]
Description=Pirates Gold v%s — Bitcoin seed phrase scanner
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=%s
WorkingDirectory=%s
Restart=on-failure
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
`, version, strings.Join(args, " "), absOut)

	if err := os.WriteFile(servicePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write %s: %w", servicePath, err)
	}

	sysctlRun("daemon-reload")
	return sysctlRun("enable", serviceName)
}

func removeService() error {
	sysctlRun("stop", serviceName)
	sysctlRun("disable", serviceName)
	if err := os.Remove(servicePath); err != nil && !os.IsNotExist(err) {
		return err
	}
	sysctlRun("daemon-reload")
	return nil
}

func sysctlRun(args ...string) error {
	return exec.Command("systemctl", args...).Run()
}

func containsFlag(args []string, flags ...string) bool {
	for _, a := range args {
		for _, f := range flags {
			if a == f {
				return true
			}
		}
	}
	return false
}
