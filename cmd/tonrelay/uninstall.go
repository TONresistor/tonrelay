package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/TONresistor/tonrelay/internal/service"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove tunnel-node, service, and all artifacts",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireSudo(); err != nil {
			return err
		}

		keepData, _ := cmd.Flags().GetBool("keep-data")
		keepConfig, _ := cmd.Flags().GetBool("keep-config")

		fmt.Println("This will remove:")
		fmt.Printf("  - Systemd service (%s)\n", ServiceName)
		fmt.Printf("  - Binary (%s)\n", BinaryPath)
		if !keepConfig {
			fmt.Printf("  - Config (%s)\n", cfgPath)
		}
		if !keepData {
			fmt.Printf("  - Data directory (%s)\n", dataDir)
			fmt.Println("\n  WARNING: Data directory contains runtime data.")
		}

		fmt.Print("\nProceed? [y/N] ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))

		if answer != "y" && answer != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}

		// Stop and disable service
		if service.IsActive() {
			fmt.Println("Stopping service...")
			if err := service.Stop(); err != nil {
				fmt.Printf("Warning: failed to stop service: %v\n", err)
			}
		}
		if service.IsEnabled() {
			if err := service.Disable(); err != nil {
				fmt.Printf("Warning: failed to disable service: %v\n", err)
			}
		}

		// Remove systemd unit
		unitPath := "/etc/systemd/system/tonrelay.service"
		if err := os.Remove(unitPath); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Warning: failed to remove %s: %v\n", unitPath, err)
		}
		if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
			fmt.Printf("Warning: failed to reload systemd: %v\n", err)
		}

		// Remove binary
		if err := os.Remove(BinaryPath); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Warning: failed to remove %s: %v\n", BinaryPath, err)
		}

		// Remove config
		if !keepConfig {
			configDir := filepath.Dir(cfgPath)
			if configDir == "/etc/tonrelay" {
				if err := os.RemoveAll(configDir); err != nil {
					fmt.Printf("Warning: failed to remove config dir: %v\n", err)
				}
			} else {
				// Custom config path — only remove the file, not the directory
				if err := os.Remove(cfgPath); err != nil && !os.IsNotExist(err) {
					fmt.Printf("Warning: failed to remove config file: %v\n", err)
				}
			}
		}

		// Remove data
		if !keepData {
			if err := os.RemoveAll(dataDir); err != nil {
				fmt.Printf("Warning: failed to remove data dir: %v\n", err)
			}
		}

		// Remove user
		fmt.Printf("Remove system user %s? [y/N] ", SystemUser)
		answer, _ = reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer == "y" || answer == "yes" {
			if err := exec.Command("userdel", SystemUser).Run(); err != nil {
				fmt.Printf("Warning: failed to remove user %s: %v\n", SystemUser, err)
			} else {
				fmt.Printf("User %s removed.\n", SystemUser)
			}
		}

		fmt.Println("Uninstall complete.")
		return nil
	},
}

func init() {
	uninstallCmd.Flags().Bool("keep-data", false, "preserve data directory")
	uninstallCmd.Flags().Bool("keep-config", false, "preserve config file")
	rootCmd.AddCommand(uninstallCmd)
}
