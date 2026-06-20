package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/davidullrich/mailgraph/internal/config"
)

var catCmd = &cobra.Command{
	Use:   "cat",
	Short: "Read logfile once and update RRD",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			cmd.PrintErrln(err)
			os.Exit(1)
		}
		cfg.Cat = true
		cfg.Serve = false

		if err := runCat(cfg); err != nil {
			cmd.PrintErrln(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(catCmd)
	bindConfigFlags(catCmd)
}