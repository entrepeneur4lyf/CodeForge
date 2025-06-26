package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/entrepeneur4lyf/codeforge/internal/config"
	"github.com/entrepeneur4lyf/codeforge/internal/embeddings"
	"github.com/entrepeneur4lyf/codeforge/internal/lsp"
	"github.com/entrepeneur4lyf/codeforge/internal/mcp"
	"github.com/entrepeneur4lyf/codeforge/internal/vectordb"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Server represents the web server
type Server struct {
	router   *mux.Router
	upgrader websocket.Upgrader
	config   *config.Config
	clients  map[*websocket.Conn]bool
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query    string `json:"query"`
	Language string `json:"language,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

// LSPRequest represents an LSP operation request
type LSPRequest struct {
	Operation string                 `json:"operation"`
	FilePath  string                 `json:"filePath"`
	Line      int                    `json:"line,omitempty"`
	Character int                    `json:"character,omitempty"`
	Query     string                 `json:"query,omitempty"`
	Args      map[string]interface{} `json:"args,omitempty"`
}

// MCPRequest represents an MCP operation request
type MCPRequest struct {
	Operation string                 `json:"operation"`
	ToolName  string                 `json:"toolName,omitempty"`
	Server    string                 `json:"server,omitempty"`
	URI       string                 `json:"uri,omitempty"`
	Args      map[string]interface{} `json:"args,omitempty"`
}

// ChatRequest represents an AI chat request
type ChatRequest struct {
	Message string `json:"message"`
	Model   string `json:"model,omitempty"`
}

// FileRequest represents a file operation request
type FileRequest struct {
	Operation string `json:"operation"`
	Path      string `json:"path"`
	Content   string `json:"content,omitempty"`
}

// BuildRequest represents a build operation request
type BuildRequest struct {
	Language string `json:"language"`
	Command  string `json:"command,omitempty"`
}

// NewServer creates a new web server
func NewServer(cfg *config.Config) *Server {
	router := mux.NewRouter()

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for development
		},
	}

	server := &Server{
		router:   router,
		upgrader: upgrader,
		config:   cfg,
		clients:  make(map[*websocket.Conn]bool),
	}

	server.setupRoutes()
	return server
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// Static files
	s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	// API routes
	api := s.router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/chat", s.handleChat).Methods("POST")
	api.HandleFunc("/files", s.handleFiles).Methods("GET", "POST")
	api.HandleFunc("/build", s.handleBuild).Methods("POST")
	api.HandleFunc("/lsp", s.handleLSP).Methods("POST")
	api.HandleFunc("/mcp", s.handleMCP).Methods("POST")
	api.HandleFunc("/mcp/tools", s.handleMCPTools).Methods("GET")
	api.HandleFunc("/mcp/resources", s.handleMCPResources).Methods("GET")
	api.HandleFunc("/status", s.handleStatus).Methods("GET")

	// WebSocket endpoint
	s.router.HandleFunc("/ws", s.handleWebSocket)

	// Main page
	s.router.HandleFunc("/", s.handleIndex).Methods("GET")
}

// handleIndex serves the main web interface (TUI-style)
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>CodeForge - AI-Powered Code Assistant</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: 'JetBrains Mono', 'Fira Code', 'Consolas', monospace;
            background: #0d1117;
            color: #c9d1d9;
            height: 100vh;
            overflow: hidden;
        }

        /* TUI-style layout */
        .main-container {
            display: grid;
            grid-template-columns: 250px 1fr 300px;
            grid-template-rows: 40px 1fr 200px;
            height: 100vh;
            gap: 1px;
            background: #21262d;
        }

        /* Header bar */
        .header-bar {
            grid-column: 1 / -1;
            background: #161b22;
            display: flex;
            align-items: center;
            padding: 0 16px;
            border-bottom: 1px solid #30363d;
        }

        .header-bar h1 {
            color: #58a6ff;
            font-size: 16px;
            font-weight: 600;
        }

        .header-status {
            margin-left: auto;
            display: flex;
            gap: 12px;
            font-size: 12px;
        }

        .status-indicator {
            display: flex;
            align-items: center;
            gap: 4px;
        }

        .status-dot {
            width: 8px;
            height: 8px;
            border-radius: 50%;
            background: #f85149;
        }

        .status-dot.active { background: #3fb950; }
        .status-dot.warning { background: #d29922; }

        /* File browser pane */
        .file-browser {
            background: #0d1117;
            border-right: 1px solid #30363d;
            overflow-y: auto;
        }

        .pane-header {
            background: #161b22;
            padding: 8px 12px;
            border-bottom: 1px solid #30363d;
            font-size: 12px;
            font-weight: 600;
            color: #7d8590;
        }

        .file-tree {
            padding: 8px;
        }

        .file-item {
            padding: 4px 8px;
            cursor: pointer;
            border-radius: 4px;
            font-size: 13px;
            display: flex;
            align-items: center;
            gap: 6px;
        }

        .file-item:hover {
            background: #21262d;
        }

        .file-item.selected {
            background: #1f6feb;
            color: white;
        }

        .file-icon {
            width: 16px;
            text-align: center;
        }

        /* Code editor pane */
        .code-editor {
            background: #0d1117;
            display: flex;
            flex-direction: column;
        }

        .editor-tabs {
            background: #161b22;
            border-bottom: 1px solid #30363d;
            display: flex;
            overflow-x: auto;
        }

        .editor-tab {
            padding: 8px 16px;
            border-right: 1px solid #30363d;
            cursor: pointer;
            font-size: 13px;
            white-space: nowrap;
            display: flex;
            align-items: center;
            gap: 6px;
        }

        .editor-tab.active {
            background: #0d1117;
            color: #58a6ff;
        }

        .editor-tab:hover:not(.active) {
            background: #21262d;
        }

        .tab-close {
            margin-left: 4px;
            opacity: 0.6;
            cursor: pointer;
        }

        .tab-close:hover {
            opacity: 1;
            color: #f85149;
        }

        .editor-content {
            flex: 1;
            padding: 16px;
            overflow: auto;
            font-family: 'JetBrains Mono', monospace;
            font-size: 14px;
            line-height: 1.5;
        }

        .code-textarea {
            width: 100%;
            height: 100%;
            background: transparent;
            border: none;
            color: #c9d1d9;
            font-family: inherit;
            font-size: inherit;
            line-height: inherit;
            resize: none;
            outline: none;
        }

        /* AI Chat pane */
        .ai-chat {
            background: #0d1117;
            border-left: 1px solid #30363d;
            display: flex;
            flex-direction: column;
        }

        .chat-messages {
            flex: 1;
            overflow-y: auto;
            padding: 12px;
        }

        .message {
            margin-bottom: 16px;
            padding: 8px 12px;
            border-radius: 6px;
            font-size: 13px;
            line-height: 1.4;
        }

        .message.user {
            background: #1f6feb;
            color: white;
            margin-left: 20px;
        }

        .message.ai {
            background: #21262d;
            margin-right: 20px;
        }

        .message.system {
            background: #2d1b00;
            color: #d29922;
            font-style: italic;
        }

        .chat-input-container {
            border-top: 1px solid #30363d;
            padding: 12px;
        }

        .chat-input {
            width: 100%;
            background: #21262d;
            border: 1px solid #30363d;
            border-radius: 6px;
            padding: 8px 12px;
            color: #c9d1d9;
            font-size: 13px;
            resize: none;
            outline: none;
        }

        .chat-input:focus {
            border-color: #58a6ff;
        }

        /* Output/Terminal pane */
        .output-terminal {
            grid-column: 1 / -1;
            background: #0d1117;
            border-top: 1px solid #30363d;
            display: flex;
            flex-direction: column;
        }

        .terminal-tabs {
            background: #161b22;
            border-bottom: 1px solid #30363d;
            display: flex;
        }

        .terminal-tab {
            padding: 6px 12px;
            border-right: 1px solid #30363d;
            cursor: pointer;
            font-size: 12px;
            color: #7d8590;
        }

        .terminal-tab.active {
            background: #0d1117;
            color: #c9d1d9;
        }

        .terminal-content {
            flex: 1;
            padding: 12px;
            overflow: auto;
            font-family: 'JetBrains Mono', monospace;
            font-size: 12px;
            line-height: 1.4;
        }

        .terminal-output {
            white-space: pre-wrap;
            color: #8b949e;
        }

        /* Responsive design */
        @media (max-width: 768px) {
            .main-container {
                grid-template-columns: 1fr;
                grid-template-rows: 40px 200px 1fr 150px;
            }

            .file-browser {
                border-right: none;
                border-bottom: 1px solid #30363d;
            }

            .ai-chat {
                border-left: none;
                border-top: 1px solid #30363d;
            }
        }

        /* Syntax highlighting */
        .keyword { color: #ff7b72; }
        .string { color: #a5d6ff; }
        .comment { color: #8b949e; font-style: italic; }
        .function { color: #d2a8ff; }
        .variable { color: #ffa657; }
    </style>
</head>
<body>
    <div class="main-container">
        <!-- Header Bar -->
        <div class="header-bar">
            <h1>🔧 CodeForge</h1>
            <div class="header-status">
                <div class="status-indicator">
                    <div class="status-dot active" id="embeddingDot"></div>
                    <span>Embedding</span>
                </div>
                <div class="status-indicator">
                    <div class="status-dot warning" id="lspDot"></div>
                    <span>LSP</span>
                </div>
                <div class="status-indicator">
                    <div class="status-dot warning" id="mcpDot"></div>
                    <span>MCP</span>
                </div>
            </div>
        </div>

        <!-- File Browser -->
        <div class="file-browser">
            <div class="pane-header">📁 FILES</div>
            <div class="file-tree" id="fileTree">
                <div class="file-item" onclick="openFile('README.md')">
                    <span class="file-icon">📄</span>
                    <span>README.md</span>
                </div>
                <div class="file-item" onclick="openFile('main.go')">
                    <span class="file-icon">🔧</span>
                    <span>main.go</span>
                </div>
                <div class="file-item" onclick="openFile('config.yaml')">
                    <span class="file-icon">⚙️</span>
                    <span>config.yaml</span>
                </div>
            </div>
        </div>

        <!-- Code Editor -->
        <div class="code-editor">
            <div class="editor-tabs">
                <div class="editor-tab active" id="welcomeTab">
                    <span>Welcome</span>
                    <span class="tab-close" onclick="closeTab('welcome')">×</span>
                </div>
            </div>
            <div class="editor-content">
                <textarea class="code-textarea" id="codeEditor" placeholder="// Welcome to CodeForge!
// Open a file from the file browser or start a new conversation with AI.

package main

import (
    &quot;fmt&quot;
)

func main() {
    fmt.Println(&quot;Hello, CodeForge!&quot;)
}"></textarea>
            </div>
        </div>

        <!-- AI Chat -->
        <div class="ai-chat">
            <div class="pane-header">🤖 AI ASSISTANT</div>
            <div class="chat-messages" id="chatMessages">
                <div class="message system">
                    CodeForge AI Assistant ready! Ask me about your code, request explanations, or get help with debugging.
                </div>
            </div>
            <div class="chat-input-container">
                <textarea class="chat-input" id="chatInput" placeholder="Ask me anything about your code..." rows="3"></textarea>
            </div>
        </div>

        <!-- Output/Terminal -->
        <div class="output-terminal">
            <div class="terminal-tabs">
                <div class="terminal-tab active" onclick="switchTerminalTab('output')">Output</div>
                <div class="terminal-tab" onclick="switchTerminalTab('terminal')">Terminal</div>
                <div class="terminal-tab" onclick="switchTerminalTab('problems')">Problems</div>
            </div>
            <div class="terminal-content">
                <div class="terminal-output" id="terminalOutput">
CodeForge initialized successfully.
Ready for development.

Use Ctrl+backtick to focus terminal, Ctrl+1 for files, Ctrl+2 for editor, Ctrl+3 for AI chat.
                </div>
            </div>
        </div>
    </div>

    <script>
        // Global state
        let currentFile = null;
        let openTabs = ['welcome'];
        let activeTab = 'welcome';

        // Initialize WebSocket connection
        const ws = new WebSocket('ws://localhost:8080/ws');

        ws.onopen = function() {
            addSystemMessage('Connected to CodeForge server');
        };

        ws.onmessage = function(event) {
            const data = JSON.parse(event.data);
            if (data.type === 'status') {
                updateStatus(data.data);
            }
        };

        // File operations
        function openFile(filename) {
            // Remove selection from other files
            document.querySelectorAll('.file-item').forEach(item => {
                item.classList.remove('selected');
            });

            // Select current file
            event.target.closest('.file-item').classList.add('selected');

            // Add tab if not already open
            if (!openTabs.includes(filename)) {
                openTabs.push(filename);
                addTab(filename);
            }

            // Switch to tab
            switchTab(filename);
            currentFile = filename;

            // Load file content (placeholder)
            loadFileContent(filename);
        }

        function addTab(filename) {
            const tabsContainer = document.querySelector('.editor-tabs');
            const tab = document.createElement('div');
            tab.className = 'editor-tab';
            tab.id = filename + 'Tab';
            tab.innerHTML = '<span>' + filename + '</span><span class="tab-close" onclick="closeTab(\'' + filename + '\')">×</span>';
            tab.onclick = () => switchTab(filename);
            tabsContainer.appendChild(tab);
        }

        function switchTab(filename) {
            // Update tab appearance
            document.querySelectorAll('.editor-tab').forEach(tab => {
                tab.classList.remove('active');
            });
            document.getElementById(filename + 'Tab').classList.add('active');

            activeTab = filename;
            loadFileContent(filename);
        }

        function closeTab(filename) {
            if (openTabs.length <= 1) return; // Keep at least one tab

            openTabs = openTabs.filter(tab => tab !== filename);
            document.getElementById(filename + 'Tab').remove();

            if (activeTab === filename) {
                switchTab(openTabs[openTabs.length - 1]);
            }
        }

        function loadFileContent(filename) {
            const editor = document.getElementById('codeEditor');

            // Placeholder content based on file type
            const content = {
                'README.md': '# CodeForge\\n\\nAI-powered coding assistant with multi-language support.\\n\\n## Features\\n- Multi-language build system\\n- AI integration\\n- LSP support\\n- Web interface',
                'main.go': 'package main\\n\\nimport (\\n    "fmt"\\n)\\n\\nfunc main() {\\n    fmt.Println("Hello, CodeForge!")\\n}',
                'config.yaml': '# CodeForge Configuration\\nworking_dir: "."\\n\\n# LLM Provider Configuration\\nllm:\\n  providers:\\n    - name: "anthropic"\\n      type: "anthropic"',
                'welcome': '// Welcome to CodeForge!\\n// Open a file from the file browser or start a new conversation with AI.\\n\\npackage main\\n\\nimport (\\n    "fmt"\\n)\\n\\nfunc main() {\\n    fmt.Println("Hello, CodeForge!")\\n}'
            };

            editor.value = content[filename] || '// New file: ' + filename;
        }

        // Chat functionality
        function sendMessage() {
            const input = document.getElementById('chatInput');
            const message = input.value.trim();

            if (!message) return;

            addUserMessage(message);
            input.value = '';

            // Send to AI (placeholder)
            setTimeout(() => {
                addAIMessage('I understand you want help with: "' + message + '". This is a placeholder response. The AI integration will provide real assistance here.');
            }, 1000);
        }

        function addUserMessage(message) {
            const messagesContainer = document.getElementById('chatMessages');
            const messageDiv = document.createElement('div');
            messageDiv.className = 'message user';
            messageDiv.textContent = message;
            messagesContainer.appendChild(messageDiv);
            messagesContainer.scrollTop = messagesContainer.scrollHeight;
        }

        function addAIMessage(message) {
            const messagesContainer = document.getElementById('chatMessages');
            const messageDiv = document.createElement('div');
            messageDiv.className = 'message ai';
            messageDiv.textContent = message;
            messagesContainer.appendChild(messageDiv);
            messagesContainer.scrollTop = messagesContainer.scrollHeight;
        }

        function addSystemMessage(message) {
            const messagesContainer = document.getElementById('chatMessages');
            const messageDiv = document.createElement('div');
            messageDiv.className = 'message system';
            messageDiv.textContent = message;
            messagesContainer.appendChild(messageDiv);
            messagesContainer.scrollTop = messagesContainer.scrollHeight;
        }

        // Terminal functionality
        function switchTerminalTab(tab) {
            document.querySelectorAll('.terminal-tab').forEach(t => {
                t.classList.remove('active');
            });
            event.target.classList.add('active');

            const output = document.getElementById('terminalOutput');
            const content = {
                'output': 'CodeForge initialized successfully.\\nReady for development.\\n\\nUse Ctrl+backtick to focus terminal, Ctrl+1 for files, Ctrl+2 for editor, Ctrl+3 for AI chat.',
                'terminal': '$ echo "Welcome to CodeForge terminal"\\nWelcome to CodeForge terminal\\n$ ',
                'problems': 'No problems detected.\\n\\nLSP diagnostics will appear here when available.'
            };

            output.textContent = content[tab] || '';
        }

        // Status updates
        function updateStatus(status) {
            document.getElementById('embeddingDot').className = 'status-dot ' + (status.embedding ? 'active' : '');
            document.getElementById('lspDot').className = 'status-dot ' + (status.lsp ? 'active' : 'warning');
            document.getElementById('mcpDot').className = 'status-dot ' + (status.mcp ? 'active' : 'warning');
        }

        // Keyboard shortcuts
        document.addEventListener('keydown', function(e) {
            if (e.ctrlKey || e.metaKey) {
                switch(e.key) {
                    case '1':
                        e.preventDefault();
                        document.querySelector('.file-browser').focus();
                        break;
                    case '2':
                        e.preventDefault();
                        document.getElementById('codeEditor').focus();
                        break;
                    case '3':
                        e.preventDefault();
                        document.getElementById('chatInput').focus();
                        break;
                    case '`':
                        e.preventDefault();
                        document.querySelector('.terminal-content').focus();
                        break;
                }
            }
        });

        // Chat input handling
        document.getElementById('chatInput').addEventListener('keydown', function(e) {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                sendMessage();
            }
        });

        // Load initial status
        fetch('/api/status')
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    updateStatus(data.data);
                }
            })
            .catch(error => {
                console.error('Failed to load status:', error);
            });
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// handleChat handles AI chat requests
func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get the default model
	defaultModel, err := llm.GetDefaultModel()
	if err != nil {
		s.sendError(w, fmt.Sprintf("Failed to get default model: %v", err), http.StatusInternalServerError)
		return
	}

	// Create completion request
	completionReq := llm.CompletionRequest{
		Model: defaultModel.ID,
		Messages: []llm.Message{
			{
				Role:    "user",
				Content: req.Message,
			},
		},
		MaxTokens:   defaultModel.DefaultMaxTokens,
		Temperature: 0.7,
	}

	// Get completion
	resp, err := llm.GetCompletion(ctx, completionReq)
	if err != nil {
		s.sendError(w, fmt.Sprintf("AI completion failed: %v", err), http.StatusInternalServerError)
		return
	}

	s.sendSuccess(w, map[string]interface{}{
		"message": resp.Content,
		"model":   defaultModel.ID,
	})
}

