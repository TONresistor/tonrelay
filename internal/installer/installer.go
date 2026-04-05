package installer

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/TONresistor/tonrelay/internal/config"
	"github.com/TONresistor/tonrelay/internal/github"
	"github.com/TONresistor/tonrelay/internal/ip"
)

type Options struct {
	ExternalIP   string
	Port         uint16
	Version      string
	ConfigPath   string
	DataDir      string
	BinaryPath   string
	User         string
	ClearnetExit bool
}

const systemdUnitTemplate = `[Unit]
Description=TON Tunnel Relay Node
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User={{.User}}
Group={{.User}}
ExecStart=/bin/sh -c 'sleep infinity | exec {{.BinaryPath}} -config {{.ConfigPath}} -v 3 -log-disable-file -metrics-listen-addr 127.0.0.1:9091{{if .ClearnetExit}} -clearnet-exit{{end}}'
WorkingDirectory={{.DataDir}}
Restart=always
RestartSec=5
LimitNOFILE=65535
LimitCORE=0
StandardInput=null
StandardOutput=journal
StandardError=journal
SyslogIdentifier=tonrelay

# Security hardening
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths={{.DataDir}}
PrivateTmp=yes

[Install]
WantedBy=multi-user.target
`

func validatePath(p string) error {
	for _, c := range p {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '/' || c == '-' || c == '_' || c == '.') {
			return fmt.Errorf("invalid character %q in path %q", string(c), p)
		}
	}
	if strings.Contains(p, "..") {
		return fmt.Errorf("path %q contains directory traversal", p)
	}
	return nil
}

func Install(opts Options) error {
	fmt.Println("Installing tonrelay...")

	if err := validatePath(opts.ConfigPath); err != nil {
		return fmt.Errorf("config path: %w", err)
	}
	if err := validatePath(opts.DataDir); err != nil {
		return fmt.Errorf("data dir: %w", err)
	}

	if err := createUser(opts.User); err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	if err := createDirs(opts.ConfigPath, opts.DataDir, opts.User); err != nil {
		return fmt.Errorf("create directories: %w", err)
	}

	version, err := downloadBinary(opts.Version, opts.BinaryPath, opts.DataDir)
	if err != nil {
		return fmt.Errorf("download binary: %w", err)
	}

	externalIP := opts.ExternalIP
	if externalIP == "" {
		fmt.Println("Detecting external IP...")
		detected, err := ip.DetectExternalIP()
		if err != nil {
			return fmt.Errorf("IP detection failed (use --ip to set manually): %w", err)
		}
		externalIP = detected
		fmt.Printf("Detected IP: %s\n", externalIP)
	}

	checkPort(opts.Port)

	if err := generateConfig(opts.ConfigPath, externalIP, opts.Port, opts.DataDir, opts.ClearnetExit); err != nil {
		return fmt.Errorf("generate config: %w", err)
	}

	// Fix ownership on config and data files written as root
	if err := chownRecursive(filepath.Dir(opts.ConfigPath), opts.User); err != nil {
		return fmt.Errorf("chown config dir: %w", err)
	}
	if err := chownRecursive(opts.DataDir, opts.User); err != nil {
		return fmt.Errorf("chown data dir: %w", err)
	}

	if err := writeSystemdUnit(opts); err != nil {
		return fmt.Errorf("write systemd unit: %w", err)
	}

	if err := storeVersion(opts.DataDir, version); err != nil {
		return fmt.Errorf("store version: %w", err)
	}

	// Final ownership fix for version file written after chown
	if err := chownRecursive(opts.DataDir, opts.User); err != nil {
		return fmt.Errorf("chown data dir: %w", err)
	}

	return nil
}

func createUser(user string) error {
	if _, err := exec.LookPath("id"); err == nil {
		if err := exec.Command("id", user).Run(); err == nil {
			fmt.Printf("User %s already exists\n", user)
			return nil
		}
	}

	fmt.Printf("Creating system user %s...\n", user)
	cmd := exec.Command("useradd", "--system", "--no-create-home", "--shell", "/usr/sbin/nologin", user)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("useradd failed: %s: %w", string(out), err)
	}
	return nil
}

