package handlers

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/starfederation/datastar-go/datastar"
	"github.com/tudor/carcasstwo/internal/game"
	"github.com/tudor/carcasstwo/internal/hub"
	"github.com/tudor/carcasstwo/internal/lobby"
)

type Server struct {
	Lobby *lobby.Manager
	Hub   *hub.Hub
}

func New(l *lobby.Manager, h *hub.Hub) *Server {
	return &Server{Lobby: l, Hub: h}
}

func (s *Server) Routes(r chi.Router) {
	r.Get("/", s.Home)
	r.Get("/room/{code}", s.RoomPage)
	r.Post("/api/rooms", s.CreateRoom)
	r.Post("/api/join", s.JoinByCode)
	r.Post("/api/rooms/{code}/join", s.JoinRoom)
	r.Post("/api/rooms/{code}/start", s.StartGame)
	r.Post("/api/rooms/{code}/place-tile", s.PlaceTile)
	r.Post("/api/rooms/{code}/place-meeple", s.PlaceMeeple)
	r.Post("/api/rooms/{code}/skip-meeple", s.SkipMeeple)
	r.Post("/api/rooms/{code}/recall-abbot", s.RecallAbbot)
	r.Get("/api/rooms/{code}/stream", s.Stream)
	r.Get("/api/rooms/{code}/state", s.State)
	r.Get("/api/rooms", s.ListRooms)
}

func playerCookie(w http.ResponseWriter, r *http.Request) string {
	c, err := r.Cookie("player_id")
	if err == nil && c.Value != "" {
		return c.Value
	}
	id := uuid.NewString()
	http.SetCookie(w, &http.Cookie{Name: "player_id", Value: id, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode})
	return id
}

func (s *Server) Home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, homeHTML)
}

func (s *Server) RoomPage(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(chi.URLParam(r, "code"))
	room, ok := s.Lobby.Get(code)
	if !ok {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}
	playerID := playerCookie(w, r)
	st := room.Engine.Public(playerID)
	stateJSON, err := json.Marshal(st)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	page := roomHTML
	page = strings.ReplaceAll(page, "{{CODE}}", html.EscapeString(code))
	page = strings.ReplaceAll(page, "{{NAME}}", html.EscapeString(st.Settings.Name))
	page = strings.Replace(page, "{{LOBBY}}", renderLobby(st, playerID), 1)
	page = strings.Replace(page, "{{HUD}}", renderHUD(st), 1)
	page = strings.Replace(page, "{{STATE}}", string(stateJSON), 1)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, page)
}

func (s *Server) State(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(chi.URLParam(r, "code"))
	room, ok := s.Lobby.Get(code)
	if !ok {
		http.Error(w, "not found", 404)
		return
	}
	playerID := playerCookie(w, r)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(room.Engine.Public(playerID))
}

type createBody struct {
	Name                    string `json:"name"`
	MaxPlayers              any    `json:"maxPlayers"`
	Private                 bool   `json:"private"`
	PlayerName              string `json:"playerName"`
	Farmers                 bool   `json:"farmers"`
	River                   bool   `json:"river"`
	Abbot                   bool   `json:"abbot"`
	InnsAndCathedrals       bool   `json:"innsAndCathedrals"`
	TradersAndBuilders      bool   `json:"tradersAndBuilders"`
	PrincessAndDragon       bool   `json:"princessAndDragon"`
	Tower                   bool   `json:"tower"`
	AbbeyAndMayor           bool   `json:"abbeyAndMayor"`
	BridgesCastlesBazaars   bool   `json:"bridgesCastlesBazaars"`
	HillsAndSheep           bool   `json:"hillsAndSheep"`
}

func asInt(v any, fallback int) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case string:
		var x int
		fmt.Sscanf(n, "%d", &x)
		if x > 0 {
			return x
		}
	}
	return fallback
}