// handleFiles handles file operations
func (s *Server) handleFiles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// List files in working directory
		files, err := s.listFiles(".")
		if err != nil {
			s.sendError(w, fmt.Sprintf("Failed to list files: %v", err), http.StatusInternalServerError)
			return
		}
		s.sendSuccess(w, files)

	case "POST":
		var req FileRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.sendError(w, "Invalid request format", http.StatusBadRequest)
			return
		}

		switch req.Operation {
		case "read":
			content, err := s.readFile(req.Path)
			if err != nil {
				s.sendError(w, fmt.Sprintf("Failed to read file: %v", err), http.StatusInternalServerError)
				return
			}
			s.sendSuccess(w, map[string]string{"content": content})

		case "write":
			if err := s.writeFile(req.Path, req.Content); err != nil {
				s.sendError(w, fmt.Sprintf("Failed to write file: %v", err), http.StatusInternalServerError)
				return
			}
			s.sendSuccess(w, map[string]string{"status": "saved"})

		default:
			s.sendError(w, "Unknown file operation", http.StatusBadRequest)
		}
	}
}

// handleBuild handles build operations
func (s *Server) handleBuild(w http.ResponseWriter, r *http.Request) {
	var req BuildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// This is a placeholder - in a real implementation, you'd integrate with the build system
	result := map[string]interface{}{
		"status":   "success",
		"language": req.Language,
		"output":   fmt.Sprintf("Build completed for %s project", req.Language),
		"command":  req.Command,
	}

	s.sendSuccess(w, result)
}

