package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "eagle-utils",
	Short: "Utilities for synchronizing data from Eagle",

	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
