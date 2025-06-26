package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/entrepeneur4lyf/codeforge/internal/mcp"
)

// mcpCmd represents the mcp command
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Model Context Protocol operations",
	Long: `Model Context Protocol operations for external tool integration including:
- Tool discovery and execution
- Resource access and management
- Server communication and monitoring`,
}

// mcpListCmd lists available MCP servers, tools, and resources
var mcpListCmd = &cobra.Command{
	Use:   "list",
	Short: "List MCP servers, tools, and resources",
	Long:  "List all available MCP servers and their capabilities",
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := mcp.GetManager()
		if manager == nil {
			return fmt.Errorf("MCP manager not initialized")
		}

		clients := manager.ListClients()
		if len(clients) == 0 {
			fmt.Println("No MCP servers configured or running")
			return nil
		}

		fmt.Printf("MCP Servers (%d):\n\n", len(clients))

		for name, client := range clients {
			fmt.Printf("🔧 %s\n", name)
			
			if !client.IsInitialized() {
				fmt.Println("   Status: Not initialized")
				fmt.Println()
				continue
			}

			fmt.Println("   Status: ✅ Ready")
			
			// Show capabilities
			caps := client.GetCapabilities()
			fmt.Print("   Capabilities: ")
			var capList []string
			if caps.Tools != nil {
				capList = append(capList, "Tools")
			}
			if caps.Resources != nil {
				capList = append(capList, "Resources")
			}
			if caps.Prompts != nil {
				capList = append(capList, "Prompts")
			}
			if caps.Logging != nil {
				capList = append(capList, "Logging")
			}
			fmt.Println(strings.Join(capList, ", "))

			// Show tools
			tools := client.GetTools()
			if len(tools) > 0 {
				fmt.Printf("   Tools (%d):\n", len(tools))
				for _, tool := range tools {
					fmt.Printf("     • %s - %s\n", tool.Name, tool.Description)
				}
			}

			// Show resources
			resources := client.GetResources()
			if len(resources) > 0 {
				fmt.Printf("   Resources (%d):\n", len(resources))
				for _, resource := range resources {
					fmt.Printf("     • %s (%s) - %s\n", resource.Name, resource.URI, resource.Description)
				}
			}

			fmt.Println()
		}

		return nil
	},
}

// mcpToolsCmd lists available tools
var mcpToolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "List available MCP tools",
	Long:  "List all available tools from all MCP servers",
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := mcp.GetManager()
		if manager == nil {
			return fmt.Errorf("MCP manager not initialized")
		}

		allTools := manager.GetAllTools()
		if len(allTools) == 0 {
			fmt.Println("No MCP tools available")
			return nil
		}

		totalTools := 0
		for _, tools := range allTools {
			totalTools += len(tools)
		}

		fmt.Printf("Available MCP Tools (%d):\n\n", totalTools)

		for serverName, tools := range allTools {
			if len(tools) == 0 {
				continue
			}

			fmt.Printf("📦 %s (%d tools):\n", serverName, len(tools))
			for _, tool := range tools {
				fmt.Printf("  🔧 %s\n", tool.Name)
				fmt.Printf("     %s\n", tool.Description)
				
				// Show input schema if available
				if tool.InputSchema != nil {
					if props, ok := tool.InputSchema["properties"].(map[string]interface{}); ok {
						fmt.Print("     Parameters: ")
						var params []string
						for param := range props {
							params = append(params, param)
						}
						fmt.Println(strings.Join(params, ", "))
					}
				}
				fmt.Println()
			}
		}

		return nil
	},
}

