package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/entrepeneur4lyf/codeforge/internal/graph"
	"github.com/entrepeneur4lyf/codeforge/internal/ml"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	fmt.Println("🧠 Testing ML-Powered Code Intelligence")
	fmt.Println("======================================")

	// Create in-memory database for Q-learning
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create codebase manager
	manager := graph.NewCodebaseManager()

	// Initialize with current directory
	fmt.Println("📁 Initializing codebase awareness...")
	if err := manager.Initialize("."); err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}

	// Wait for scanning to complete
	time.Sleep(2 * time.Second)

	// Get the graph from the manager (we'll need to add a getter method)
	fmt.Println("🧠 Creating ML intelligence system...")
	
	// For now, create a new graph (in real implementation, we'd get it from manager)
	codeGraph := graph.NewCodeGraph(".")
	scanner := graph.NewSimpleScanner(codeGraph)
	if err := scanner.ScanRepository("."); err != nil {
		log.Fatalf("Failed to scan repository: %v", err)
	}

	// Create ML intelligence system
	intelligence, err := ml.NewCodeIntelligence(codeGraph, db)
	if err != nil {
		log.Fatalf("Failed to create ML intelligence: %v", err)
	}

	fmt.Println("✅ ML intelligence system initialized!")

	// Test queries for ML-enhanced search
	testQueries := []string{
		"graph search algorithm",
		"API signature extraction", 
		"database operations",
		"file watcher",
		"hybrid search",
		"MCTS implementation",
		"Q-learning table",
	}

	ctx := context.Background()

	fmt.Println("\n🔍 Testing ML-Enhanced Smart Search:")
	fmt.Println("===================================")

	for i, query := range testQueries {
		fmt.Printf("\n--- ML Query %d: '%s' ---\n", i+1, query)
		
		start := time.Now()
		
		// Get intelligent context using ML
		context, err := intelligence.GetIntelligentContext(ctx, query, 5)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		
		duration := time.Since(start)
		fmt.Printf("Query time: %v\n", duration)
		fmt.Println(context)
		
		// Simulate user feedback (random for demo)
		userFeedback := 0.7 + 0.3*float64(i%3)/2.0 // Varies between 0.7-1.0
		
		// Learn from feedback
		err = intelligence.LearnFromFeedback(ctx, query, []string{}, userFeedback)
		if err != nil {
			fmt.Printf("Learning error: %v\n", err)
		} else {
			fmt.Printf("📚 Learned from feedback: %.2f\n", userFeedback)
		}
		
		if i < len(testQueries)-1 {
			fmt.Println("\n" + strings.Repeat("-", 60))
		}
	}

	// Test ML statistics
	fmt.Println("\n📊 ML Performance Statistics:")
	fmt.Println("=============================")
	
	stats, err := intelligence.GetMLStats(ctx)
	if err != nil {
		fmt.Printf("Error getting stats: %v\n", err)
	} else {
		fmt.Printf("**Q-Learning Entries:** %d\n", stats.QLearningTotalEntries)
		fmt.Printf("**Average Q-Value:** %.3f\n", stats.QLearningAverageQ)
		fmt.Printf("**Current Epsilon:** %.3f\n", stats.QLearningCurrentEpsilon)
		fmt.Printf("**Total Experiences:** %d\n", stats.TotalExperiences)
		fmt.Printf("**Average Reward:** %.3f\n", stats.AverageReward)
		fmt.Printf("**Last Updated:** %s\n", stats.LastUpdated.Format("2006-01-02 15:04:05"))
	}

	// Test configuration updates
	fmt.Println("\n⚙️ Testing ML Configuration:")
	fmt.Println("============================")
	
	config := ml.DefaultMLConfig()
	fmt.Printf("**Default Learning Rate:** %.3f\n", config.LearningRate)
	fmt.Printf("**Default Epsilon:** %.3f\n", config.Epsilon)
	fmt.Printf("**MCTS Iterations:** %d\n", config.MCTSIterations)
	
	// Update configuration
	config.LearningRate = 0.05
	config.Epsilon = 0.2
	intelligence.UpdateConfig(config)
	fmt.Println("✅ Configuration updated!")

	// Test enabling/disabling ML
	fmt.Println("\n🔧 Testing ML Enable/Disable:")
	fmt.Println("=============================")
	
	intelligence.SetEnabled(false)
	fmt.Println("ML disabled - testing fallback...")
	
	fallbackResult, err := intelligence.SmartSearch(ctx, "test query", "")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Fallback result: %s\n", fallbackResult.Explanation)
	}
	
	intelligence.SetEnabled(true)
	fmt.Println("ML re-enabled!")

	// Demonstrate learning over time
	fmt.Println("\n📈 Demonstrating Learning Over Time:")
	fmt.Println("===================================")
	
	learningQuery := "code analysis"
	
	for round := 1; round <= 3; round++ {
		fmt.Printf("\n--- Learning Round %d ---\n", round)
		
		result, err := intelligence.SmartSearch(ctx, learningQuery, "")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		
		fmt.Printf("Confidence: %.3f\n", result.Confidence)
		fmt.Printf("Explanation: %s\n", result.Explanation)
		
		// Provide positive feedback to improve learning
		feedback := 0.8 + 0.1*float64(round) // Increasing feedback
		err = intelligence.LearnFromFeedback(ctx, learningQuery, result.BestPath, feedback)
		if err != nil {
			fmt.Printf("Learning error: %v\n", err)
		} else {
			fmt.Printf("Feedback provided: %.2f\n", feedback)
		}
	}

	// Final statistics
	fmt.Println("\n📊 Final ML Statistics:")
	fmt.Println("======================")
	
	finalStats, err := intelligence.GetMLStats(ctx)
	if err != nil {
		fmt.Printf("Error getting final stats: %v\n", err)
	} else {
		fmt.Printf("**Total Q-Learning Entries:** %d\n", finalStats.QLearningTotalEntries)
		fmt.Printf("**Total Experiences Collected:** %d\n", finalStats.TotalExperiences)
		fmt.Printf("**Final Average Reward:** %.3f\n", finalStats.AverageReward)
		fmt.Printf("**Learning Progress:** %.1f%%\n", finalStats.LearningProgress*100)
	}

	fmt.Println("\n✅ ML-powered code intelligence test completed!")
	fmt.Println("\n🎯 Key Features Demonstrated:")
	fmt.Println("- MCTS-based intelligent code exploration")
	fmt.Println("- Q-Learning with database persistence") 
	fmt.Println("- Adaptive learning from user feedback")
	fmt.Println("- Hybrid ML + graph-based search")
	fmt.Println("- Real-time performance metrics")
	fmt.Println("- Configurable ML parameters")
}
