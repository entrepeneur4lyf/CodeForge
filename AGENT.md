# CodeForge Agent Memory

This file serves as memory for AI agents working in this CodeForge repository.

## Project Overview
- **Type**: Go application with TUI interface
- **Purpose**: AI-powered code assistant and development tool
- **Architecture**: Bubble Tea TUI with modular component system
- **Framework**: Built on Charm ecosystem (Bubble Tea, Lipgloss, Bubbles)

## Build Commands
- Build: `go build ./cmd/codeforge`
- Run: `./codeforge tui`
- Test: `go test ./...`
- Clean: `go clean`
- Format: `gofmt -w .`

## Code Style Guidelines
- Use Go standard formatting (`gofmt`)
- Follow Go naming conventions (PascalCase for exported, camelCase for unexported)
- Use structured logging: `log.Info()`, `log.Error()`, `log.Debug()`
- Import organization: standard library, third-party, local packages
- Keep functions focused and under 50 lines when possible
- Use interfaces for testability and modularity
- Handle errors explicitly, never ignore them
- Use meaningful variable and function names
- Add comments for exported functions and complex logic

## Architecture Components
- **TUI Framework**: Bubble Tea v2 with Lipgloss v2 styling
- **Chat Interface**: AI conversation with Glamour markdown rendering
- **File Management**: Bubbles filepicker with filtering
- **Dialog System**: Modal dialogs for settings, initialization, provider config
- **Theme System**: Pluggable color themes (CodeForge, Catppuccin, Dracula, etc.)
- **Animation**: Harmonica physics-based smooth animations
- **Provider System**: Multi-LLM support (OpenAI, Anthropic, Groq, Local, etc.)

## Key Dependencies
- `github.com/charmbracelet/bubbletea/v2` - TUI framework
- `github.com/charmbracelet/lipgloss/v2` - Styling and layout
- `github.com/charmbracelet/bubbles` - UI components
- `github.com/charmbracelet/glamour` - Markdown rendering
- `github.com/charmbracelet/harmonica` - Animation physics
- `github.com/charmbracelet/log` - Structured logging

## Development Patterns
- Follow OpenCode architectural patterns for consistency
- Use the provider pattern for AI model integration
- Implement proper error handling with context
- Write comprehensive tests for UI components
- Maintain responsive design principles
- Use dependency injection for testability
- Keep state management centralized in app model
- Use message passing for component communication

## File Structure
- `cmd/codeforge/` - Main application entry point
- `internal/tui/` - TUI implementation and components
- `internal/llm/` - LLM provider implementations
- `internal/models/` - Data models and types
- `internal/config/` - Configuration management

## Testing Guidelines
- Write unit tests for all business logic
- Use table-driven tests for multiple scenarios
- Mock external dependencies (LLM APIs, file system)
- Test UI components with bubble tea testing utilities
- Maintain >80% code coverage
- Use integration tests for critical user flows

## Performance Considerations
- Lazy load file trees for large repositories
- Implement virtual scrolling for large chat histories
- Use efficient rendering with Lipgloss caching
- Minimize allocations in hot paths
- Profile memory usage regularly
- Optimize startup time with async initialization
