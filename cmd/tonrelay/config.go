package main

import (
	"fmt"
	"net"

	"github.com/TONresistor/tonrelay/internal/config"
	"github.com/TONresistor/tonrelay/internal/service"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current config (wallet key masked)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		fmt.Print(config.MaskedDisplay(cfg))
		return nil
	},
}

var configSetIPCmd = &cobra.Command{
	Use:   "set-ip <ip>",
	Short: "Update external IP address",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireSudo(); err != nil {
			return err
		}

		newIP := args[0]
		if net.ParseIP(newIP) == nil {
			return fmt.Errorf("invalid IP address: %q", newIP)
		}

		cfg, err := config.Load(cfgPath)
		if err != nil {
			return err
		}

		cfg.ExternalIP = newIP
		if err := config.Save(cfg, cfgPath); err != nil {
			return err
		}

		fmt.Printf("External IP updated to %s\n", newIP)

		if service.IsActive() {
			fmt.Println("Restarting service to apply changes...")
			if err := service.Restart(); err != nil {
				return fmt.Errorf("restart failed: %w", err)
			}
			fmt.Println("Service restarted.")
		}

		return nil
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetIPCmd)
	rootCmd.AddCommand(configCmd)
}
