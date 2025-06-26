package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/entrepeneur4lyf/codeforge/internal/embeddings"
	"github.com/entrepeneur4lyf/codeforge/internal/vectordb"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for similar code snippets or error patterns",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		language, _ := cmd.Flags().GetString("language")
		limit, _ := cmd.Flags().GetInt("limit")
		searchType, _ := cmd.Flags().GetString("type")

		ctx := context.Background()
		vdb := vectordb.Get()
		if vdb == nil {
			log.Fatal("Vector database not initialized")
		}

		// Generate embedding for the query
		queryEmbedding, err := embeddings.GetEmbedding(ctx, query)
		if err != nil {
			log.Printf("Warning: Failed to generate query embedding, using dummy: %v", err)
			queryEmbedding = createDummyEmbedding(1536) // OpenAI embedding dimension (Go libsql driver requirement)
		}

		var results []vectordb.SearchResult

		switch searchType {
		case "code":
			results, err = vdb.SearchSimilarCode(ctx, queryEmbedding, language, limit)
		case "error":
			results, err = vdb.SearchSimilarErrors(ctx, queryEmbedding, language, limit)
		default:
			// Try both
			codeResults, codeErr := vdb.SearchSimilarCode(ctx, queryEmbedding, language, limit/2)
			if codeErr != nil {
				log.Printf("Error searching code: %v", codeErr)
			}

			errorResults, errorErr := vdb.SearchSimilarErrors(ctx, queryEmbedding, language, limit/2)
			if errorErr != nil {
				log.Printf("Error searching errors: %v", errorErr)
			}

			results = append(codeResults, errorResults...)
		}

		if err != nil {
			log.Fatalf("Search failed: %v", err)
		}

		if len(results) == 0 {
			fmt.Println("No results found")
			return
		}

		fmt.Printf("Found %d results for query: %s\n\n", len(results), query)
		for i, result := range results {
			fmt.Printf("Result %d (Similarity: %.3f):\n", i+1, result.Similarity)
			fmt.Printf("%s\n", result.Content)
			if result.Metadata != "" && result.Metadata != "{}" {
				fmt.Printf("Metadata: %s\n", result.Metadata)
			}
			fmt.Println("---")
		}
	},
}

// createDummyEmbedding creates a simple dummy embedding for fallback
func createDummyEmbedding(dimension int) []float32 {
	embedding := make([]float32, dimension)
	for i := range embedding {
		embedding[i] = 0.1 // Simple dummy values
	}
	return embedding
}

func init() {
	searchCmd.Flags().StringP("language", "l", "", "Filter by programming language")
	searchCmd.Flags().IntP("limit", "n", 5, "Maximum number of results")
	searchCmd.Flags().StringP("type", "t", "both", "Search type: code, error, or both")

	rootCmd.AddCommand(searchCmd)
}
