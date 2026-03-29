package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/TONresistor/tonrelay/internal/updater"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update tunnel-node to latest version",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireSudo(); err != nil {
			return err
		}

		version, _ := cmd.Flags().GetString("version")

		if version == "" {
			info, err := updater.CheckUpdate(dataDir)
			if err != nil {
				return err
			}

			fmt.Printf("Current version: %s\n", info.CurrentVersion)
			fmt.Printf("Latest version:  %s\n", info.LatestVersion)

			if !info.UpdateAvailable {
				fmt.Println("Already on the latest version.")
				return nil
			}

			version = info.LatestVersion
		}

		fmt.Printf("\nUpdate to %s? [y/N] ", version)
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))

		if answer != "y" && answer != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}

		return updater.Update(version, BinaryPath, dataDir)
	},
}

func init() {
	updateCmd.Flags().String("version", "", "target version (default: latest)")
	rootCmd.AddCommand(updateCmd)
}
