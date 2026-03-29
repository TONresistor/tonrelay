package ip

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

func DetectExternalIP() (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get("https://api.ipify.org")
	if err != nil {
		return "", fmt.Errorf("failed to detect external IP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("IP detection returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read IP response: %w", err)
	}

	ip := strings.TrimSpace(string(body))
	if net.ParseIP(ip) == nil {
		return "", fmt.Errorf("invalid IP returned: %q", ip)
	}

	return ip, nil
}
