package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
)

// handleMessages reads and processes messages from the MCP server
func (c *Client) handleMessages() {
	scanner := bufio.NewScanner(c.stdout)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var response MCPResponse
		if err := json.Unmarshal([]byte(line), &response); err != nil {
			fmt.Printf("Failed to parse MCP response: %v\n", err)
			continue
		}

		// Handle response
		if response.ID != nil {
			c.handleResponse(&response)
		} else {
			// Handle notification
			c.handleNotification(&response)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading from MCP server: %v\n", err)
	}
}

// handleResponse handles a response message
func (c *Client) handleResponse(response *MCPResponse) {
	c.requestMu.RLock()
	ch, exists := c.requests[response.ID]
	c.requestMu.RUnlock()

	if exists {
		select {
		case ch <- response:
		default:
			// Channel full, drop response
		}
	}
}

// handleNotification handles a notification message
func (c *Client) handleNotification(response *MCPResponse) {
	// Handle server notifications (tools/list changed, etc.)
	// For now, just log them
	fmt.Printf("MCP notification: %s\n", response.Method)
}

// Call makes a request and waits for the response
func (c *Client) Call(ctx context.Context, method string, params interface{}, result interface{}) error {
	id := atomic.AddInt64(&c.nextID, 1)

	request := MCPRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	// Create response channel
	ch := make(chan *MCPResponse, 1)
	c.requestMu.Lock()
	c.requests[id] = ch
	c.requestMu.Unlock()

	defer func() {
		c.requestMu.Lock()
		delete(c.requests, id)
		c.requestMu.Unlock()
	}()

	// Send request
	if err := c.sendRequest(&request); err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	// Wait for response
	select {
	case response := <-ch:
		if response.Error != nil {
			return fmt.Errorf("MCP error %d: %s", response.Error.Code, response.Error.Message)
		}

		if result != nil && response.Result != nil {
			if err := json.Unmarshal(response.Result, result); err != nil {
				return fmt.Errorf("failed to unmarshal result: %w", err)
			}
		}

		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Notify sends a notification (no response expected)
func (c *Client) Notify(ctx context.Context, method string, params interface{}) error {
	request := MCPRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	return c.sendRequest(&request)
}

// sendRequest sends a request to the MCP server
func (c *Client) sendRequest(request *MCPRequest) error {
	data, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Add newline for line-based protocol
	data = append(data, '\n')

	_, err = c.stdin.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write request: %w", err)
	}

	return nil
}

// CallTool executes a tool with the given arguments
func (c *Client) CallTool(ctx context.Context, toolName string, arguments map[string]interface{}) (*ToolResult, error) {
	c.mu.RLock()
	initialized := c.initialized
	c.mu.RUnlock()

	if !initialized {
		return nil, fmt.Errorf("MCP client not initialized")
	}

	params := map[string]interface{}{
		"name":      toolName,
		"arguments": arguments,
	}

	var result ToolResult
	if err := c.Call(ctx, "tools/call", params, &result); err != nil {
		return nil, fmt.Errorf("tool call failed: %w", err)
	}

	return &result, nil
}

// ReadResource reads a resource from the server
func (c *Client) ReadResource(ctx context.Context, uri string) (*ResourceContent, error) {
	c.mu.RLock()
	initialized := c.initialized
	c.mu.RUnlock()

	if !initialized {
		return nil, fmt.Errorf("MCP client not initialized")
	}

	params := map[string]interface{}{
		"uri": uri,
	}

	var result struct {
		Contents []ResourceContent `json:"contents"`
	}

	if err := c.Call(ctx, "resources/read", params, &result); err != nil {
		return nil, fmt.Errorf("resource read failed: %w", err)
	}

	if len(result.Contents) == 0 {
		return nil, fmt.Errorf("no content returned for resource: %s", uri)
	}

	return &result.Contents[0], nil
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// ToolContent represents content returned by a tool
type ToolContent struct {
	Type     string      `json:"type"`
	Text     string      `json:"text,omitempty"`
	Data     interface{} `json:"data,omitempty"`
	MimeType string      `json:"mimeType,omitempty"`
}

// ResourceContent represents the content of a resource
type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
	Blob     []byte `json:"blob,omitempty"`
}

// IsInitialized returns whether the client is initialized
func (c *Client) IsInitialized() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.initialized
}

// GetName returns the client name
func (c *Client) GetName() string {
	return c.name
}

// GetCapabilities returns the server capabilities
func (c *Client) GetCapabilities() ServerCapabilities {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.capabilities
}

// RefreshTools refreshes the list of available tools
func (c *Client) RefreshTools(ctx context.Context) error {
	return c.loadTools(ctx)
}

// RefreshResources refreshes the list of available resources
func (c *Client) RefreshResources(ctx context.Context) error {
	return c.loadResources(ctx)
}

// GetAllTools returns tools from all initialized clients
func (m *Manager) GetAllTools() map[string][]Tool {
	m.clientsMu.RLock()
	defer m.clientsMu.RUnlock()

	result := make(map[string][]Tool)
	for name, client := range m.clients {
		if client.IsInitialized() {
			result[name] = client.GetTools()
		}
	}
	return result
}

// GetAllResources returns resources from all initialized clients
func (m *Manager) GetAllResources() map[string][]Resource {
	m.clientsMu.RLock()
	defer m.clientsMu.RUnlock()

	result := make(map[string][]Resource)
	for name, client := range m.clients {
		if client.IsInitialized() {
			result[name] = client.GetResources()
		}
	}
	return result
}

// CallToolByName calls a tool by name across all clients
func (m *Manager) CallToolByName(ctx context.Context, toolName string, arguments map[string]interface{}) (*ToolResult, error) {
	m.clientsMu.RLock()
	defer m.clientsMu.RUnlock()

	// Find the client that has this tool
	for _, client := range m.clients {
		if !client.IsInitialized() {
			continue
		}

		tools := client.GetTools()
		for _, tool := range tools {
			if tool.Name == toolName {
				return client.CallTool(ctx, toolName, arguments)
			}
		}
	}

	return nil, fmt.Errorf("tool not found: %s", toolName)
}

// Close closes all MCP clients
func (m *Manager) Close() error {
	m.clientsMu.Lock()
	defer m.clientsMu.Unlock()

	for _, client := range m.clients {
		client.Close()
	}

	m.clients = make(map[string]*Client)
	return nil
}
