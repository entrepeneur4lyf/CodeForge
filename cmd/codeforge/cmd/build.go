package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/shawn/codeforge/internal/builder"
	"github.com/shawn/codeforge/internal/llm"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the project and attempt to fix errors",
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			fmt.Println("Error: OPENAI_API_KEY environment variable not set.")
			os.Exit(1)
		}
		llm.Init(apiKey)

		for i := 0; i < 3; i++ { // Limit to 3 attempts
			output, err := builder.BuildGo()
			if err == nil {
				fmt.Println("Build successful!")
				return
			}

			errorStr := string(output)
			fmt.Printf("Build failed:\n%s\n", errorStr)

			filePath, _ := builder.ParseError(errorStr)
			if filePath == "" {
				fmt.Println("Could not parse file path from error.")
				return
			}

			fileContent, err := ioutil.ReadFile(filePath)
			if err != nil {
				fmt.Printf("Error reading file %s: %s\n", filePath, err)
				return
			}

			prompt := fmt.Sprintf(
				"The following Go code in file '%s' failed to build. Please provide only the corrected code, without any explanation.\n\nFile Content:\n```go\n%s\n```\n\nError:\n%s\n",
				filePath, string(fileContent), errorStr,
			)

			fix, err := llm.GetCompletion(prompt)
			if err != nil {
				fmt.Printf("Error getting fix from LLM: %s\n", err)
				return
			}

			code := builder.ExtractCode(fix)

			diff, err := builder.GenerateDiff(filePath, code)
			if err != nil {
				fmt.Printf("Error generating diff: %s\n", err)
				return
			}

			fmt.Printf("\nSuggested changes:\n%s\n", diff)
			fmt.Print("Apply changes? (y/n): ")

			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')

			if strings.ToLower(strings.TrimSpace(response)) == "y" {
				fmt.Println("Applying fix...")
				if err := builder.ApplyFix(filePath, code); err != nil {
					fmt.Printf("Error applying fix: %s\n", err)
					return
				}
			} else {
				fmt.Println("Changes rejected.")
				return
			}
		}
		fmt.Println("Failed to fix the build after 3 attempts.")
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}