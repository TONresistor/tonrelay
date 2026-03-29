package service

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const serviceName = "tonrelay"

func Start() error {
	return systemctl("start", serviceName)
}

func Stop() error {
	return systemctl("stop", serviceName)
}

func Restart() error {
	return systemctl("restart", serviceName)
}

func Disable() error {
	return systemctl("disable", serviceName)
}

func IsActive() bool {
	err := exec.Command("systemctl", "is-active", "--quiet", serviceName).Run()
	return err == nil
}

func IsEnabled() bool {
	err := exec.Command("systemctl", "is-enabled", "--quiet", serviceName).Run()
	return err == nil
}

func IsInstalled() bool {
	out, err := exec.Command("systemctl", "cat", serviceName).CombinedOutput()
	if err != nil {
		return false
	}
	return len(out) > 0
}

type Status struct {
	Active     bool
	Enabled    bool
	SubState   string
	MainPID    string
	ActiveTime string
}

func GetStatus() (*Status, error) {
	s := &Status{
		Active:  IsActive(),
		Enabled: IsEnabled(),
	}

	out, _ := exec.Command("systemctl", "show", serviceName,
		"--property=SubState,MainPID,ActiveEnterTimestamp").CombinedOutput()

	for _, line := range strings.Split(string(out), "\n") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		switch parts[0] {
		case "SubState":
			s.SubState = parts[1]
		case "MainPID":
			s.MainPID = parts[1]
		case "ActiveEnterTimestamp":
			s.ActiveTime = parts[1]
		}
	}

	return s, nil
}

func HealthCheck(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if IsActive() {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("service did not become active within %s", timeout)
}

func systemctl(action, service string) error {
	cmd := exec.Command("systemctl", action, service)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("systemctl %s %s failed: %s: %w", action, service, strings.TrimSpace(string(out)), err)
	}
	return nil
}
