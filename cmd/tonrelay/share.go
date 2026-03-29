package main

import (
	"fmt"

	"github.com/TONresistor/tonrelay/internal/config"
	"github.com/TONresistor/tonrelay/internal/share"
	"github.com/spf13/cobra"
)

var shareCmd = &cobra.Command{
	Use:   "share",
	Short: "Generate shareable config for clients",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		outFile, _ := cmd.Flags().GetString("output")

		if outFile != "" {
			if err := share.GenerateToFile(cfg, outFile); err != nil {
				return err
			}
			fmt.Printf("Shared config written to %s\n", outFile)
			return nil
		}

		data, err := share.Generate(cfg)
		if err != nil {
			return err
		}

		fmt.Println(string(data))
		return nil
	},
}

func init() {
	shareCmd.Flags().StringP("output", "o", "", "write to file instead of stdout")
	rootCmd.AddCommand(shareCmd)
}
