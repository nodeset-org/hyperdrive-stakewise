/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// restartCmd represents the restart command
var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Stops and restarts node containers.",
	Long: `Runs a "docker compose down" and "docker compose up -d".
`,
	Run: func(cmd *cobra.Command, args []string) {
		stopCmd.Run(cmd, args)
		startCmd.Run(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(restartCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// restartCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// restartCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
