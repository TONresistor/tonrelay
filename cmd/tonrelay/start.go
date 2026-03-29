package main

import (
	"fmt"
	"time"

	"github.com/TONresistor/tonrelay/internal/service"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the relay",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireSudo(); err != nil {
			return err
		}

		if !service.IsInstalled() {
			return &ExitError{Code: ExitNotInstalled, Msg: "relay is not installed. Run: sudo tonrelay install"}
		}

		if service.IsActive() {
			fmt.Println("Relay is already running.")
			return nil
		}

		fmt.Println("Starting relay...")
		if err := service.Start(); err != nil {
			return err
		}

		if err := service.HealthCheck(30 * time.Second); err != nil {
			return fmt.Errorf("relay started but not healthy. Check: tonrelay logs")
		}

		fmt.Println("Relay is running.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
