package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show tonrelay and tunnel-node versions",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("tonrelay %s\n", Version)

		versionFile := filepath.Join(dataDir, "version")
		data, err := os.ReadFile(versionFile)
		if err != nil {
			fmt.Println("tunnel-node: not installed")
			return nil
		}
		fmt.Printf("tunnel-node %s\n", strings.TrimSpace(string(data)))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
