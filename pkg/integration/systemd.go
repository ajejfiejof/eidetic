package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func getServicePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "systemd", "user")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "eidetic.service"), nil
}

// InstallService installs and enables the systemd user service for the background watcher.
func InstallService() error {
	binPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not determine executable path: %w", err)
	}

	servicePath, err := getServicePath()
	if err != nil {
		return err
	}

	content := fmt.Sprintf(`[Unit]
Description=Eidetic - Sovereign Cognitive Memory Daemon
After=graphical-session.target

[Service]
Type=simple
ExecStart=%s watch
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=default.target
`, binPath)

	if err := os.WriteFile(servicePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	_ = exec.Command("systemctl", "--user", "daemon-reload").Run()
	if err := exec.Command("systemctl", "--user", "enable", "--now", "eidetic.service").Run(); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}

	return nil
}

// UninstallService disables and deletes the systemd user service.
func UninstallService() error {
	_ = exec.Command("systemctl", "--user", "disable", "--now", "eidetic.service").Run()

	servicePath, err := getServicePath()
	if err != nil {
		return err
	}

	if err := os.Remove(servicePath); err != nil && !os.IsNotExist(err) {
		return err
	}

	_ = exec.Command("systemctl", "--user", "daemon-reload").Run()
	return nil
}

// ServiceStatus reports whether the service is currently running.
func ServiceStatus() string {
	out, err := exec.Command("systemctl", "--user", "is-active", "eidetic.service").Output()
	if err != nil {
		return "inactive / not installed"
	}
	return strings.TrimSpace(string(out))
}