func (s *Server) CreateRoom(w http.ResponseWriter, r *http.Request) {
	playerID := playerCookie(w, r)
	var body createBody
	_ = readSignalsOrJSON(r, &body)
	maxPlayers := asInt(body.MaxPlayers, 4)
	if body.PlayerName == "" {
		body.PlayerName = "Host"
	}
	exp := game.Expansions{
		Farmers:            body.Farmers,
		River:              body.River,
		Abbot:              body.Abbot,
		InnsAndCathedrals:  body.InnsAndCathedrals,
		TradersAndBuilders: body.TradersAndBuilders,
		PrincessAndDragon:  body.PrincessAndDragon,
		Tower:              body.Tower,
		AbbeyAndMayor:      body.AbbeyAndMayor,
		BridgesCastlesBaz:  body.BridgesCastlesBazaars,
		HillsAndSheep:      body.HillsAndSheep,
	}
	if maxPlayers > exp.MaxPlayersCap() {
		maxPlayers = exp.MaxPlayersCap()
	}
	room := s.Lobby.Create(game.Settings{
		MaxPlayers: maxPlayers,
		Private:    body.Private,
		Name:       body.Name,
		Expansions: exp,
	})
	_, err := room.Engine.AddPlayer(playerID, body.PlayerName)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	sse := datastar.NewSSE(w, r)
	_ = sse.Redirect("/room/" + room.Code)
}

func (s *Server) JoinByCode(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PlayerName string `json:"playerName"`
		Code       string `json:"code"`
	}
	_ = readSignalsOrJSON(r, &body)
	code := strings.ToUpper(strings.TrimSpace(body.Code))
	s.joinRoom(w, r, code, body.PlayerName)
}

func (s *Server) JoinRoom(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(chi.URLParam(r, "code"))
	var body struct {
		PlayerName string `json:"playerName"`
		Code       string `json:"code"`
	}
	_ = readSignalsOrJSON(r, &body)
	s.joinRoom(w, r, code, body.PlayerName)
}

func (s *Server) joinRoom(w http.ResponseWriter, r *http.Request, code, playerName string) {
	room, ok := s.Lobby.Get(code)
	if !ok {
		http.Error(w, "Room not found", 404)
		return
	}
	playerID := playerCookie(w, r)
	if playerName == "" {
		playerName = "Player"
	}
	if _, err := room.Engine.AddPlayer(playerID, playerName); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.Hub.Broadcast(code)
	sse := datastar.NewSSE(w, r)
	_ = sse.Redirect("/room/" + code)
}

func (s *Server) ListRooms(w http.ResponseWriter, r *http.Request) {
	rooms := s.Lobby.ListPublic()
	var b strings.Builder
	b.WriteString(`<div id="room-list" class="room-list">`)
	if len(rooms) == 0 {
		b.WriteString(`<p class="muted">No open tables — create one.</p>`)
	}
	for _, room := range rooms {
		st := room.Engine.Public("")
		b.WriteString(fmt.Sprintf(
			`<button class="room-row" data-on:click="$code='%s'; $playerName = $playerName || 'Player'; @post('/api/rooms/%s/join')">
				<span class="room-name">%s</span>
				<span class="room-meta">%s · %d/%d</span>
			</button>`,
			room.Code, room.Code,
			html.EscapeString(st.Settings.Name),
			room.Code,
			len(st.Players), st.Settings.MaxPlayers,
		))
	}
	b.WriteString(`</div>`)
	sse := datastar.NewSSE(w, r)
	_ = sse.PatchElements(b.String())
}