func createDirs(configPath, dataDir, user string) error {
	configDir := filepath.Dir(configPath)

	dirs := []struct {
		path string
		perm os.FileMode
	}{
		{configDir, 0700},
		{dataDir, 0700},
	}

	for _, d := range dirs {
		fmt.Printf("Creating %s...\n", d.path)
		if err := os.MkdirAll(d.path, d.perm); err != nil {
			return fmt.Errorf("mkdir %s: %w", d.path, err)
		}
		if err := chownRecursive(d.path, user); err != nil {
			return fmt.Errorf("chown %s: %w", d.path, err)
		}
	}
	return nil
}

func chownRecursive(path, user string) error {
	cmd := exec.Command("chown", "-R", user+":"+user, path)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %w", string(out), err)
	}
	return nil
}

func downloadBinary(version, binaryPath, dataDir string) (string, error) {
	gh := github.NewClient()

	var release *github.Release
	var err error
	if version == "" || version == "latest" {
		fmt.Println("Fetching latest tunnel-node release...")
		release, err = gh.GetLatestRelease()
	} else {
		fmt.Printf("Fetching tunnel-node %s...\n", version)
		release, err = gh.GetReleaseByTag(version)
	}
	if err != nil {
		return "", err
	}

	assetName := github.AssetName()
	asset, err := release.FindAsset(assetName)
	if err != nil {
		return "", err
	}

	checksums, err := gh.DownloadChecksums(release)
	if err != nil {
		fmt.Printf("Warning: could not download checksums.txt: %v\n", err)
	}

	fmt.Printf("Downloading %s (%s)...\n", assetName, release.TagName)
	checksum, err := gh.DownloadAsset(asset, binaryPath)
	if err != nil {
		return "", err
	}

	if expected, ok := checksums[assetName]; ok {
		if checksum != expected {
			os.Remove(binaryPath)
			return "", fmt.Errorf("checksum mismatch: expected %s, got %s", expected, checksum)
		}
		fmt.Println("Checksum verified against upstream checksums.txt")
	}

	checksumPath := filepath.Join(dataDir, "tunnel-node.sha256")
	if err := os.WriteFile(checksumPath, []byte(checksum), 0600); err != nil {
		return "", fmt.Errorf("failed to store checksum: %w", err)
	}

	fmt.Printf("Binary installed at %s (SHA256: %s)\n", binaryPath, checksum[:16]+"...")
	return release.TagName, nil
}

func generateConfig(configPath, externalIP string, port uint16, dataDir string, clearnetExit bool) error {
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Config already exists at %s, skipping generation\n", configPath)
		return nil
	}

	fmt.Println("Generating config...")
	cfg, err := config.GenerateDefault(externalIP, port, dataDir)
	if err != nil {
		return err
	}

	if clearnetExit {
		cfg.AllowClearnetExit = true
		cfg.ClearnetExitPorts = []int{443}
		cfg.MaxTCPConnsPerSection = 64
	}

	if err := config.Validate(cfg); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	if err := config.Save(cfg, configPath); err != nil {
		return err
	}

	fmt.Printf("Config written to %s\n", configPath)
	return nil
}

func writeSystemdUnit(opts Options) error {
	unitPath := "/etc/systemd/system/tonrelay.service"
	fmt.Printf("Writing systemd unit to %s...\n", unitPath)

	tmpl, err := template.New("unit").Parse(systemdUnitTemplate)
	if err != nil {
		return err
	}

	f, err := os.Create(unitPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := tmpl.Execute(f, opts); err != nil {
		return err
	}

	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("daemon-reload failed: %w", err)
	}

	if err := exec.Command("systemctl", "enable", "tonrelay").Run(); err != nil {
		return fmt.Errorf("enable service failed: %w", err)
	}

	fmt.Println("Service enabled.")
	return nil
}

func storeVersion(dataDir, version string) error {
	versionPath := filepath.Join(dataDir, "version")
	return os.WriteFile(versionPath, []byte(version+"\n"), 0600)
}

func checkPort(port uint16) {
	portStr := strconv.Itoa(int(port))

	// Check if port is already in use
	conn, err := net.ListenPacket("udp", ":"+portStr)
	if err != nil {
		fmt.Printf("\n  Port %s/udp is already in use.\n", portStr)
		fmt.Printf("  Check what's using it: ss -ulnp | grep %s\n\n", portStr)
		return
	}
	conn.Close()
	fmt.Printf("Port %s/udp is available.\n", portStr)
	fmt.Printf("  Make sure your firewall allows incoming UDP on port %s.\n", portStr)
}
