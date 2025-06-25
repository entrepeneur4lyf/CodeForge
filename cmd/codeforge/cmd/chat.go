package cmd

import (
	"fmt"
	"os"

	"github.com/shawn/codeforge/internal/tui"
	"github.com/spf13/cobra"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start an interactive chat session",
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			fmt.Println("Error: OPENAI_API_KEY environment variable not set.")
			os.Exit(1)
		}
		tui.Start(apiKey)
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)
}