func (s *Server) StartGame(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(chi.URLParam(r, "code"))
	room, ok := s.Lobby.Get(code)
	if !ok {
		http.Error(w, "not found", 404)
		return
	}
	if err := room.Engine.Start(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.Hub.Broadcast(code)
	sse := datastar.NewSSE(w, r)
	_ = sse.PatchElements(renderHUD(room.Engine.Public(playerCookie(w, r))))
}

func (s *Server) PlaceTile(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(chi.URLParam(r, "code"))
	room, ok := s.Lobby.Get(code)
	if !ok {
		http.Error(w, "not found", 404)
		return
	}
	playerID := playerCookie(w, r)
	var body struct {
		X   int `json:"x"`
		Y   int `json:"y"`
		Rot int `json:"rot"`
	}
	if err := readSignalsOrJSON(r, &body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if err := room.Engine.PlaceTile(playerID, body.X, body.Y, body.Rot); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.Hub.Broadcast(code)
	sse := datastar.NewSSE(w, r)
	st := room.Engine.Public(playerID)
	_ = sse.PatchElements(renderHUD(st))
	_ = sse.DispatchCustomEvent("carcassonne:state", st)
}

func (s *Server) PlaceMeeple(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(chi.URLParam(r, "code"))
	room, ok := s.Lobby.Get(code)
	if !ok {
		http.Error(w, "not found", 404)
		return
	}
	playerID := playerCookie(w, r)
	var body struct {
		Kind  game.FeatureKind `json:"kind"`
		Slot  int              `json:"slot"`
		Abbot bool             `json:"abbot"`
	}
	if err := readSignalsOrJSON(r, &body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if err := room.Engine.PlaceMeeple(playerID, body.Kind, body.Slot, body.Abbot); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.Hub.Broadcast(code)
	sse := datastar.NewSSE(w, r)
	st := room.Engine.Public(playerID)
	_ = sse.PatchElements(renderHUD(st))
	_ = sse.DispatchCustomEvent("carcassonne:state", st)
}

func (s *Server) SkipMeeple(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(chi.URLParam(r, "code"))
	room, ok := s.Lobby.Get(code)
	if !ok {
		http.Error(w, "not found", 404)
		return
	}
	playerID := playerCookie(w, r)
	if err := room.Engine.SkipMeeple(playerID); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.Hub.Broadcast(code)
	sse := datastar.NewSSE(w, r)
	st := room.Engine.Public(playerID)
	_ = sse.PatchElements(renderHUD(st))
	_ = sse.DispatchCustomEvent("carcassonne:state", st)
}

func (s *Server) RecallAbbot(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(chi.URLParam(r, "code"))
	room, ok := s.Lobby.Get(code)
	if !ok {
		http.Error(w, "not found", 404)
		return
	}
	playerID := playerCookie(w, r)
	if err := room.Engine.RecallAbbot(playerID); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.Hub.Broadcast(code)
	sse := datastar.NewSSE(w, r)
	st := room.Engine.Public(playerID)
	_ = sse.PatchElements(renderHUD(st))
	_ = sse.DispatchCustomEvent("carcassonne:state", st)
}

func (s *Server) Stream(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(chi.URLParam(r, "code"))
	room, ok := s.Lobby.Get(code)
	if !ok {
		http.Error(w, "not found", 404)
		return
	}
	playerID := playerCookie(w, r)
	sub := s.Hub.Subscribe(code, playerID)
	defer s.Hub.Unsubscribe(sub)

	sse := datastar.NewSSE(w, r)
	push := func() {
		st := room.Engine.Public(playerID)
		_ = sse.PatchElements(renderHUD(st))
		_ = sse.PatchElements(renderLobby(st, playerID))
		_ = sse.DispatchCustomEvent("carcassonne:state", st)
	}
	push()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-sub.Notify:
			push()
		}
	}
}

func readSignalsOrJSON(r *http.Request, dst any) error {
	if err := datastar.ReadSignals(r, dst); err == nil {
		return nil
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(dst)
}

func renderLobby(st game.PublicState, youID string) string {
	if st.Phase != game.PhaseLobby {
		return `<div id="lobby-panel" class="hidden"></div>`
	}
	var b strings.Builder
	b.WriteString(`<div id="lobby-panel" class="lobby-panel">`)
	b.WriteString(fmt.Sprintf(`<h2>%s</h2>`, html.EscapeString(st.Settings.Name)))
	if st.Settings.Private {
		b.WriteString(`<p class="vis-badge">Private</p>`)
		b.WriteString(fmt.Sprintf(`<p class="code">Room <strong>%s</strong></p>`, st.RoomCode))
		b.WriteString(`<p class="muted">Share this code — the table is hidden from Open tables.</p>`)
	} else {
		b.WriteString(fmt.Sprintf(`<p class="code">Room <strong>%s</strong></p>`, st.RoomCode))
	}
	b.WriteString(`<ul class="players">`)
	for _, p := range st.Players {
		you := ""
		if p.ID == youID {
			you = " (you)"
		}
		b.WriteString(fmt.Sprintf(`<li><span class="swatch" style="--c:var(--%s)"></span>%s%s</li>`, p.Color, html.EscapeString(p.Name), you))
	}
	b.WriteString(`</ul>`)
	b.WriteString(fmt.Sprintf(`<p class="muted">%d / %d players</p>`, len(st.Players), st.Settings.MaxPlayers))
	if labels := st.Settings.Expansions.Labels(); len(labels) > 0 {
		b.WriteString(`<ul class="exp-list">`)
		for _, l := range labels {
			b.WriteString(fmt.Sprintf(`<li>%s</li>`, html.EscapeString(l)))
		}
		b.WriteString(`</ul>`)
	}
	if len(st.Players) >= 2 {
		b.WriteString(fmt.Sprintf(`<button class="btn primary" data-on:click="@post('/api/rooms/%s/start')">Start game</button>`, st.RoomCode))
	} else {
		b.WriteString(`<p class="muted">Waiting for another player…</p>`)
	}
	b.WriteString(`</div>`)
	return b.String()
}

func renderHUD(st game.PublicState) string {
	if st.Phase == game.PhaseLobby {
		return `<div id="hud"></div>`
	}
	var b strings.Builder
	b.WriteString(`<div id="hud">`)
	b.WriteString(fmt.Sprintf(`<div class="hud-msg">%s</div>`, html.EscapeString(st.Message)))
	b.WriteString(`<div class="scores">`)
	for i, p := range st.Players {
		turn := ""
		if st.Phase != game.PhaseGameOver && i == st.Turn {
			turn = " turn"
		}
		b.WriteString(fmt.Sprintf(
			`<div class="score%s" style="--c:var(--%s)"><span class="name">%s</span><span class="pts">%d</span><span class="m">♟%d</span>%s</div>`,
			turn, p.Color, html.EscapeString(p.Name), p.Score, p.Meeple,
			func() string {
				if st.Settings.Expansions.Abbot {
					return fmt.Sprintf(`<span class="m">✝%d</span>`, p.Abbot)
				}
				return ""
			}(),
		))
	}
	b.WriteString(`</div>`)
	b.WriteString(fmt.Sprintf(`<div class="meta"><span>Tiles left: %d</span>`, st.TilesLeft))
	if st.CurrentTile != nil {
		b.WriteString(`<span class="cur">Draw ready</span>`)
	}
	b.WriteString(`</div>`)
	if st.Phase == game.PhasePlaceMeeple {
		b.WriteString(`<div class="meeple-actions">`)
		for _, slot := range st.MeepleSlots {
			abbot := "false"
			if slot.Abbot {
				abbot = "true"
			}
			b.WriteString(fmt.Sprintf(
				`<button class="btn" data-on:click="$kind='%s'; $slot=%d; $abbot=%s; @post('/api/rooms/%s/place-meeple')">%s</button>`,
				slot.Kind, slot.Slot, abbot, st.RoomCode, html.EscapeString(slot.Label),
			))
		}
		b.WriteString(fmt.Sprintf(`<button class="btn ghost" data-on:click="@post('/api/rooms/%s/skip-meeple')">Skip</button>`, st.RoomCode))
		if st.Settings.Expansions.Abbot {
			b.WriteString(fmt.Sprintf(`<button class="btn ghost" data-on:click="@post('/api/rooms/%s/recall-abbot')">Recall abbot</button>`, st.RoomCode))
		}
		b.WriteString(`</div>`)
	}
	if st.Phase == game.PhasePlaceTile {
		b.WriteString(`<div class="hint">Click a highlighted cell · R rotates</div>`)
	}
	b.WriteString(`</div>`)
	return b.String()
}
