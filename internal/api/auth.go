package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// LocalhostAuth provides secure authentication for localhost-only connections
type LocalhostAuth struct {
	sessions map[string]*Session
	tokens   map[string]*Token
	mu       sync.RWMutex
}

// Session represents an authenticated session
type Session struct {
	ID        string    `json:"id"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
}

// Token represents an API token
type Token struct {
	Hash      string    `json:"hash"`
	SessionID string    `json:"session_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// NewLocalhostAuth creates a new localhost authentication system
func NewLocalhostAuth() *LocalhostAuth {
	auth := &LocalhostAuth{
		sessions: make(map[string]*Session),
		tokens:   make(map[string]*Token),
	}
	
	// Start cleanup goroutine
	go auth.cleanupExpiredSessions()
	
	return auth
}

// isLocalhost checks if the request is from localhost
func (auth *LocalhostAuth) isLocalhost(r *http.Request) bool {
	// Get the real IP address
	ip := auth.getRealIP(r)
	
	// Parse the IP
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	
	// Check if it's localhost/loopback
	return parsedIP.IsLoopback() || ip == "127.0.0.1" || ip == "::1"
}

// getRealIP gets the real IP address from the request
func (auth *LocalhostAuth) getRealIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the chain
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	
	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	
	return ip
}

// generateSecureToken generates a cryptographically secure token
func (auth *LocalhostAuth) generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// hashToken creates a SHA-256 hash of the token
func (auth *LocalhostAuth) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// CreateSession creates a new authenticated session for localhost
func (auth *LocalhostAuth) CreateSession(r *http.Request) (*Session, error) {
	// Verify this is a localhost request
	if !auth.isLocalhost(r) {
		return nil, fmt.Errorf("authentication only available for localhost connections")
	}
	
	// Generate session ID and token
	sessionID, err := auth.generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}
	
	token, err := auth.generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}
	
	// Create session
	session := &Session{
		ID:        sessionID,
		Token:     token,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hour sessions
		IPAddress: auth.getRealIP(r),
		UserAgent: r.UserAgent(),
	}
	
	// Create token hash
	tokenHash := auth.hashToken(token)
	tokenEntry := &Token{
		Hash:      tokenHash,
		SessionID: sessionID,
		CreatedAt: time.Now(),
		ExpiresAt: session.ExpiresAt,
	}
	
	// Store session and token
	auth.mu.Lock()
	auth.sessions[sessionID] = session
	auth.tokens[tokenHash] = tokenEntry
	auth.mu.Unlock()
	
	return session, nil
}

// ValidateToken validates an API token and returns the session
func (auth *LocalhostAuth) ValidateToken(tokenString string) (*Session, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("token required")
	}
	
	// Hash the provided token
	tokenHash := auth.hashToken(tokenString)
	
	auth.mu.RLock()
	token, exists := auth.tokens[tokenHash]
	auth.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("invalid token")
	}
	
	// Check if token is expired
	if time.Now().After(token.ExpiresAt) {
		// Clean up expired token
		auth.mu.Lock()
		delete(auth.tokens, tokenHash)
		delete(auth.sessions, token.SessionID)
		auth.mu.Unlock()
		return nil, fmt.Errorf("token expired")
	}
	
	// Get session
	auth.mu.RLock()
	session, exists := auth.sessions[token.SessionID]
	auth.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("session not found")
	}
	
	return session, nil
}

// AuthMiddleware provides authentication middleware for localhost
func (auth *LocalhostAuth) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health check and auth endpoints
		if strings.HasSuffix(r.URL.Path, "/health") || 
		   strings.HasSuffix(r.URL.Path, "/auth") ||
		   strings.HasSuffix(r.URL.Path, "/login") {
			next.ServeHTTP(w, r)
			return
		}
		
		// Verify localhost
		if !auth.isLocalhost(r) {
			http.Error(w, "Access denied: localhost only", http.StatusForbidden)
			return
		}
		
		// Get token from header or query parameter
		token := r.Header.Get("Authorization")
		if token != "" && strings.HasPrefix(token, "Bearer ") {
			token = strings.TrimPrefix(token, "Bearer ")
		} else {
			// Fallback to query parameter for WebSocket connections
			token = r.URL.Query().Get("token")
		}
		
		// Validate token
		session, err := auth.ValidateToken(token)
		if err != nil {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}
		
		// Add session to request context
		r = r.WithContext(WithSession(r.Context(), session))
		
		next.ServeHTTP(w, r)
	})
}

// cleanupExpiredSessions periodically removes expired sessions
func (auth *LocalhostAuth) cleanupExpiredSessions() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	
	for range ticker.C {
		now := time.Now()
		
		auth.mu.Lock()
		// Clean up expired sessions
		for sessionID, session := range auth.sessions {
			if now.After(session.ExpiresAt) {
				delete(auth.sessions, sessionID)
			}
		}
		
		// Clean up expired tokens
		for tokenHash, token := range auth.tokens {
			if now.After(token.ExpiresAt) {
				delete(auth.tokens, tokenHash)
			}
		}
		auth.mu.Unlock()
	}
}

// GetStats returns authentication statistics
func (auth *LocalhostAuth) GetStats() map[string]interface{} {
	auth.mu.RLock()
	defer auth.mu.RUnlock()
	
	return map[string]interface{}{
		"active_sessions": len(auth.sessions),
		"active_tokens":   len(auth.tokens),
	}
}
