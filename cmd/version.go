/*
Copyright © 2025 Jonathan Bowe jonathan@bowedev.com
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var VERSION string = "v0.1.0"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number for vincent-vimgo",
	Long: `To return the version or not to return a version,
that is the question.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s\n", VERSION)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