// Helper functions for file operations
func (s *Server) listFiles(dir string) ([]map[string]interface{}, error) {
	// Placeholder implementation
	files := []map[string]interface{}{
		{"name": "README.md", "type": "file", "icon": "📄"},
		{"name": "main.go", "type": "file", "icon": "🔧"},
		{"name": "config.yaml", "type": "file", "icon": "⚙️"},
		{"name": "internal/", "type": "directory", "icon": "📁"},
		{"name": "cmd/", "type": "directory", "icon": "📁"},
	}
	return files, nil
}

func (s *Server) readFile(path string) (string, error) {
	// Placeholder content based on file type
	content := map[string]string{
		"README.md":    "# CodeForge\n\nAI-powered coding assistant with multi-language support.\n\n## Features\n- Multi-language build system\n- AI integration\n- LSP support\n- Web interface",
		"main.go":      "package main\n\nimport (\n    \"fmt\"\n)\n\nfunc main() {\n    fmt.Println(\"Hello, CodeForge!\")\n}",
		"config.yaml":  "# CodeForge Configuration\nworking_dir: \".\"\n\n# LLM Provider Configuration\nllm:\n  providers:\n    - name: \"anthropic\"\n      type: \"anthropic\"",
	}

	if c, exists := content[path]; exists {
		return c, nil
	}
	return "// New file: " + path, nil
}

