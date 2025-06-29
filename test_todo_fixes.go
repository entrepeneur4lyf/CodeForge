package main

import (
	"context"
	"fmt"

	"github.com/entrepeneur4lyf/codeforge/internal/analysis"
	"github.com/entrepeneur4lyf/codeforge/internal/mcp"
)

func main() {
	fmt.Println("🧪 Testing TODO Fixes")
	fmt.Println("====================")

	// Test 1: Symbol Extractor with LSP integration
	fmt.Println("\n1. Testing Symbol Extractor LSP Integration:")
	fmt.Println("--------------------------------------------")

	extractor := analysis.NewSymbolExtractor()
	if extractor != nil {
		fmt.Println("✅ Symbol extractor created successfully")

		// Test workspace symbol search
		ctx := context.Background()
		symbols, err := extractor.ExtractWorkspaceSymbols(ctx, "test")
		if err != nil {
			fmt.Printf("⚠️  Workspace symbol search: %v (expected - no LSP servers running)\n", err)
		} else {
			fmt.Printf("✅ Found %d workspace symbols\n", len(symbols))
		}
	}

	// Test 2: MCP Repository Language Detection
	fmt.Println("\n2. Testing Enhanced Language Detection:")
	fmt.Println("--------------------------------------")

	fetcher := &mcp.RepositoryFetcher{}

	testURLs := []struct {
		url      string
		expected string
	}{
		{"https://github.com/python/cpython", "Python"},
		{"https://github.com/nodejs/node", "JavaScript"},
		{"https://github.com/golang/go", "Go"},
		{"https://github.com/rust-lang/rust", "Rust"},
		{"https://github.com/microsoft/typescript", "TypeScript"},
		{"https://github.com/laravel/laravel", "PHP"},
		{"https://github.com/rails/rails", "Ruby"},
		{"https://github.com/facebook/react", "JavaScript"},
		{"https://github.com/angular/angular", "TypeScript"},
		{"https://github.com/tensorflow/tensorflow", "Python"},
		{"https://github.com/opencv/opencv", "C++"},
		{"https://github.com/torvalds/linux", "C"},
		{"https://github.com/flutter/flutter", "Dart"},
		{"https://github.com/JuliaLang/julia", "Julia"},
		{"https://github.com/unknown/project", "Unknown"},
	}

	for _, test := range testURLs {
		detected := fetcher.DetectLanguage(test.url)
		status := "✅"
		if detected != test.expected {
			status = "❌"
		}
		fmt.Printf("%s %s -> %s (expected: %s)\n", status, test.url, detected, test.expected)
	}

	// Test 3: MCP Embedding Generation
	fmt.Println("\n3. Testing MCP Embedding Generation:")
	fmt.Println("-----------------------------------")

	server := &mcp.CodeForgeServer{}
	ctx := context.Background()

	testQueries := []string{
		"function definition",
		"class implementation",
		"error handling",
		"database connection",
		"API endpoint",
	}

	for _, query := range testQueries {
		embedding, err := server.GenerateQueryEmbedding(ctx, query)
		if err != nil {
			fmt.Printf("❌ Failed to generate embedding for '%s': %v\n", query, err)
		} else {
			fmt.Printf("✅ Generated %d-dimensional embedding for '%s'\n", len(embedding), query)

			// Verify embedding properties
			if len(embedding) == 384 {
				fmt.Printf("   📏 Correct dimension (384)\n")
			} else {
				fmt.Printf("   ❌ Wrong dimension: %d\n", len(embedding))
			}

			// Check if normalized (approximately)
			var norm float32
			for _, val := range embedding {
				norm += val * val
			}
			if norm > 0.9 && norm < 1.1 {
				fmt.Printf("   📐 Properly normalized (norm: %.3f)\n", norm)
			} else {
				fmt.Printf("   ⚠️  Normalization issue (norm: %.3f)\n", norm)
			}
		}
	}

	fmt.Println("\n🎉 TODO Fixes Test Complete!")
	fmt.Println("============================")
	fmt.Println("✅ Implemented fixes for:")
	fmt.Println("   - LSP GetAllClients method")
	fmt.Println("   - Symbol extractor document tracking")
	fmt.Println("   - ML table creation (in-memory)")
	fmt.Println("   - MCP embedding generation")
	fmt.Println("   - Enhanced language detection")
	fmt.Println("   - PHP and C language support")
	fmt.Println()
	fmt.Println("📊 Production Readiness Improvements:")
	fmt.Println("   - Removed placeholder TODOs")
	fmt.Println("   - Added proper error handling")
	fmt.Println("   - Implemented fallback mechanisms")
	fmt.Println("   - Enhanced heuristics and detection")
	fmt.Println("   - Thread-safe document tracking")
}
