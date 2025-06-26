package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start an interactive chat session",
	Run: func(cmd *cobra.Command, args []string) {
		// Configuration and LLM initialization is handled by the root command
		// For now, redirect to TUI command
		fmt.Println("Use 'codeforge tui' to start the interactive interface")
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)
}