func (s *Server) writeFile(path, content string) error {
	// Placeholder - in real implementation, write to actual file
	fmt.Printf("Writing to %s: %d bytes\n", path, len(content))
	return nil
}

// handleSearch handles semantic code search requests
func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if req.Limit == 0 {
		req.Limit = 10
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Generate embedding for query
	embedding, err := embeddings.GetEmbedding(ctx, req.Query)
	if err != nil {
		s.sendError(w, fmt.Sprintf("Failed to generate embedding: %v", err), http.StatusInternalServerError)
		return
	}

	// Search vector database
	vdb := vectordb.Get()
	if vdb == nil {
		s.sendError(w, "Vector database not available", http.StatusServiceUnavailable)
		return
	}

	results, err := vdb.SearchSimilarCode(ctx, embedding, req.Language, req.Limit)
	if err != nil {
		s.sendError(w, fmt.Sprintf("Search failed: %v", err), http.StatusInternalServerError)
		return
	}

	s.sendSuccess(w, results)
}

// handleLSP handles LSP operation requests
func (s *Server) handleLSP(w http.ResponseWriter, r *http.Request) {
	var req LSPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	manager := lsp.GetManager()
	if manager == nil {
		s.sendError(w, "LSP manager not available", http.StatusServiceUnavailable)
		return
	}

	switch req.Operation {
	case "symbols":
		client := manager.GetClientForLanguage("go") // Default to Go for demo
		if client == nil {
			s.sendError(w, "No LSP client available", http.StatusServiceUnavailable)
			return
		}

		symbols, err := client.GetWorkspaceSymbols(ctx, req.Query)
		if err != nil {
			s.sendError(w, fmt.Sprintf("Symbol search failed: %v", err), http.StatusInternalServerError)
			return
		}
		s.sendSuccess(w, symbols)

	case "definition":
		client := manager.GetClientForFile(req.FilePath)
		if client == nil {
			s.sendError(w, "No LSP client available for file", http.StatusServiceUnavailable)
			return
		}

		locations, err := client.GetDefinition(ctx, req.FilePath, req.Line, req.Character)
		if err != nil {
			s.sendError(w, fmt.Sprintf("Definition search failed: %v", err), http.StatusInternalServerError)
			return
		}
		s.sendSuccess(w, locations)

	default:
		s.sendError(w, "Unknown LSP operation", http.StatusBadRequest)
	}
}

