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
	fmt.Println("🔍 Testing Hybrid Search System")
	fmt.Println("===============================")

	// Create codebase manager
	manager := graph.NewCodebaseManager()

	// Initialize with current directory
	fmt.Println("📁 Initializing codebase awareness...")
	if err := manager.Initialize("."); err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}

	// Wait a moment for scanning to complete
	time.Sleep(2 * time.Second)

	// Test queries
	testQueries := []string{
		"model selector",
		"openrouter",
		"graph search",
		"chat handler",
		"vector store",
		"authentication",
		"error handling",
		"main function",
	}

	fmt.Println("\n🧠 Testing Graph-Based Search:")
	fmt.Println("==============================")

	ctx := context.Background()

	for i, query := range testQueries {
		fmt.Printf("\n--- Query %d: '%s' ---\n", i+1, query)

		start := time.Now()
		result := manager.SmartSearch(ctx, query)
		duration := time.Since(start)

		fmt.Printf("Query time: %v\n", duration)
		fmt.Println(result)

		if i < len(testQueries)-1 {
			fmt.Println("\n" + strings.Repeat("-", 50))
		}
	}

	// Test codebase stats
	fmt.Println("\n📊 Codebase Statistics:")
	fmt.Println("=======================")
	stats := manager.GetStats()
	fmt.Println(stats)

	// Test file context
	fmt.Println("\n📄 File Context Example:")
	fmt.Println("========================")
	fileContext := manager.GetFileContext("internal/graph/graph.go")
	fmt.Println(fileContext)

	// Test quick context
	fmt.Println("\n⚡ Quick Context Example:")
	fmt.Println("========================")
	quickContext := manager.GetQuickContext("how does the graph system work?")
	fmt.Println(quickContext)

	fmt.Println("\n✅ Hybrid search system test completed!")
}
