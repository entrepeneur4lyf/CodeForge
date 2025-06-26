package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/entrepeneur4lyf/codeforge/internal/lsp"
	"github.com/spf13/cobra"
)

var completeCmd = &cobra.Command{
	Use:   "complete [file] [line] [char]",
	Short: "Get code completions for a given file and position",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		file := args[0]
		line := 0
		char := 0
		fmt.Sscanf(args[1], "%d", &line)
		fmt.Sscanf(args[2], "%d", &char)

		ctx := context.Background()

		// Get the appropriate LSP client based on file extension
		var client *lsp.Client
		var exists bool

		// Try to get gopls for Go files
		if file[len(file)-3:] == ".go" {
			client, exists = lsp.GetClient("gopls")
		} else if file[len(file)-3:] == ".rs" {
			client, exists = lsp.GetClient("rust-analyzer")
		} else if file[len(file)-3:] == ".py" {
			client, exists = lsp.GetClient("pylsp")
		} else if file[len(file)-3:] == ".ts" || file[len(file)-3:] == ".js" {
			client, exists = lsp.GetClient("typescript-language-server")
		}

		if !exists || client == nil {
			fmt.Printf("No LSP client available for file: %s\n", file)
			return
		}

		// Open the file in the LSP server
		if err := client.OpenFile(ctx, file); err != nil {
			log.Printf("Failed to open file in LSP: %v", err)
			return
		}
		defer client.CloseFile(ctx, file)

		// Get completions
		completions, err := client.GetCompletion(ctx, file, line, char)
		if err != nil {
			log.Printf("Failed to get completions: %v", err)
			return
		}

		if completions == nil {
			fmt.Println("No completions available")
			return
		}

		for _, item := range completions.Items {
			fmt.Printf("%s - %s\n", item.Label, item.Detail)
		}
	},
}

func init() {
	rootCmd.AddCommand(completeCmd)
}
