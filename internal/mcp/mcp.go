package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/entrepeneur4lyf/codeforge/internal/config"
)

// MCPVersion represents the MCP protocol version
const MCPVersion = "2024-11-05"

// Tool represents an MCP tool definition
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// Resource represents an MCP resource
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType,omitempty"`
}

// MCPRequest represents a request to an MCP server
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCPResponse represents a response from an MCP server
type MCPResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *MCPError       `json:"error,omitempty"`
}

// MCPError represents an MCP error
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ServerCapabilities represents MCP server capabilities
type ServerCapabilities struct {
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
	Prompts   *PromptsCapability   `json:"prompts,omitempty"`
	Logging   *LoggingCapability   `json:"logging,omitempty"`
}

// ToolsCapability represents tools capability
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ResourcesCapability represents resources capability
type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// PromptsCapability represents prompts capability
type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// LoggingCapability represents logging capability
type LoggingCapability struct {
	Level string `json:"level,omitempty"`
}

// Client represents an MCP client connection to a server
type Client struct {
	name         string
	cmd          *exec.Cmd
	stdin        io.WriteCloser
	stdout       io.ReadCloser
	stderr       io.ReadCloser
	capabilities ServerCapabilities
	tools        []Tool
	resources    []Resource

	// Request tracking
	nextID    int64
	requests  map[interface{}]chan *MCPResponse
	requestMu sync.RWMutex

	// State
	initialized bool
	mu          sync.RWMutex
}

// Manager manages multiple MCP clients
type Manager struct {
	clients   map[string]*Client
	clientsMu sync.RWMutex
	config    *config.Config
}

// Global manager instance
var manager *Manager

// Initialize sets up the MCP manager with configuration
func Initialize(cfg *config.Config) error {
	manager = &Manager{
		clients: make(map[string]*Client),
		config:  cfg,
	}

	// Initialize MCP clients based on configuration
	ctx := context.Background()
	for name, mcpConfig := range cfg.MCP {
		if len(mcpConfig.Command) > 0 {
			go manager.createAndStartMCPClient(ctx, name, mcpConfig)
		}
	}

	return nil
}

// GetManager returns the global MCP manager instance
func GetManager() *Manager {
	return manager
}

// createAndStartMCPClient creates and initializes an MCP client
func (m *Manager) createAndStartMCPClient(ctx context.Context, name string, mcpConfig config.MCPConfig) {
	client, err := NewClient(ctx, name, mcpConfig.Command[0], mcpConfig.Command[1:]...)
	if err != nil {
		fmt.Printf("Failed to create MCP client for %s: %v\n", name, err)
		return
	}

	// Initialize with timeout
	initCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := client.Initialize(initCtx); err != nil {
		fmt.Printf("Failed to initialize MCP client %s: %v\n", name, err)
		client.Close()
		return
	}

	// Store the client
	m.clientsMu.Lock()
	m.clients[name] = client
	m.clientsMu.Unlock()

	fmt.Printf("MCP client %s initialized successfully\n", name)
}

// NewClient creates a new MCP client
func NewClient(ctx context.Context, name, command string, args ...string) (*Client, error) {
	cmd := exec.CommandContext(ctx, command, args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdin.Close()
		stdout.Close()
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	client := &Client{
		name:     name,
		cmd:      cmd,
		stdin:    stdin,
		stdout:   stdout,
		stderr:   stderr,
		requests: make(map[interface{}]chan *MCPResponse),
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		stderr.Close()
		return nil, fmt.Errorf("failed to start MCP server: %w", err)
	}

	// Start message handling
	go client.handleMessages()

	return client, nil
}

// Initialize initializes the MCP client with the server
func (c *Client) Initialize(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.initialized {
		return nil
	}

	// Send initialize request
	initParams := map[string]interface{}{
		"protocolVersion": MCPVersion,
		"capabilities": map[string]interface{}{
			"roots": map[string]interface{}{
				"listChanged": true,
			},
			"sampling": map[string]interface{}{},
		},
		"clientInfo": map[string]interface{}{
			"name":    "codeforge",
			"version": "0.1.0",
		},
	}

	var result map[string]interface{}
	if err := c.Call(ctx, "initialize", initParams, &result); err != nil {
		return fmt.Errorf("initialize failed: %w", err)
	}

	// Parse server capabilities
	if caps, ok := result["capabilities"].(map[string]interface{}); ok {
		if err := json.Unmarshal([]byte(fmt.Sprintf("%v", caps)), &c.capabilities); err != nil {
			fmt.Printf("Warning: failed to parse server capabilities: %v\n", err)
		}
	}

	// Send initialized notification
	if err := c.Notify(ctx, "notifications/initialized", nil); err != nil {
		return fmt.Errorf("initialized notification failed: %w", err)
	}

	// Load available tools and resources
	if err := c.loadTools(ctx); err != nil {
		fmt.Printf("Warning: failed to load tools: %v\n", err)
	}

	if err := c.loadResources(ctx); err != nil {
		fmt.Printf("Warning: failed to load resources: %v\n", err)
	}

	c.initialized = true
	return nil
}

// loadTools loads available tools from the server
func (c *Client) loadTools(ctx context.Context) error {
	var result struct {
		Tools []Tool `json:"tools"`
	}

	if err := c.Call(ctx, "tools/list", nil, &result); err != nil {
		return err
	}

	c.tools = result.Tools
	return nil
}

// loadResources loads available resources from the server
func (c *Client) loadResources(ctx context.Context) error {
	var result struct {
		Resources []Resource `json:"resources"`
	}

	if err := c.Call(ctx, "resources/list", nil, &result); err != nil {
		return err
	}

	c.resources = result.Resources
	return nil
}

// GetTools returns the list of available tools
func (c *Client) GetTools() []Tool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]Tool, len(c.tools))
	copy(result, c.tools)
	return result
}

// GetResources returns the list of available resources
func (c *Client) GetResources() []Resource {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]Resource, len(c.resources))
	copy(result, c.resources)
	return result
}

// Close closes the MCP client connection
func (c *Client) Close() error {
	if c.cmd != nil && c.cmd.Process != nil {
		c.cmd.Process.Kill()
	}

	if c.stdin != nil {
		c.stdin.Close()
	}
	if c.stdout != nil {
		c.stdout.Close()
	}
	if c.stderr != nil {
		c.stderr.Close()
	}

	return nil
}

// GetClient returns an MCP client by name
func (m *Manager) GetClient(name string) *Client {
	m.clientsMu.RLock()
	defer m.clientsMu.RUnlock()

	return m.clients[name]
}

// ListClients returns all available MCP clients
func (m *Manager) ListClients() map[string]*Client {
	m.clientsMu.RLock()
	defer m.clientsMu.RUnlock()

	result := make(map[string]*Client)
	for name, client := range m.clients {
		result[name] = client
	}
	return result
}
