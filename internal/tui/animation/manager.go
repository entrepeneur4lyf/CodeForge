package animation

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/harmonica"
	"github.com/charmbracelet/log"
)

// AnimationType represents different types of animations
type AnimationType int

const (
	AnimationTabSwitch AnimationType = iota
	AnimationDialogFade
	AnimationFileTreeExpand
	AnimationProviderSelection
	AnimationChatMessage
	AnimationProgressBar
)

// Animation represents a single animation instance
type Animation struct {
	ID       string
	Type     AnimationType
	Spring   harmonica.Spring
	Target   float64
	Current  float64
	Velocity float64
	Duration time.Duration
	Started  time.Time
	Active   bool
}

// AnimationTickMsg is sent on each animation frame
type AnimationTickMsg struct {
	Time time.Time
}

// AnimationCompleteMsg is sent when an animation completes
type AnimationCompleteMsg struct {
	ID   string
	Type AnimationType
}

// Manager manages all animations in the application
type Manager struct {
	animations map[string]*Animation
	ticker     *time.Ticker
	running    bool
}

// NewManager creates a new animation manager
func NewManager() *Manager {
	return &Manager{
		animations: make(map[string]*Animation),
		running:    false,
	}
}

// StartAnimation starts a new animation
func (m *Manager) StartAnimation(id string, animType AnimationType, target float64, damping, frequency float64) {
	animation := &Animation{
		ID:       id,
		Type:     animType,
		Spring:   harmonica.NewSpring(harmonica.FPS(60), frequency, damping),
		Target:   target,
		Current:  0.0,
		Velocity: 0.0,
		Started:  time.Now(),
		Active:   true,
	}

	// Set spring parameters based on animation type
	switch animType {
	case AnimationTabSwitch:
		animation.Duration = 300 * time.Millisecond
	case AnimationDialogFade:
		animation.Duration = 200 * time.Millisecond
	case AnimationFileTreeExpand:
		animation.Duration = 250 * time.Millisecond
	case AnimationProviderSelection:
		animation.Duration = 150 * time.Millisecond
	case AnimationChatMessage:
		animation.Duration = 400 * time.Millisecond
	case AnimationProgressBar:
		animation.Duration = 500 * time.Millisecond
	default:
		animation.Duration = 300 * time.Millisecond
	}

	m.animations[id] = animation

	if !m.running {
		m.startTicker()
	}

	log.Debug("Animation started", "id", id, "type", animType, "target", target)
}

// UpdateAnimation updates a specific animation
func (m *Manager) UpdateAnimation(id string, target float64) {
	if animation, exists := m.animations[id]; exists {
		animation.Target = target
		animation.Active = true
		log.Debug("Animation updated", "id", id, "target", target)
	}
}

// StopAnimation stops a specific animation
func (m *Manager) StopAnimation(id string) {
	if animation, exists := m.animations[id]; exists {
		animation.Active = false
		delete(m.animations, id)
		log.Debug("Animation stopped", "id", id)
	}

	if len(m.animations) == 0 && m.running {
		m.stopTicker()
	}
}

// GetAnimationValue returns the current value of an animation
func (m *Manager) GetAnimationValue(id string) float64 {
	if animation, exists := m.animations[id]; exists {
		return animation.Current
	}
	return 0.0
}

// IsAnimating returns whether a specific animation is active
func (m *Manager) IsAnimating(id string) bool {
	if animation, exists := m.animations[id]; exists {
		return animation.Active
	}
	return false
}

// Update processes animation tick messages
func (m *Manager) Update(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case AnimationTickMsg:
		return m.updateAnimations()
	}
	return nil
}

// updateAnimations updates all active animations
func (m *Manager) updateAnimations() tea.Cmd {
	var completedAnimations []string
	var cmds []tea.Cmd

	for id, animation := range m.animations {
		if !animation.Active {
			continue
		}

		// Update spring physics
		animation.Current, animation.Velocity = animation.Spring.Update(
			animation.Current,
			animation.Velocity,
			animation.Target,
		)

		// Check if animation is complete
		elapsed := time.Since(animation.Started)
		threshold := 0.01

		if elapsed > animation.Duration ||
			(animation.Target != 0 && abs(animation.Current-animation.Target) < threshold) ||
			(animation.Target == 0 && abs(animation.Current) < threshold) {

			animation.Current = animation.Target
			animation.Active = false
			completedAnimations = append(completedAnimations, id)

			// Send completion message
			cmds = append(cmds, func() tea.Msg {
				return AnimationCompleteMsg{
					ID:   id,
					Type: animation.Type,
				}
			})
		}
	}

	// Clean up completed animations
	for _, id := range completedAnimations {
		delete(m.animations, id)
		log.Debug("Animation completed", "id", id)
	}

	// Stop ticker if no animations are running
	if len(m.animations) == 0 && m.running {
		m.stopTicker()
	}

	// Continue ticking if animations are still running
	if len(m.animations) > 0 {
		cmds = append(cmds, m.tick())
	}

	return tea.Batch(cmds...)
}

// startTicker starts the animation ticker
func (m *Manager) startTicker() {
	m.running = true
	log.Debug("Animation ticker started")
}

// stopTicker stops the animation ticker
func (m *Manager) stopTicker() {
	m.running = false
	log.Debug("Animation ticker stopped")
}

// tick returns a command that sends an animation tick message
func (m *Manager) tick() tea.Cmd {
	return tea.Tick(time.Millisecond*16, func(t time.Time) tea.Msg {
		return AnimationTickMsg{Time: t}
	})
}

// GetActiveAnimationCount returns the number of active animations
func (m *Manager) GetActiveAnimationCount() int {
	count := 0
	for _, animation := range m.animations {
		if animation.Active {
			count++
		}
	}
	return count
}

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Preset animation configurations
var (
	// TabSwitchAnimation - smooth tab transitions
	TabSwitchAnimation = struct {
		Damping   float64
		Frequency float64
		Duration  time.Duration
	}{
		Damping:   0.8,
		Frequency: 0.15,
		Duration:  300 * time.Millisecond,
	}

	// DialogFadeAnimation - modal dialog fade in/out
	DialogFadeAnimation = struct {
		Damping   float64
		Frequency float64
		Duration  time.Duration
	}{
		Damping:   0.9,
		Frequency: 0.2,
		Duration:  200 * time.Millisecond,
	}

	// FileTreeExpandAnimation - smooth tree expansion
	FileTreeExpandAnimation = struct {
		Damping   float64
		Frequency float64
		Duration  time.Duration
	}{
		Damping:   0.7,
		Frequency: 0.12,
		Duration:  250 * time.Millisecond,
	}

	// ProviderSelectionAnimation - provider table selection
	ProviderSelectionAnimation = struct {
		Damping   float64
		Frequency float64
		Duration  time.Duration
	}{
		Damping:   0.85,
		Frequency: 0.25,
		Duration:  150 * time.Millisecond,
	}

	// ChatMessageAnimation - message slide-in
	ChatMessageAnimation = struct {
		Damping   float64
		Frequency float64
		Duration  time.Duration
	}{
		Damping:   0.75,
		Frequency: 0.1,
		Duration:  400 * time.Millisecond,
	}

	// ProgressBarAnimation - smooth progress updates
	ProgressBarAnimation = struct {
		Damping   float64
		Frequency float64
		Duration  time.Duration
	}{
		Damping:   0.9,
		Frequency: 0.08,
		Duration:  500 * time.Millisecond,
	}
)
