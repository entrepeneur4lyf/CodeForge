package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// ChatSession represents a chat session
type ChatSession struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Model     string    `json:"model"`
	Provider  string    `json:"provider"`
}

// ChatMessage represents a chat message
type ChatMessage struct {
	ID        string                 `json:"id"`
	SessionID string                 `json:"session_id"`
	Role      string                 `json:"role"` // "user", "assistant", "system"
	Content   string                 `json:"content"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ChatRequest represents a new chat message request
type ChatRequest struct {
	Message  string                 `json:"message"`
	Model    string                 `json:"model,omitempty"`
	Provider string                 `json:"provider,omitempty"`
	Context  map[string]interface{} `json:"context,omitempty"`
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error,omitempty"`
	EventID string      `json:"event_id,omitempty"`
}

// ChatStorage manages chat sessions and messages in memory
type ChatStorage struct {
	sessions map[string]*ChatSession
	messages map[string][]ChatMessage
	mu       sync.RWMutex
}

// NewChatStorage creates a new chat storage instance
func NewChatStorage() *ChatStorage {
	return &ChatStorage{
		sessions: make(map[string]*ChatSession),
		messages: make(map[string][]ChatMessage),
	}
}

// CreateSession creates a new chat session
func (cs *ChatStorage) CreateSession(title string) *ChatSession {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	sessionID := fmt.Sprintf("session-%d", time.Now().Unix())
	session := &ChatSession{
		ID:        sessionID,
		Title:     title,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Status:    "active",
	}

	cs.sessions[sessionID] = session
	cs.messages[sessionID] = []ChatMessage{}
	return session
}

// GetSession retrieves a session by ID
func (cs *ChatStorage) GetSession(sessionID string) (*ChatSession, bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	session, exists := cs.sessions[sessionID]
	return session, exists
}

// GetAllSessions returns all sessions
func (cs *ChatStorage) GetAllSessions() []*ChatSession {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	sessions := make([]*ChatSession, 0, len(cs.sessions))
	for _, session := range cs.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// DeleteSession removes a session and its messages
func (cs *ChatStorage) DeleteSession(sessionID string) bool {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if _, exists := cs.sessions[sessionID]; !exists {
		return false
	}

	delete(cs.sessions, sessionID)
	delete(cs.messages, sessionID)
	return true
}

// AddMessage adds a message to a session
func (cs *ChatStorage) AddMessage(sessionID string, message ChatMessage) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if _, exists := cs.sessions[sessionID]; !exists {
		return fmt.Errorf("session not found")
	}

	message.ID = fmt.Sprintf("msg-%d", time.Now().UnixNano())
	message.Timestamp = time.Now()

	cs.messages[sessionID] = append(cs.messages[sessionID], message)

	// Update session timestamp
	cs.sessions[sessionID].UpdatedAt = time.Now()

	return nil
}

// GetMessages retrieves all messages for a session
func (cs *ChatStorage) GetMessages(sessionID string) ([]ChatMessage, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	if _, exists := cs.sessions[sessionID]; !exists {
		return nil, fmt.Errorf("session not found")
	}

	messages := cs.messages[sessionID]
	result := make([]ChatMessage, len(messages))
	copy(result, messages)
	return result, nil
}

// handleChatSessions handles GET /chat/sessions and POST /chat/sessions
func (s *Server) handleChatSessions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.getChatSessions(w, r)
	case "POST":
		s.createChatSession(w, r)
	}
}

// getChatSessions returns all chat sessions
func (s *Server) getChatSessions(w http.ResponseWriter, r *http.Request) {
	sessions := s.chatStorage.GetAllSessions()

	s.writeJSON(w, map[string]interface{}{
		"sessions": sessions,
		"total":    len(sessions),
	})
}

// createChatSession creates a new chat session
func (s *Server) createChatSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title    string `json:"title"`
		Model    string `json:"model,omitempty"`
		Provider string `json:"provider,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		req.Title = "New Chat Session"
	}

	session := s.chatStorage.CreateSession(req.Title)
	session.Model = req.Model
	session.Provider = req.Provider

	w.WriteHeader(http.StatusCreated)
	s.writeJSON(w, session)
}

// handleChatSession handles GET /chat/sessions/{id} and DELETE /chat/sessions/{id}
func (s *Server) handleChatSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	switch r.Method {
	case "GET":
		s.getChatSession(w, r, sessionID)
	case "DELETE":
		s.deleteChatSession(w, r, sessionID)
	}
}

// getChatSession returns a specific chat session
func (s *Server) getChatSession(w http.ResponseWriter, r *http.Request, sessionID string) {
	session, exists := s.chatStorage.GetSession(sessionID)
	if !exists {
		s.writeError(w, "Session not found", http.StatusNotFound)
		return
	}

	s.writeJSON(w, session)
}

// deleteChatSession deletes a chat session
func (s *Server) deleteChatSession(w http.ResponseWriter, r *http.Request, sessionID string) {
	// TODO: Implement actual session deletion
	w.WriteHeader(http.StatusNoContent)
}

// handleChatMessages handles GET /chat/sessions/{id}/messages and POST /chat/sessions/{id}/messages
func (s *Server) handleChatMessages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	switch r.Method {
	case "GET":
		s.getChatMessages(w, r, sessionID)
	case "POST":
		s.sendChatMessage(w, r, sessionID)
	}
}

// getChatMessages returns messages for a session
func (s *Server) getChatMessages(w http.ResponseWriter, r *http.Request, sessionID string) {
	// TODO: Implement actual message retrieval
	messages := []ChatMessage{
		{
			ID:        "msg-1",
			SessionID: sessionID,
			Role:      "user",
			Content:   "Can you help me review this Go code?",
			Timestamp: time.Now().Add(-30 * time.Minute),
		},
		{
			ID:        "msg-2",
			SessionID: sessionID,
			Role:      "assistant",
			Content:   "I'd be happy to help you review your Go code. Please share the code you'd like me to look at.",
			Timestamp: time.Now().Add(-29 * time.Minute),
		},
	}

	s.writeJSON(w, map[string]interface{}{
		"messages": messages,
		"total":    len(messages),
	})
}

// sendChatMessage sends a new message in a session
func (s *Server) sendChatMessage(w http.ResponseWriter, r *http.Request, sessionID string) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create user message
	userMessage := ChatMessage{
		ID:        generateMessageID(),
		SessionID: sessionID,
		Role:      "user",
		Content:   req.Message,
		Timestamp: time.Now(),
		Metadata:  req.Context,
	}

	// TODO: Store user message and generate AI response

	// For now, return the user message
	s.writeJSON(w, userMessage)
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	return "session-" + time.Now().Format("20060102-150405")
}

// generateMessageID generates a unique message ID
func generateMessageID() string {
	return "msg-" + time.Now().Format("20060102-150405-000")
}
