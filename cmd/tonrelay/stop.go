package main

import (
	"fmt"

	"github.com/TONresistor/tonrelay/internal/service"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the relay",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireSudo(); err != nil {
			return err
		}

		if !service.IsActive() {
			fmt.Println("Relay is not running.")
			return nil
		}

		fmt.Println("Stopping relay...")
		if err := service.Stop(); err != nil {
			return err
		}

		fmt.Println("Relay stopped.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
