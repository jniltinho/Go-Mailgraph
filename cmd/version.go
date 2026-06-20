package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"mailgraph/internal/buildinfo"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("mailgraph %s (Go port)\n", buildinfo.Version)
		fmt.Printf("  Build: %s\n", buildinfo.BuildDate)
		fmt.Printf("  Commit: %s\n", buildinfo.GitCommit)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}