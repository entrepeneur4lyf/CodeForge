package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func main() {
	fmt.Println("🌐 Checking OpenRouter API Directly")
	fmt.Println("===================================")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://openrouter.ai/api/v1/models", nil)
	if err != nil {
		fmt.Printf("❌ Error creating request: %v\n", err)
		return
	}

	req.Header.Set("User-Agent", "CodeForge/1.0")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Error making request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("📡 Response Status: %d\n", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ Bad status code: %d\n", resp.StatusCode)
		return
	}

	var response struct {
		Data []struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Created     int64  `json:"created"`
			Description string `json:"description"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Printf("❌ Error decoding JSON: %v\n", err)
		return
	}

	fmt.Printf("✅ Successfully fetched %d models\n", len(response.Data))

	// Show the first 15 models
	fmt.Println("\n📋 First 15 Models from OpenRouter API:")
	fmt.Println("=======================================")
	for i, model := range response.Data {
		if i >= 15 {
			break
		}
		
		// Convert timestamp to readable date
		created := time.Unix(model.Created, 0)
		
		fmt.Printf("%2d. %-45s | %s | %s\n", 
			i+1, 
			model.Name, 
			model.ID, 
			created.Format("2006-01-02"))
	}

	// Show some recent models (sort by created date)
	fmt.Println("\n🆕 Most Recent Models (by creation date):")
	fmt.Println("=========================================")
	
	// Simple sort by created timestamp (newest first)
	models := response.Data
	for i := 0; i < len(models)-1; i++ {
		for j := i + 1; j < len(models); j++ {
			if models[i].Created < models[j].Created {
				models[i], models[j] = models[j], models[i]
			}
		}
	}
	
	// Show top 10 newest
	for i, model := range models {
		if i >= 10 {
			break
		}
		
		created := time.Unix(model.Created, 0)
		fmt.Printf("%2d. %-45s | %s | %s\n", 
			i+1, 
			model.Name, 
			model.ID, 
			created.Format("2006-01-02"))
	}
}
