package hub

import "sync"

type Sub struct {
	Room     string
	PlayerID string
	Notify   chan struct{}
}

type Hub struct {
	mu   sync.RWMutex
	subs map[string]map[*Sub]struct{}
}

func New() *Hub {
	return &Hub{subs: make(map[string]map[*Sub]struct{})}
}

func (h *Hub) Subscribe(room, playerID string) *Sub {
	s := &Sub{Room: room, PlayerID: playerID, Notify: make(chan struct{}, 1)}
	h.mu.Lock()
	if h.subs[room] == nil {
		h.subs[room] = make(map[*Sub]struct{})
	}
	h.subs[room][s] = struct{}{}
	h.mu.Unlock()
	return s
}

func (h *Hub) Unsubscribe(s *Sub) {
	h.mu.Lock()
	defer h.mu.Unlock()
	room := h.subs[s.Room]
	if room == nil {
		return
	}
	delete(room, s)
	if len(room) == 0 {
		delete(h.subs, s.Room)
	}
}

func (h *Hub) Broadcast(room string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for s := range h.subs[room] {
		select {
		case s.Notify <- struct{}{}:
		default:
		}
	}
}
