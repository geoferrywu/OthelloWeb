package game

import (
	"sync"

	"github.com/google/uuid"
)

// GameMode is the type of game session.
type GameMode string

const (
	ModePVE       GameMode = "PVE"
	ModePVP       GameMode = "PVP"
	ModePVPOnline GameMode = "PVP_ONLINE"
)

type AISettings struct {
	Algorithm AIAlgorithmName `json:"algorithm"`
	Level     AILevel         `json:"level"`
}

type HintSettings struct {
	Algorithm AIAlgorithmName `json:"algorithm"`
	Level     AILevel         `json:"level"`
}

// Session represents a single game session (PvE or PvP).
type Session struct {
	ID       string
	Mode     GameMode
	PairCode string
	State    *GameState
	Players  map[Player]string // player color -> player ID
	AI       *AI
	Ready    bool // PvP/PvPOnline: both players connected
	Mutex    sync.Mutex

	AISettings   AISettings
	HintSettings map[Player]HintSettings
	LastHint     map[Player]*Position
}

// Manager handles session creation and lookup.
type Manager struct {
	sessions        map[string]*Session
	pvpOnlineByCode map[string]string
	mu              sync.RWMutex
}

// NewManager creates a new session manager.
func NewManager() *Manager {
	return &Manager{
		sessions:        make(map[string]*Session),
		pvpOnlineByCode: make(map[string]string),
	}
}

// CreateSession creates a new game session.
func (m *Manager) CreateSession(mode GameMode, color Player, size int, aiAlgorithm AIAlgorithmName, aiLevel AILevel) *Session {
	id := uuid.New().String()
	gs := NewGameState(size)

	aiColor := color.Opponent()
	ai := NewAI(size, aiColor, aiAlgorithm, aiLevel)

	s := &Session{
		ID:       id,
		Mode:     mode,
		PairCode: "",
		State:    gs,
		Players:  make(map[Player]string),
		AI:       ai,
		Ready:    mode == ModePVE,
		AISettings: AISettings{
			Algorithm: aiAlgorithm,
			Level:     aiLevel,
		},
		HintSettings: make(map[Player]HintSettings),
		LastHint:     make(map[Player]*Position),
	}

	s.Players[color] = ""
	if mode == ModePVE {
		s.Players[aiColor] = "AI"
	}
	s.HintSettings[color] = HintSettings{Algorithm: aiAlgorithm, Level: aiLevel}
	s.HintSettings[aiColor] = HintSettings{Algorithm: aiAlgorithm, Level: aiLevel}

	m.mu.Lock()
	m.sessions[id] = s
	m.mu.Unlock()
	return s
}

// JoinPvpSession joins an existing PvP session or creates one.
func (m *Manager) JoinPvpSession(color Player, size int, existingID string, aiAlgorithm AIAlgorithmName, aiLevel AILevel) *Session {
	if existingID != "" {
		m.mu.RLock()
		s, ok := m.sessions[existingID]
		m.mu.RUnlock()
		if ok && s.Mode == ModePVP && !s.Ready {
			s.Players[color] = ""
			s.Ready = true
			if _, exists := s.HintSettings[color]; !exists {
				s.HintSettings[color] = HintSettings{Algorithm: aiAlgorithm, Level: aiLevel}
			}
			return s
		}
	}

	s := m.CreateSession(ModePVP, color, size, aiAlgorithm, aiLevel)
	s.Players[color] = ""
	s.Ready = false
	return s
}

type OnlineJoinResult struct {
	Session       *Session
	IsHost        bool
	AssignedColor Player
	Reject        string
}

func (m *Manager) JoinPvpOnlineSession(pairCode string, color Player, size int, aiAlgorithm AIAlgorithmName, aiLevel AILevel) OnlineJoinResult {
	m.mu.Lock()
	defer m.mu.Unlock()

	if sessionID, ok := m.pvpOnlineByCode[pairCode]; ok {
		s, exists := m.sessions[sessionID]
		if !exists || s.Mode != ModePVPOnline {
			delete(m.pvpOnlineByCode, pairCode)
		} else {
			_, blackTaken := s.Players[BLACK]
			_, whiteTaken := s.Players[WHITE]
			if blackTaken && whiteTaken {
				return OnlineJoinResult{Reject: "该配对码已满，请重新输入配对码"}
			}

			guestColor := color
			if _, exists := s.Players[guestColor]; exists {
				guestColor = guestColor.Opponent()
			}
			if _, exists := s.Players[guestColor]; exists {
				return OnlineJoinResult{Reject: "该配对码已满，请重新输入配对码"}
			}

			s.Players[guestColor] = ""
			s.Ready = true
			if _, exists := s.HintSettings[guestColor]; !exists {
				s.HintSettings[guestColor] = HintSettings{Algorithm: aiAlgorithm, Level: aiLevel}
			}
			return OnlineJoinResult{Session: s, IsHost: false, AssignedColor: guestColor}
		}
	}

	id := uuid.New().String()
	gs := NewGameState(size)
	s := &Session{
		ID:       id,
		Mode:     ModePVPOnline,
		PairCode: pairCode,
		State:    gs,
		Players:  make(map[Player]string),
		AI:       nil,
		Ready:    false,
		AISettings: AISettings{
			Algorithm: aiAlgorithm,
			Level:     aiLevel,
		},
		HintSettings: make(map[Player]HintSettings),
		LastHint:     make(map[Player]*Position),
	}
	s.Players[color] = ""
	s.HintSettings[color] = HintSettings{Algorithm: aiAlgorithm, Level: aiLevel}
	s.HintSettings[color.Opponent()] = HintSettings{Algorithm: aiAlgorithm, Level: aiLevel}

	m.sessions[id] = s
	m.pvpOnlineByCode[pairCode] = id
	return OnlineJoinResult{Session: s, IsHost: true, AssignedColor: color}
}

func (m *Manager) GetSession(id string) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[id]
	return s, ok
}

func (m *Manager) DeleteSession(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for code, sessionID := range m.pvpOnlineByCode {
		if sessionID == id {
			delete(m.pvpOnlineByCode, code)
		}
	}
	delete(m.sessions, id)
}

func (m *Manager) InvalidateOnlineCodeBySessionID(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for code, sessionID := range m.pvpOnlineByCode {
		if sessionID == id {
			delete(m.pvpOnlineByCode, code)
		}
	}
}
