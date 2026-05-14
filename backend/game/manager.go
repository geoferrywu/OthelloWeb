package game

import (
	"sync"

	"github.com/google/uuid"
)

// GameMode is the type of game session.
type GameMode string

const (
	ModePVE GameMode = "PVE"
	ModePVP GameMode = "PVP"
)

// Session represents a single game session (PvE or PvP).
type Session struct {
	ID      string
	Mode    GameMode
	State   *GameState
	Players map[Player]string // player color -> player ID
	AI      *AI
	Ready   bool // PvP: both players connected
	Mutex   sync.Mutex
}

// Manager handles session creation and lookup.
type Manager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewManager creates a new session manager.
func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]*Session),
	}
}

// CreateSession creates a new game session.
func (m *Manager) CreateSession(mode GameMode, color Player, size int) *Session {
	id := uuid.New().String()
	gs := NewGameState(size)

	aiColor := color.Opponent()
	ai := NewAI(size, aiColor, AIDepth(size))

	s := &Session{
		ID:      id,
		Mode:    mode,
		State:   gs,
		Players: make(map[Player]string),
		AI:      ai,
		Ready:   mode == ModePVE, // PvE is ready immediately
	}

	s.Players[color] = "" // placeholder until WS connects
	if mode == ModePVE {
		s.Players[aiColor] = "AI"
	}

	m.mu.Lock()
	m.sessions[id] = s
	m.mu.Unlock()
	return s
}

// JoinPvpSession joins an existing PvP session or creates one.
func (m *Manager) JoinPvpSession(color Player, size int, existingID string) *Session {
	if existingID != "" {
		m.mu.RLock()
		s, ok := m.sessions[existingID]
		m.mu.RUnlock()
		if ok && s.Mode == ModePVP && !s.Ready {
			s.Players[color] = ""
			s.Ready = true
			return s
		}
	}

	// Create a new session waiting for second player
	s := m.CreateSession(ModePVP, color, size)
	s.Players[color] = ""
	s.Ready = false
	return s
}

// GetSession retrieves a session by ID.
func (m *Manager) GetSession(id string) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[id]
	return s, ok
}

// DeleteSession removes a session.
func (m *Manager) DeleteSession(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, id)
}
