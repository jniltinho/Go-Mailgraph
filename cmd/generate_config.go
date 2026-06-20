package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

//go:embed files/config.default.toml
var defaultConfig []byte

var generateConfigCmd = &cobra.Command{
	Use:   "generate-config",
	Short: "Generate a default config.toml file",
	Run: func(cmd *cobra.Command, args []string) {
		name := fmt.Sprintf("config_%s.toml", time.Now().Format("2006-01-02_150405"))
		if err := os.WriteFile(name, defaultConfig, 0o644); err != nil {
			cmd.PrintErrf("error writing %s: %v\n", name, err)
			os.Exit(1)
		}
		fmt.Printf("Successfully generated %s\n", name)
	},
}

func init() {
	rootCmd.AddCommand(generateConfigCmd)
}