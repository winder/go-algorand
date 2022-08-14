package main

import (
	"github.com/spf13/cobra"
)

var goranRoot = &cobra.Command{
	Use:   "goran",
	Short: "Build automation tools for Algorand development.",
	Long:  `Goran is the CLI for development on the Algorand blockchain. The binary 'goran' provides common utilities and integration points to serve as the core interface for development on algorand.`,
	Run: func(cmd *cobra.Command, args []string) {
		//If no arguments passed, we should fallback to help
		cmd.HelpFunc()(cmd, args)
	},
}
