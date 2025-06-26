package cmd

import (
	"github.com/entrepeneur4lyf/codeforge/internal/tui"
	"github.com/spf13/cobra"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start an interactive chat session",
	Run: func(cmd *cobra.Command, args []string) {
		// Configuration and LLM initialization is handled by the root command
		tui.Start()
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)
}