// handleMCP handles MCP operation requests
func (s *Server) handleMCP(w http.ResponseWriter, r *http.Request) {
	var req MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	manager := mcp.GetManager()
	if manager == nil {
		s.sendError(w, "MCP manager not available", http.StatusServiceUnavailable)
		return
	}

	switch req.Operation {
	case "call":
		result, err := manager.CallToolByName(ctx, req.ToolName, req.Args)
		if err != nil {
			s.sendError(w, fmt.Sprintf("Tool call failed: %v", err), http.StatusInternalServerError)
			return
		}
		s.sendSuccess(w, result)

	case "read":
		client := manager.GetClient(req.Server)
		if client == nil {
			s.sendError(w, "MCP server not found", http.StatusNotFound)
			return
		}

		content, err := client.ReadResource(ctx, req.URI)
		if err != nil {
			s.sendError(w, fmt.Sprintf("Resource read failed: %v", err), http.StatusInternalServerError)
			return
		}
		s.sendSuccess(w, content)

	default:
		s.sendError(w, "Unknown MCP operation", http.StatusBadRequest)
	}
}

// handleMCPTools returns available MCP tools
func (s *Server) handleMCPTools(w http.ResponseWriter, r *http.Request) {
	manager := mcp.GetManager()
	if manager == nil {
		s.sendError(w, "MCP manager not available", http.StatusServiceUnavailable)
		return
	}

	tools := manager.GetAllTools()
	s.sendSuccess(w, tools)
}

