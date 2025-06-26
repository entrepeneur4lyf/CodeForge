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

// VectorDB represents a vector database using libsql
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

	// Initialize database schema
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

	// Create code embeddings table with F32_BLOB vector column
	codeEmbeddingsSQL := `
	CREATE TABLE IF NOT EXISTS code_embeddings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_path TEXT NOT NULL,
		content TEXT NOT NULL,
		language TEXT NOT NULL,
		embedding vector(1536) NOT NULL, -- OpenAI embedding dimension (Go libsql driver requirement)
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

	// Create vector index exactly like the Rust implementation
	vectorIndexSQL := `
	CREATE INDEX IF NOT EXISTS idx_code_embeddings_vector
	ON code_embeddings(libsql_vector_idx(embedding,
		'metric=cosine',
		'max_neighbors=50',
		'search_l=400',
		'alpha=1.0'
	))
	`

	if _, err := vdb.db.ExecContext(ctx, vectorIndexSQL); err != nil {
		fmt.Printf("Warning: Vector index creation failed: %v\n", err)
	} else {
		fmt.Println("Vector index created successfully")
	}

	// Create error patterns table with F32_BLOB vector column
	errorPatternsSQL := `
	CREATE TABLE IF NOT EXISTS error_patterns (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		error_type TEXT NOT NULL,
		pattern TEXT NOT NULL,
		solution TEXT NOT NULL,
		language TEXT NOT NULL,
		embedding vector(256) NOT NULL, -- model2vec embedding dimension
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

	// Create project context table with F32_BLOB vector column
	projectContextSQL := `
	CREATE TABLE IF NOT EXISTS project_context (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		context_type TEXT NOT NULL,
		content TEXT NOT NULL,
		embedding vector(256) NOT NULL, -- model2vec embedding dimension
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
	query := `
	INSERT OR REPLACE INTO code_embeddings
	(file_path, content, language, embedding, metadata, updated_at)
	VALUES (?, ?, ?, vector(?), ?, CURRENT_TIMESTAMP)
	`

	// Convert embedding to JSON string for the vector() function (like Rust implementation)
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

// SearchSimilarCode searches for similar code snippets using vector similarity
func (vdb *VectorDB) SearchSimilarCode(ctx context.Context, queryEmbedding []float32, language string, limit int) ([]SearchResult, error) {
	// Performance monitoring
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if duration > 100*time.Millisecond {
			fmt.Printf("Warning: Vector search took %v (>100ms)\n", duration)
		}
	}()
	// Convert query embedding to JSON for vector function
	queryJSON, err := json.Marshal(queryEmbedding)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query embedding: %w", err)
	}

	// Use optimized vector_top_k with larger search space for better results
	searchLimit := limit * 3 // Search more candidates for better quality
	if searchLimit > 100 {
		searchLimit = 100 // Cap to avoid excessive computation
	}

	// Primary query using vector index for maximum performance
	vectorQuery := `
	SELECT ce.id, ce.content, vector_distance_cos(ce.embedding, vector(?)) as distance, ce.metadata
	FROM vector_top_k('idx_code_embeddings_vector', vector(?), ?) vt
	JOIN code_embeddings ce ON ce.id = vt.id
	WHERE ce.language = ? OR ? = ''
	ORDER BY distance ASC
	LIMIT ?
	`

	rows, err := vdb.db.QueryContext(ctx, vectorQuery,
		string(queryJSON), string(queryJSON), searchLimit,
		language, language, limit)
	if err != nil {
		// Optimized fallback without vector_top_k but still using vector functions
		fallbackQuery := `
		SELECT id, content, vector_distance_cos(embedding, vector(?)) as distance, metadata
		FROM code_embeddings
		WHERE (language = ? OR ? = '')
		ORDER BY distance ASC
		LIMIT ?
		`

		rows, err = vdb.db.QueryContext(ctx, fallbackQuery,
			string(queryJSON), language, language, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to execute vector search: %w", err)
		}
	}
	defer rows.Close()

	// Pre-allocate results slice for better performance
	results := make([]SearchResult, 0, limit)

	for rows.Next() {
		var id int64
		var content, metadata string
		var distance float64

		if err := rows.Scan(&id, &content, &distance, &metadata); err != nil {
			continue // Skip malformed rows
		}

		// Convert cosine distance to similarity score (1 - distance)
		similarity := 1.0 - distance

		// Filter out poor matches (similarity < 0.1)
		if similarity < 0.1 {
			continue
		}

		results = append(results, SearchResult{
			ID:         id,
			Content:    content,
			Similarity: similarity,
			Metadata:   metadata,
		})

		// Early termination when we have enough high-quality results
		if len(results) >= limit {
			break
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error processing search results: %w", err)
	}

	return results, nil
}

// SearchSimilarErrors searches for similar error patterns using vector similarity
func (vdb *VectorDB) SearchSimilarErrors(ctx context.Context, queryEmbedding []float32, language string, limit int) ([]SearchResult, error) {
	// Convert query embedding to JSON for the vector32() function
	queryJSON, err := json.Marshal(queryEmbedding)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query embedding: %w", err)
	}

	// Use vector_top_k for efficient similarity search
	query := `
	SELECT ep.id, ep.pattern, ep.solution, vector_distance_cos(ep.embedding, vector(?)) as distance, ep.metadata
	FROM vector_top_k('idx_error_patterns_vector', vector(?), ?) vt
	JOIN error_patterns ep ON ep.id = vt.id
	WHERE ep.language = ? OR ? = ''
	ORDER BY distance ASC
	`

	rows, err := vdb.db.QueryContext(ctx, query, string(queryJSON), string(queryJSON), limit*2, language, language)
	if err != nil {
		// Fallback to regular query if vector_top_k is not available
		fallbackQuery := `
		SELECT id, pattern, solution, vector_distance_cos(embedding, vector(?)) as distance, metadata
		FROM error_patterns
		WHERE language = ? OR ? = ''
		ORDER BY distance ASC
		LIMIT ?
		`

		rows, err = vdb.db.QueryContext(ctx, fallbackQuery, string(queryJSON), language, language, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to query error patterns: %w", err)
		}
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var id int64
		var pattern, solution, metadata string
		var distance float64

		if err := rows.Scan(&id, &pattern, &solution, &distance, &metadata); err != nil {
			continue
		}

		// Convert distance to similarity
		similarity := 1.0 - distance
		content := fmt.Sprintf("Pattern: %s\nSolution: %s", pattern, solution)

		results = append(results, SearchResult{
			ID:         id,
			Content:    content,
			Similarity: similarity,
			Metadata:   metadata,
		})

		if len(results) >= limit {
			break
		}
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
