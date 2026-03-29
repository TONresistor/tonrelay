package share

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"os"

	tunnelconfig "github.com/ton-blockchain/adnl-tunnel/config"
)

func Generate(cfg *tunnelconfig.Config) ([]byte, error) {
	if len(cfg.TunnelServerKey) != ed25519.SeedSize {
		return nil, fmt.Errorf("invalid TunnelServerKey length: %d (expected %d)", len(cfg.TunnelServerKey), ed25519.SeedSize)
	}
	shared := &tunnelconfig.SharedConfig{
		NodesPool: []tunnelconfig.TunnelRouteSection{
			{
				Key: ed25519.NewKeyFromSeed(cfg.TunnelServerKey).Public().(ed25519.PublicKey),
			},
		},
	}

	return json.MarshalIndent(shared, "", "\t")
}

func GenerateToFile(cfg *tunnelconfig.Config, outPath string) error {
	data, err := Generate(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(outPath, data, 0644)
}
