package lobby

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	"github.com/tudor/carcasstwo/internal/game"
)

type Room struct {
	Code       string
	Engine     *game.Engine
	Created    time.Time
	botRunning bool
	botMu      sync.Mutex
}

func (r *Room) TryStartBots() bool {
	r.botMu.Lock()
	defer r.botMu.Unlock()
	if r.botRunning {
		return false
	}
	r.botRunning = true
	return true
}

func (r *Room) StopBots() {
	r.botMu.Lock()
	r.botRunning = false
	r.botMu.Unlock()
}

type Manager struct {
	mu    sync.RWMutex
	rooms map[string]*Room
}

func NewManager() *Manager {
	return &Manager{rooms: make(map[string]*Room)}
}

func (m *Manager) Create(settings game.Settings) *Room {
	if settings.MaxPlayers < 2 {
		settings.MaxPlayers = 2
	}
	if settings.MaxPlayers > 5 {
		settings.MaxPlayers = 5
	}
	if settings.Name == "" {
		settings.Name = "Carcassonne"
	}
	code := genCode()
	e := game.NewEngine(code, settings, time.Now().UnixNano())
	r := &Room{Code: code, Engine: e, Created: time.Now()}
	m.mu.Lock()
	m.rooms[code] = r
	m.mu.Unlock()
	return r
}

func (m *Manager) Get(code string) (*Room, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.rooms[code]
	return r, ok
}

func (m *Manager) ListPublic() []*Room {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := []*Room{}
	for _, r := range m.rooms {
		st := r.Engine.Public("")
		if !st.Settings.Private && st.Phase == game.PhaseLobby {
			out = append(out, r)
		}
	}
	return out
}

func genCode() string {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 5)
	rand.Read(b)
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b)
}

func (m *Manager) Ensure(code string) (*Room, error) {
	if r, ok := m.Get(code); ok {
		return r, nil
	}
	return nil, fmt.Errorf("room not found")
}
