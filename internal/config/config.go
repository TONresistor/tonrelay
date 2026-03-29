package config

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"

	tunnelconfig "github.com/ton-blockchain/adnl-tunnel/config"
)

const (
	DefaultListenAddr    = "0.0.0.0:17330"
	DefaultNetworkConfig = "https://ton-blockchain.github.io/global.config.json"
	DefaultDBPath        = "/var/lib/tonrelay/payments-db/"
)

func Load(path string) (*tunnelconfig.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg tunnelconfig.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

func Save(cfg *tunnelconfig.Config, path string) error {
	data, err := json.MarshalIndent(cfg, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func GenerateDefault(externalIP string, port uint16, dataDir string) (*tunnelconfig.Config, error) {
	tunnelKey, err := generateSeed()
	if err != nil {
		return nil, err
	}
	walletKey, err := generateSeed()
	if err != nil {
		return nil, err
	}
	adnlKey, err := generateSeed()
	if err != nil {
		return nil, err
	}
	paymentsKey, err := generateSeed()
	if err != nil {
		return nil, err
	}

	dbPath := DefaultDBPath
	if dataDir != "" {
		dbPath = dataDir + "/payments-db/"
	}

	listenAddr := fmt.Sprintf("0.0.0.0:%d", port)

	cfg := &tunnelconfig.Config{
		TunnelServerKey:  tunnelKey,
		TunnelListenAddr: listenAddr,
		TunnelThreads:    uint(runtime.NumCPU()),
		NetworkConfigUrl: DefaultNetworkConfig,
		ExternalIP:       externalIP,
		PaymentsEnabled:  false,
		Payments: tunnelconfig.PaymentsConfig{
			ADNLServerKey:    adnlKey,
			PaymentsNodeKey:  paymentsKey,
			WalletPrivateKey: walletKey,
			DBPath:           dbPath,
		},
	}

	return cfg, nil
}

func Validate(cfg *tunnelconfig.Config) error {
	if len(cfg.TunnelServerKey) != 32 {
		return fmt.Errorf("TunnelServerKey must be 32 bytes")
	}
	if cfg.TunnelListenAddr == "" {
		return fmt.Errorf("TunnelListenAddr is required")
	}
	if cfg.ExternalIP != "" {
		if net.ParseIP(cfg.ExternalIP) == nil {
			return fmt.Errorf("invalid ExternalIP: %q", cfg.ExternalIP)
		}
	}
	if cfg.PaymentsEnabled {
		if len(cfg.Payments.WalletPrivateKey) != 32 {
			return fmt.Errorf("WalletPrivateKey must be 32 bytes when payments enabled")
		}
		if len(cfg.Payments.PaymentsNodeKey) != 32 {
			return fmt.Errorf("PaymentsNodeKey must be 32 bytes when payments enabled")
		}
		if len(cfg.Payments.ADNLServerKey) != 32 {
			return fmt.Errorf("ADNLServerKey must be 32 bytes when payments enabled")
		}
	}
	return nil
}

func MaskedDisplay(cfg *tunnelconfig.Config) string {
	var b strings.Builder

	b.WriteString("Configuration:\n")
	b.WriteString(fmt.Sprintf("  Listen Address:  %s\n", cfg.TunnelListenAddr))
	b.WriteString(fmt.Sprintf("  External IP:     %s\n", cfg.ExternalIP))
	b.WriteString(fmt.Sprintf("  Tunnel Threads:  %d\n", cfg.TunnelThreads))
	b.WriteString(fmt.Sprintf("  Network Config:  %s\n", cfg.NetworkConfigUrl))

	adnlID := GetADNLID(cfg)
	b.WriteString(fmt.Sprintf("  ADNL ID:         %s\n", adnlID))

	b.WriteString(fmt.Sprintf("  Payments:        %v\n", cfg.PaymentsEnabled))
	if cfg.PaymentsEnabled {
		b.WriteString(fmt.Sprintf("  Route Price:     %d nano\n", cfg.Payments.MinPricePerPacketRoute))
		b.WriteString(fmt.Sprintf("  Out Price:       %d nano\n", cfg.Payments.MinPricePerPacketInOut))
		b.WriteString(fmt.Sprintf("  DB Path:         %s\n", cfg.Payments.DBPath))
		b.WriteString("  Wallet Key:      ********\n")
	}

	return b.String()
}

func GetADNLID(cfg *tunnelconfig.Config) string {
	if len(cfg.TunnelServerKey) != 32 {
		return "unknown"
	}
	priv := ed25519.NewKeyFromSeed(cfg.TunnelServerKey)
	pub := priv.Public().(ed25519.PublicKey)
	return base64.StdEncoding.EncodeToString(pub)
}

func generateSeed() ([]byte, error) {
	seed := make([]byte, 32)
	if _, err := rand.Read(seed); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	return seed, nil
}
