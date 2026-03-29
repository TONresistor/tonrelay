package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tunnelconfig "github.com/ton-blockchain/adnl-tunnel/config"
)

func TestGenerateDefault(t *testing.T) {
	cfg, err := GenerateDefault("1.2.3.4", 17330, "/tmp/test-tonrelay")
	if err != nil {
		t.Fatalf("GenerateDefault: %v", err)
	}

	if cfg.ExternalIP != "1.2.3.4" {
		t.Errorf("ExternalIP = %q, want %q", cfg.ExternalIP, "1.2.3.4")
	}
	if cfg.TunnelListenAddr != "0.0.0.0:17330" {
		t.Errorf("ListenAddr = %q, want %q", cfg.TunnelListenAddr, "0.0.0.0:17330")
	}
	if len(cfg.TunnelServerKey) != 32 {
		t.Errorf("TunnelServerKey length = %d, want 32", len(cfg.TunnelServerKey))
	}
	if len(cfg.Payments.WalletPrivateKey) != 32 {
		t.Errorf("WalletPrivateKey length = %d, want 32", len(cfg.Payments.WalletPrivateKey))
	}
	if cfg.TunnelThreads == 0 {
		t.Error("TunnelThreads should be > 0")
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg, err := GenerateDefault("10.0.0.1", 17330, dir)
	if err != nil {
		t.Fatalf("GenerateDefault: %v", err)
	}

	if err := Save(cfg, path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.ExternalIP != cfg.ExternalIP {
		t.Errorf("loaded ExternalIP = %q, want %q", loaded.ExternalIP, cfg.ExternalIP)
	}
	if loaded.TunnelListenAddr != cfg.TunnelListenAddr {
		t.Errorf("loaded ListenAddr = %q, want %q", loaded.TunnelListenAddr, cfg.TunnelListenAddr)
	}

	// Verify file permissions
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("file perm = %o, want 0600", info.Mode().Perm())
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*tunnelconfig.Config)
		wantErr bool
	}{
		{
			name:    "valid config",
			modify:  func(c *tunnelconfig.Config) {},
			wantErr: false,
		},
		{
			name:    "empty tunnel key",
			modify:  func(c *tunnelconfig.Config) { c.TunnelServerKey = nil },
			wantErr: true,
		},
		{
			name:    "empty listen addr",
			modify:  func(c *tunnelconfig.Config) { c.TunnelListenAddr = "" },
			wantErr: true,
		},
		{
			name:    "invalid external IP",
			modify:  func(c *tunnelconfig.Config) { c.ExternalIP = "not-an-ip" },
			wantErr: true,
		},
		{
			name: "payments enabled missing wallet key",
			modify: func(c *tunnelconfig.Config) {
				c.PaymentsEnabled = true
				c.Payments.WalletPrivateKey = nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _ := GenerateDefault("1.2.3.4", 17330, "/tmp/test")
			tt.modify(cfg)
			err := Validate(cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMaskedDisplay(t *testing.T) {
	cfg, _ := GenerateDefault("1.2.3.4", 17330, "/tmp/test")
	cfg.PaymentsEnabled = true

	display := MaskedDisplay(cfg)

	if !strings.Contains(display, "1.2.3.4") {
		t.Error("display should contain IP")
	}
	if !strings.Contains(display, "********") {
		t.Error("display should mask wallet key")
	}
	if strings.Contains(display, string(cfg.Payments.WalletPrivateKey)) {
		t.Error("display should NOT contain actual wallet key")
	}
}

func TestGetADNLID(t *testing.T) {
	cfg, _ := GenerateDefault("1.2.3.4", 17330, "/tmp/test")
	id := GetADNLID(cfg)

	if id == "" || id == "unknown" {
		t.Error("ADNL ID should not be empty")
	}
	if len(id) < 40 {
		t.Errorf("ADNL ID too short: %q", id)
	}
}

func TestConfigJSONRoundtrip(t *testing.T) {
	cfg, _ := GenerateDefault("5.6.7.8", 17330, "/tmp/test")

	data, err := json.MarshalIndent(cfg, "", "\t")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var loaded tunnelconfig.Config
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if loaded.ExternalIP != cfg.ExternalIP {
		t.Errorf("ExternalIP = %q, want %q", loaded.ExternalIP, cfg.ExternalIP)
	}
}
