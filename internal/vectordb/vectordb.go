package vectordb

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/entrepeneur4lyf/codeforge/internal/config"
	_ "github.com/tursodatabase/go-libsql"
)

// VectorDB provides vector database operations using libsql
// Uses basic similarity search instead of broken libsql-vectors extension
type VectorDB struct {
	db     *sql.DB
	config *config.Config
}

// CodeEmbedding represents a code snippet with its embedding
type CodeEmbedding struct {
	ID        int64     `json:"id"`
	FilePath  string    `json:"file_path"`
	Content   string    `json:"content"`
	Language  string    `json:"language"`
	Embedding []float32 `json:"embedding"`
	Metadata  string    `json:"metadata"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ErrorPattern represents an error pattern with its solution
type ErrorPattern struct {
	ID        int64     `json:"id"`
	ErrorType string    `json:"error_type"`
	Pattern   string    `json:"pattern"`
	Solution  string    `json:"solution"`
	Language  string    `json:"language"`
	Embedding []float32 `json:"embedding"`
	Metadata  string    `json:"metadata"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SearchResult represents a search result with similarity score
type SearchResult struct {
	ID         int64   `json:"id"`
	Content    string  `json:"content"`
	Similarity float64 `json:"similarity"`
	Metadata   string  `json:"metadata"`
}

// Global vector database instance
var vectorDB *VectorDB

// Initialize sets up the vector database
func Initialize(cfg *config.Config) error {
	// Create data directory if it doesn't exist
	dataDir := cfg.Data.Directory
	if !filepath.IsAbs(dataDir) {
		dataDir = filepath.Join(cfg.WorkingDir, dataDir)
	}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Database path
	dbPath := filepath.Join(dataDir, "vectors.db")

	// Connect to libsql database using sql.Open
	db, err := sql.Open("libsql", "file:"+dbPath)
	if err != nil {
		return fmt.Errorf("failed to open libsql database: %w", err)
	}

	vectorDB = &VectorDB{
		db:     db,
		config: cfg,
	}

	// Initialize libsql database schema (for metadata)
	if err := vectorDB.initializeSchema(); err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	return nil
}

// Get returns the global vector database instance
func Get() *VectorDB {
	return vectorDB
}

// initializeSchema creates the necessary tables
func (vdb *VectorDB) initializeSchema() error {
	ctx := context.Background()

	// Create code embeddings table with JSON embedding storage (avoids libsql-vectors issues)
	codeEmbeddingsSQL := `
	CREATE TABLE IF NOT EXISTS code_embeddings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_path TEXT NOT NULL,
		content TEXT NOT NULL,
		language TEXT NOT NULL,
		embedding_json TEXT NOT NULL, -- JSON format for embeddings (no libsql-vectors needed)
		metadata TEXT DEFAULT '{}',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_code_embeddings_file_path ON code_embeddings(file_path);
	CREATE INDEX IF NOT EXISTS idx_code_embeddings_language ON code_embeddings(language);
	`

	if _, err := vdb.db.ExecContext(ctx, codeEmbeddingsSQL); err != nil {
		return fmt.Errorf("failed to create code_embeddings table: %w", err)
	}

	// Vector indexes disabled (using JSON storage instead of libsql-vectors)
	fmt.Println("✅ Code embeddings table created (JSON storage mode)")

	// Create error patterns table with JSON embedding storage (avoids libsql-vectors issues)
	errorPatternsSQL := `
	CREATE TABLE IF NOT EXISTS error_patterns (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		error_type TEXT NOT NULL,
		pattern TEXT NOT NULL,
		solution TEXT NOT NULL,
		language TEXT NOT NULL,
		embedding_json TEXT NOT NULL, -- JSON format for embeddings (no libsql-vectors needed)
		metadata TEXT DEFAULT '{}',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_error_patterns_error_type ON error_patterns(error_type);
	CREATE INDEX IF NOT EXISTS idx_error_patterns_language ON error_patterns(language);
	`

	if _, err := vdb.db.ExecContext(ctx, errorPatternsSQL); err != nil {
		return fmt.Errorf("failed to create error_patterns table: %w", err)
	}

	// Vector indexes disabled for error patterns (using JSON storage)
	fmt.Println("Vector indexes disabled for error patterns")

	// Create project context table with JSON embedding storage (avoids libsql-vectors issues)
	projectContextSQL := `
	CREATE TABLE IF NOT EXISTS project_context (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		context_type TEXT NOT NULL,
		content TEXT NOT NULL,
		embedding_json TEXT NOT NULL, -- JSON format for embeddings (no libsql-vectors needed)
		metadata TEXT DEFAULT '{}',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_project_context_type ON project_context(context_type);
	`

	if _, err := vdb.db.ExecContext(ctx, projectContextSQL); err != nil {
		return fmt.Errorf("failed to create project_context table: %w", err)
	}

	// Vector indexes disabled for project context (using JSON storage)
	fmt.Println("Vector indexes disabled for project context")

	return nil
}

// StoreCodeEmbedding stores a code embedding in the database
func (vdb *VectorDB) StoreCodeEmbedding(ctx context.Context, embedding *CodeEmbedding) error {
	// Store embedding with JSON format (avoid broken libsql-vectors)
	query := `
	INSERT OR REPLACE INTO code_embeddings
	(file_path, content, language, embedding_json, metadata, updated_at)
	VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`

	// Convert embedding to JSON for storage
	embeddingJSON, err := json.Marshal(embedding.Embedding)
	if err != nil {
		return fmt.Errorf("failed to marshal embedding: %w", err)
	}

	result, err := vdb.db.ExecContext(ctx, query,
		embedding.FilePath,
		embedding.Content,
		embedding.Language,
		string(embeddingJSON),
		embedding.Metadata,
	)
	if err != nil {
		return fmt.Errorf("failed to store code embedding: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	embedding.ID = id
	return nil
}

// StoreErrorPattern stores an error pattern in the database
func (vdb *VectorDB) StoreErrorPattern(ctx context.Context, pattern *ErrorPattern) error {
	query := `
	INSERT OR REPLACE INTO error_patterns
	(error_type, pattern, solution, language, embedding, metadata, updated_at)
	VALUES (?, ?, ?, ?, vector(?), ?, CURRENT_TIMESTAMP)
	`

	// Convert embedding to JSON for the vector32() function
	embeddingJSON, err := json.Marshal(pattern.Embedding)
	if err != nil {
		return fmt.Errorf("failed to marshal embedding: %w", err)
	}

	result, err := vdb.db.ExecContext(ctx, query,
		pattern.ErrorType,
		pattern.Pattern,
		pattern.Solution,
		pattern.Language,
		string(embeddingJSON),
		pattern.Metadata,
	)
	if err != nil {
		return fmt.Errorf("failed to store error pattern: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	pattern.ID = id
	return nil
}

// SearchSimilarCode searches for similar code snippets using hybrid approach:
// - ChromaDB for fast vector similarity search (if available)
// - libsql fallback for basic text search
func (vdb *VectorDB) SearchSimilarCode(ctx context.Context, queryEmbedding []float32, language string, limit int) ([]SearchResult, error) {
	// Performance monitoring
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if duration > 100*time.Millisecond {
			fmt.Printf("Warning: Vector search took %v (>100ms)\n", duration)
		}
	}()

	// Use basic cosine similarity search (no libsql-vectors needed)
	query := `
	SELECT id, content, embedding_json, metadata
	FROM code_embeddings
	WHERE (language = ? OR ? = '')
	ORDER BY id DESC
	LIMIT ?
	`

	rows, err := vdb.db.QueryContext(ctx, query, language, language, limit*10) // Get more for filtering
	if err != nil {
		return nil, fmt.Errorf("failed to query embeddings: %w", err)
	}
	defer rows.Close()

	var candidates []struct {
		ID         int64
		Content    string
		Embedding  []float32
		Metadata   string
		Similarity float64
	}

	// Load all candidates and calculate similarity
	for rows.Next() {
		var id int64
		var content, embeddingJSON, metadata string

		if err := rows.Scan(&id, &content, &embeddingJSON, &metadata); err != nil {
			continue // Skip invalid rows
		}

		// Parse embedding JSON
		var embedding []float32
		if err := json.Unmarshal([]byte(embeddingJSON), &embedding); err != nil {
			continue // Skip invalid embeddings
		}

		// Calculate cosine similarity
		similarity := cosineSimilarity(queryEmbedding, embedding)

		candidates = append(candidates, struct {
			ID         int64
			Content    string
			Embedding  []float32
			Metadata   string
			Similarity float64
		}{
			ID:         id,
			Content:    content,
			Embedding:  embedding,
			Metadata:   metadata,
			Similarity: similarity,
		})
	}

	// Sort by similarity (highest first) and return top results
	for i := 0; i < len(candidates)-1; i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[i].Similarity < candidates[j].Similarity {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	// Convert to SearchResult format
	var results []SearchResult
	maxResults := limit
	if maxResults > len(candidates) {
		maxResults = len(candidates)
	}

	for i := 0; i < maxResults; i++ {
		results = append(results, SearchResult{
			ID:         candidates[i].ID,
			Content:    candidates[i].Content,
			Similarity: candidates[i].Similarity,
			Metadata:   candidates[i].Metadata,
		})
	}

	return results, nil
}

// cosineSimilarity calculates the cosine similarity between two vectors
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0.0 || normB == 0.0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// SearchSimilarErrors searches for similar error patterns using basic similarity
func (vdb *VectorDB) SearchSimilarErrors(ctx context.Context, queryEmbedding []float32, language string, limit int) ([]SearchResult, error) {
	// Similar implementation to SearchSimilarCode but for error_patterns table
	query := `
	SELECT id, pattern, embedding_json, metadata
	FROM error_patterns
	WHERE (language = ? OR ? = '')
	ORDER BY id DESC
	LIMIT ?
	`

	rows, err := vdb.db.QueryContext(ctx, query, language, language, limit*10)
	if err != nil {
		return nil, fmt.Errorf("failed to query error patterns: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var id int64
		var pattern, embeddingJSON, metadata string

		if err := rows.Scan(&id, &pattern, &embeddingJSON, &metadata); err != nil {
			continue
		}

		var embedding []float32
		if err := json.Unmarshal([]byte(embeddingJSON), &embedding); err != nil {
			continue
		}

		similarity := cosineSimilarity(queryEmbedding, embedding)

		results = append(results, SearchResult{
			ID:         id,
			Content:    pattern,
			Similarity: similarity,
			Metadata:   metadata,
		})
	}

	return results, nil
}

// Close closes the database connection
func (vdb *VectorDB) Close() error {
	if vdb.db != nil {
		return vdb.db.Close()
	}
	return nil
}
