package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// version is set via ldflags at build time.
var version = "dev"

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of cred",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "cred version %s\n", version)
	},
}
