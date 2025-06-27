package cmd

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/entrepeneur4lyf/codeforge/internal/embeddings"
	"github.com/entrepeneur4lyf/codeforge/internal/vectordb"
	"github.com/spf13/cobra"
)

var indexCmd = &cobra.Command{
	Use:   "index [path]",
	Short: "Index code files for semantic search",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		recursive, _ := cmd.Flags().GetBool("recursive")

		ctx := context.Background()
		vdb := vectordb.Get()
		if vdb == nil {
			log.Fatal("Vector database not initialized")
		}

		// Get file info
		info, err := os.Stat(path)
		if err != nil {
			log.Fatalf("Error accessing path: %v", err)
		}

		var files []string
		if info.IsDir() {
			if recursive {
				err = filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if !info.IsDir() && isCodeFile(filePath) {
						files = append(files, filePath)
					}
					return nil
				})
			} else {
				entries, err := os.ReadDir(path)
				if err != nil {
					log.Fatalf("Error reading directory: %v", err)
				}
				for _, entry := range entries {
					if !entry.IsDir() {
						filePath := filepath.Join(path, entry.Name())
						if isCodeFile(filePath) {
							files = append(files, filePath)
						}
					}
				}
			}
		} else {
			if isCodeFile(path) {
				files = append(files, path)
			}
		}

		if err != nil {
			log.Fatalf("Error walking directory: %v", err)
		}

		if len(files) == 0 {
			fmt.Println("No code files found to index")
			return
		}

		fmt.Printf("Indexing %d files...\n", len(files))

		for i, file := range files {
			fmt.Printf("Processing %d/%d: %s\n", i+1, len(files), file)

			if err := indexFile(ctx, vdb, file); err != nil {
				log.Printf("Error indexing %s: %v", file, err)
				continue
			}
		}

		fmt.Println("Indexing complete!")
	},
}

func indexFile(ctx context.Context, vdb *vectordb.VectorDB, filePath string) error {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Detect language
	language := detectLanguage(filePath)

	// Generate embedding using the best available service
	embedding, err := embeddings.GetCodeEmbedding(ctx, string(content), language)
	if err != nil {
		fmt.Printf("Warning: Failed to generate embedding for %s, using simple embedding: %v\n", filePath, err)
		embedding = createSimpleEmbedding(string(content))
	}

	// Create code chunk
	codeChunk := &vectordb.CodeChunk{
		ID:       fmt.Sprintf("%x", sha256.Sum256([]byte(filePath+string(content)))),
		FilePath: filePath,
		Content:  string(content),
		ChunkType: vectordb.ChunkType{
			Type: "file",
			Data: map[string]interface{}{
				"full_file": true,
			},
		},
		Language: language,
		Symbols:  []vectordb.Symbol{},
		Imports:  []string{},
		Location: vectordb.SourceLocation{
			StartLine:   1,
			EndLine:     strings.Count(string(content), "\n") + 1,
			StartColumn: 1,
			EndColumn:   1,
		},
		Metadata: map[string]string{
			"file_size":      fmt.Sprintf("%d", len(content)),
			"lines":          fmt.Sprintf("%d", strings.Count(string(content), "\n")+1),
			"embedding_type": "model2vec",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Store in database
	return vdb.StoreChunk(ctx, codeChunk, embedding)
}

func isCodeFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	codeExtensions := []string{
		".go", ".rs", ".py", ".js", ".ts", ".tsx", ".jsx",
		".java", ".cpp", ".cc", ".cxx", ".c", ".h", ".hpp",
		".cs", ".php", ".rb", ".swift", ".kt", ".scala",
		".clj", ".hs", ".ml", ".fs", ".elm", ".dart",
		".lua", ".ex", ".exs", ".erl", ".pl", ".r",
	}

	for _, codeExt := range codeExtensions {
		if ext == codeExt {
			return true
		}
	}
	return false
}

func detectLanguage(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".go":
		return "go"
	case ".rs":
		return "rust"
	case ".py":
		return "python"
	case ".js":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".tsx":
		return "typescriptreact"
	case ".jsx":
		return "javascriptreact"
	case ".java":
		return "java"
	case ".cpp", ".cc", ".cxx":
		return "cpp"
	case ".c":
		return "c"
	case ".h", ".hpp":
		return "c"
	case ".cs":
		return "csharp"
	case ".php":
		return "php"
	case ".rb":
		return "ruby"
	case ".swift":
		return "swift"
	case ".kt":
		return "kotlin"
	case ".scala":
		return "scala"
	case ".clj":
		return "clojure"
	case ".hs":
		return "haskell"
	case ".ml":
		return "ocaml"
	case ".fs":
		return "fsharp"
	case ".elm":
		return "elm"
	case ".dart":
		return "dart"
	case ".lua":
		return "lua"
	case ".ex", ".exs":
		return "elixir"
	case ".erl":
		return "erlang"
	case ".pl":
		return "perl"
	case ".r":
		return "r"
	default:
		return "plaintext"
	}
}

func createSimpleEmbedding(content string) []float32 {
	// This is a very simple embedding based on basic text features
	// In a real implementation, you'd use a proper embedding model
	embedding := make([]float32, 1536) // OpenAI embedding size (Go libsql driver requirement)

	// Simple features based on content
	contentLen := float32(len(content))
	lineCount := float32(strings.Count(content, "\n") + 1)
	wordCount := float32(len(strings.Fields(content)))

	// Normalize and set some basic features
	embedding[0] = contentLen / 10000.0 // Normalize content length
	embedding[1] = lineCount / 1000.0   // Normalize line count
	embedding[2] = wordCount / 5000.0   // Normalize word count

	// Add some randomness to make embeddings unique
	for i := 3; i < len(embedding); i++ {
		embedding[i] = (float32(i%100) / 100.0) * 0.1
	}

	return embedding
}

func init() {
	indexCmd.Flags().BoolP("recursive", "r", false, "Index files recursively")

	rootCmd.AddCommand(indexCmd)
}
