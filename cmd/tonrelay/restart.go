package main

import (
	"fmt"
	"time"

	"github.com/TONresistor/tonrelay/internal/service"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the relay service",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireSudo(); err != nil {
			return err
		}

		if !service.IsInstalled() {
			return &ExitError{Code: ExitNotInstalled, Msg: "relay is not installed. Run: sudo tonrelay install"}
		}

		fmt.Println("Restarting tonrelay service...")
		if err := service.Restart(); err != nil {
			return err
		}

		fmt.Println("Waiting for service to start...")
		if err := service.HealthCheck(30 * time.Second); err != nil {
			return fmt.Errorf("health check failed: %w", err)
		}

		fmt.Println("Service restarted successfully.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(restartCmd)
}
