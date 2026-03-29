package main

import (
	"os"
	"os/exec"
	"strconv"

	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show relay logs from journalctl",
	RunE: func(cmd *cobra.Command, args []string) error {
		lines, _ := cmd.Flags().GetInt("lines")
		follow, _ := cmd.Flags().GetBool("follow")

		jArgs := []string{"-u", "tonrelay", "--no-pager"}

		if follow {
			jArgs = append(jArgs, "-f")
		} else {
			jArgs = append(jArgs, "-n", strconv.Itoa(lines))
		}

		c := exec.Command("journalctl", jArgs...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

func init() {
	logsCmd.Flags().IntP("lines", "n", 50, "number of lines to show")
	logsCmd.Flags().BoolP("follow", "f", false, "follow log output")
	rootCmd.AddCommand(logsCmd)
}
