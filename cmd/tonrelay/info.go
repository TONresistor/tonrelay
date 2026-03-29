package main

import (
	"encoding/json"
	"fmt"

	"github.com/TONresistor/tonrelay/internal/config"
	"github.com/TONresistor/tonrelay/internal/service"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show relay info",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return fmt.Errorf("no config found at %s. Run: sudo tonrelay install", cfgPath)
		}

		adnlID := config.GetADNLID(cfg)
		active := service.IsActive()

		if jsonOut {
			data, err := json.MarshalIndent(map[string]interface{}{
				"adnl_id":          adnlID,
				"external_ip":      cfg.ExternalIP,
				"listen_addr":      cfg.TunnelListenAddr,
				"threads":          cfg.TunnelThreads,
				"payments_enabled": cfg.PaymentsEnabled,
				"service_active":   active,
			}, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal output: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("adnl:     %s\n", adnlID)
		fmt.Printf("endpoint: %s:%s\n", cfg.ExternalIP, extractPort(cfg.TunnelListenAddr))
		fmt.Printf("threads:  %d\n", cfg.TunnelThreads)

		if cfg.PaymentsEnabled {
			fmt.Printf("mode:     paid (%d nano/pkt)\n", cfg.Payments.MinPricePerPacketRoute)
		} else {
			fmt.Println("mode:     free")
		}

		if active {
			fmt.Println("service:  running")
		} else {
			fmt.Println("service:  stopped")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