// mcpCallCmd calls an MCP tool
var mcpCallCmd = &cobra.Command{
	Use:   "call <tool-name> [arguments...]",
	Short: "Call an MCP tool",
	Long:  "Execute an MCP tool with the specified arguments",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		toolName := args[0]
		
		// Parse arguments as JSON key=value pairs
		arguments := make(map[string]interface{})
		for _, arg := range args[1:] {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid argument format: %s (expected key=value)", arg)
			}
			
			key := parts[0]
			value := parts[1]
			
			// Try to parse as JSON, fallback to string
			var jsonValue interface{}
			if err := json.Unmarshal([]byte(value), &jsonValue); err == nil {
				arguments[key] = jsonValue
			} else {
				arguments[key] = value
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		manager := mcp.GetManager()
		if manager == nil {
			return fmt.Errorf("MCP manager not initialized")
		}

		fmt.Printf("Calling tool: %s\n", toolName)
		if len(arguments) > 0 {
			fmt.Printf("Arguments: %v\n", arguments)
		}
		fmt.Println()

		result, err := manager.CallToolByName(ctx, toolName, arguments)
		if err != nil {
			return fmt.Errorf("tool call failed: %w", err)
		}

		if result.IsError {
			fmt.Println("❌ Tool execution failed:")
		} else {
			fmt.Println("✅ Tool execution successful:")
		}

		for i, content := range result.Content {
			if i > 0 {
				fmt.Println("---")
			}
			
			switch content.Type {
			case "text":
				fmt.Println(content.Text)
			case "image":
				fmt.Printf("Image content (%s)\n", content.MimeType)
			case "resource":
				fmt.Printf("Resource content: %v\n", content.Data)
			default:
				fmt.Printf("Content type: %s\n", content.Type)
				if content.Text != "" {
					fmt.Println(content.Text)
				}
				if content.Data != nil {
					fmt.Printf("Data: %v\n", content.Data)
				}
			}
		}

		return nil
	},
}

// mcpResourcesCmd lists available resources
var mcpResourcesCmd = &cobra.Command{
	Use:   "resources",
	Short: "List available MCP resources",
	Long:  "List all available resources from all MCP servers",
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := mcp.GetManager()
		if manager == nil {
			return fmt.Errorf("MCP manager not initialized")
		}

		allResources := manager.GetAllResources()
		if len(allResources) == 0 {
			fmt.Println("No MCP resources available")
			return nil
		}

		totalResources := 0
		for _, resources := range allResources {
			totalResources += len(resources)
		}

		fmt.Printf("Available MCP Resources (%d):\n\n", totalResources)

		for serverName, resources := range allResources {
			if len(resources) == 0 {
				continue
			}

			fmt.Printf("📦 %s (%d resources):\n", serverName, len(resources))
			for _, resource := range resources {
				fmt.Printf("  📄 %s\n", resource.Name)
				fmt.Printf("     URI: %s\n", resource.URI)
				fmt.Printf("     Description: %s\n", resource.Description)
				if resource.MimeType != "" {
					fmt.Printf("     Type: %s\n", resource.MimeType)
				}
				fmt.Println()
			}
		}

		return nil
	},
}

// mcpReadCmd reads an MCP resource
var mcpReadCmd = &cobra.Command{
	Use:   "read <server-name> <resource-uri>",
	Short: "Read an MCP resource",
	Long:  "Read the content of an MCP resource from a specific server",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverName := args[0]
		resourceURI := args[1]

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		manager := mcp.GetManager()
		if manager == nil {
			return fmt.Errorf("MCP manager not initialized")
		}

		client := manager.GetClient(serverName)
		if client == nil {
			return fmt.Errorf("MCP server not found: %s", serverName)
		}

		if !client.IsInitialized() {
			return fmt.Errorf("MCP server not initialized: %s", serverName)
		}

		fmt.Printf("Reading resource: %s\n", resourceURI)
		fmt.Printf("From server: %s\n", serverName)
		fmt.Println()

		content, err := client.ReadResource(ctx, resourceURI)
		if err != nil {
			return fmt.Errorf("failed to read resource: %w", err)
		}

		fmt.Printf("Content Type: %s\n", content.MimeType)
		fmt.Printf("URI: %s\n", content.URI)
		fmt.Println("---")

		if content.Text != "" {
			fmt.Println(content.Text)
		} else if len(content.Blob) > 0 {
			fmt.Printf("Binary content (%d bytes)\n", len(content.Blob))
		} else {
			fmt.Println("No content available")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
	
	// Add subcommands
	mcpCmd.AddCommand(mcpListCmd)
	mcpCmd.AddCommand(mcpToolsCmd)
	mcpCmd.AddCommand(mcpCallCmd)
	mcpCmd.AddCommand(mcpResourcesCmd)
	mcpCmd.AddCommand(mcpReadCmd)
}
