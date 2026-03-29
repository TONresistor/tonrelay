package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	Version = "dev"

	cfgPath string
	dataDir string
	noColor bool
	jsonOut bool
)

const (
	DefaultConfigPath = "/etc/tonrelay/config.json"
	DefaultDataDir    = "/var/lib/tonrelay/"
	ServiceName       = "tonrelay"
	BinaryPath        = "/usr/local/bin/tunnel-node"
	SystemUser        = "tonrelay"
)

const (
	ExitNotInstalled = 3
	ExitPermission   = 6
)

type ExitError struct {
	Code int
	Msg  string
}

func (e *ExitError) Error() string { return e.Msg }

var rootCmd = &cobra.Command{
	Use:   "tonrelay",
	Short: "CLI tool for TON tunnel relay operators",
	Long:  "tonrelay wraps tunnel-node (ton-blockchain/adnl-tunnel) to provide relay operators with one-command installation, service management, live monitoring, and configuration.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", DefaultConfigPath, "config file path")
	rootCmd.PersistentFlags().StringVar(&dataDir, "data-dir", DefaultDataDir, "data directory path")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "output in JSON format")

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if noColor {
			os.Setenv("NO_COLOR", "1")
		}
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		if exitErr, ok := err.(*ExitError); ok {
			os.Exit(exitErr.Code)
		}
		os.Exit(1)
	}
}

func requireSudo() error {
	if os.Geteuid() != 0 {
		return &ExitError{
			Code: ExitPermission,
			Msg:  fmt.Sprintf("this command needs root privileges. Run: sudo tonrelay %s", os.Args[1]),
		}
	}
	return nil
}
