package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/entrepeneur4lyf/codeforge/internal/graph"
)

func main() {
	fmt.Println("🔍 Testing API Signature & Data Structure Extraction")
	fmt.Println("===================================================")

	// Create codebase manager
	manager := graph.NewCodebaseManager()

	// Initialize with current directory
	fmt.Println("📁 Initializing codebase awareness...")
	if err := manager.Initialize("."); err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}

	// Wait a moment for scanning to complete
	time.Sleep(2 * time.Second)

	// Test specific files that should have rich API information
	testFiles := []string{
		"internal/graph/types.go",
		"internal/graph/graph.go",
		"internal/graph/awareness.go",
		"internal/graph/scanner_simple.go",
		"internal/graph/hybrid_search.go",
	}

	fmt.Println("\n🔍 Testing File Context with API Signatures:")
	fmt.Println("============================================")

	for i, filePath := range testFiles {
		fmt.Printf("\n--- File %d: %s ---\n", i+1, filePath)

		start := time.Now()
		result := manager.GetFileContext(filePath)
		duration := time.Since(start)

		fmt.Printf("Query time: %v\n", duration)
		fmt.Println(result)

		if i < len(testFiles)-1 {
			fmt.Println("\n" + strings.Repeat("=", 80))
		}
	}

	// Test search for specific API patterns
	fmt.Println("\n🔍 Testing Search for API Patterns:")
	fmt.Println("===================================")

	apiQueries := []string{
		"NewCodeGraph",
		"AddNode",
		"GetNode",
		"APISignature",
		"DataStructure",
		"Parameter",
		"ReturnType",
	}

	ctx := context.Background()

	for i, query := range apiQueries {
		fmt.Printf("\n--- API Query %d: '%s' ---\n", i+1, query)

		start := time.Now()
		result := manager.SmartSearch(ctx, query)
		duration := time.Since(start)

		fmt.Printf("Query time: %v\n", duration)
		fmt.Println(result)

		if i < len(apiQueries)-1 {
			fmt.Println("\n" + strings.Repeat("-", 50))
		}
	}

	fmt.Println("\n✅ API extraction test completed!")
}