// handleMCPResources returns available MCP resources
func (s *Server) handleMCPResources(w http.ResponseWriter, r *http.Request) {
	manager := mcp.GetManager()
	if manager == nil {
		s.sendError(w, "MCP manager not available", http.StatusServiceUnavailable)
		return
	}

	resources := manager.GetAllResources()
	s.sendSuccess(w, resources)
}

// handleStatus returns system status
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"embedding": embeddings.GetNative() != nil && embeddings.GetNative().IsInitialized(),
		"vectordb":  vectordb.Get() != nil,
		"lsp":       lsp.GetManager() != nil,
		"mcp":       mcp.GetManager() != nil,
		"timestamp": time.Now().Unix(),
	}

	s.sendSuccess(w, status)
}

// handleWebSocket handles WebSocket connections for real-time updates
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("WebSocket upgrade failed: %v\n", err)
		return
	}
	defer conn.Close()

	s.clients[conn] = true
	defer delete(s.clients, conn)

	// Send initial status
	status := map[string]interface{}{
		"type": "status",
		"data": map[string]bool{
			"embedding": embeddings.GetNative() != nil && embeddings.GetNative().IsInitialized(),
			"vectordb":  vectordb.Get() != nil,
			"lsp":       lsp.GetManager() != nil,
			"mcp":       mcp.GetManager() != nil,
		},
	}
	conn.WriteJSON(status)

	// Keep connection alive and handle messages
	for {
		var msg map[string]interface{}
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}

		// Echo back for now (can be extended for real-time features)
		conn.WriteJSON(map[string]interface{}{
			"type": "echo",
			"data": msg,
		})
	}
}

// sendSuccess sends a successful API response
func (s *Server) sendSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
	})
}

// sendError sends an error API response
func (s *Server) sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Error:   message,
	})
}

// Start starts the web server
func (s *Server) Start(port int) error {
	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("🌐 Web interface starting on http://localhost%s\n", addr)
	return http.ListenAndServe(addr, s.router)
}
