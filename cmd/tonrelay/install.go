package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/TONresistor/tonrelay/internal/config"
	"github.com/TONresistor/tonrelay/internal/installer"
	"github.com/TONresistor/tonrelay/internal/service"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Download tunnel-node and set up a relay",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireSudo(); err != nil {
			return err
		}

		ip, _ := cmd.Flags().GetString("ip")
		port, _ := cmd.Flags().GetUint16("port")
		version, _ := cmd.Flags().GetString("version")
		clearnetExit, _ := cmd.Flags().GetBool("clearnet-exit")

		opts := installer.Options{
			ExternalIP:   ip,
			Port:         port,
			Version:      version,
			ConfigPath:   cfgPath,
			DataDir:      dataDir,
			BinaryPath:   BinaryPath,
			User:         SystemUser,
			ClearnetExit: clearnetExit,
		}

		if err := installer.Install(opts); err != nil {
			return err
		}

		// Auto-start
		fmt.Println("\nStarting relay...")
		if err := service.Start(); err != nil {
			return fmt.Errorf("failed to start: %w\nYou can start manually with: sudo tonrelay start", err)
		}

		if err := service.HealthCheck(30 * time.Second); err != nil {
			return fmt.Errorf("relay started but health check failed: %w\nCheck logs with: tonrelay logs", err)
		}

		printSummary()

		return nil
	},
}

func printSummary() {
	fmt.Println()
	fmt.Println("────────────────────────────────────")
	fmt.Println("  Relay is online")
	fmt.Println("────────────────────────────────────")

	if c, err := config.Load(cfgPath); err == nil {
		fmt.Printf("  IP:      %s\n", c.ExternalIP)
		fmt.Printf("  Port:    %s\n", extractPort(c.TunnelListenAddr))
		fmt.Printf("  ADNL ID: %s\n", truncateID(config.GetADNLID(c)))
		if c.AllowClearnetExit {
			fmt.Println("  Mode:    free + clearnet exit")
		} else {
			fmt.Println("  Mode:    free")
		}
	}
	fmt.Println("────────────────────────────────────")
	fmt.Println()
	fmt.Println("Your relay will appear in the network within ~5 minutes.")
	fmt.Println()
	fmt.Println("Useful commands:")
	fmt.Println("  tonrelay status        quick status check")
	fmt.Println("  tonrelay status --live live dashboard")
	fmt.Println("  tonrelay logs -f       follow logs")
}

func extractPort(addr string) string {
	parts := strings.SplitN(addr, ":", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return addr
}

func truncateID(id string) string {
	if len(id) > 20 {
		return id[:20] + "..."
	}
	return id
}

func init() {
	installCmd.Flags().String("ip", "", "external IP (auto-detected if not set)")
	installCmd.Flags().Uint16("port", 17330, "UDP listen port")
	installCmd.Flags().String("version", "", "tunnel-node version (default: latest)")
	installCmd.Flags().Bool("clearnet-exit", false, "Enable clearnet TCP exit mode (dual: relay + exit)")
	rootCmd.AddCommand(installCmd)
}
