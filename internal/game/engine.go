package game

import (
	"fmt"
	"math/rand"
	"sync"
)

type Engine struct {
	mu         sync.Mutex
	RoomCode   string
	Settings   Settings
	Phase      Phase
	Players    []Player
	Turn       int
	Board      map[string]PlacedTile
	Meeples    []Meeple
	Deck       []string
	Current    *string
	LastPlaced *PlacedTile
	Catalog    map[string]TileDef
	Message    string
	WinnerIDs  []string
	rng        *rand.Rand
}

func NewEngine(code string, settings Settings, seed int64) *Engine {
	catalog := BuildCatalogWithExpansions(settings.Expansions)
	e := &Engine{
		RoomCode: code,
		Settings: settings,
		Phase:    PhaseLobby,
		Board:    make(map[string]PlacedTile),
		Catalog:  catalog,
		rng:      rand.New(rand.NewSource(seed)),
	}
	return e
}

func key(x, y int) string {
	return fmt.Sprintf("%d,%d", x, y)
}

func (e *Engine) AddPlayer(id, name string) (*Player, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.Phase != PhaseLobby {
		return nil, fmt.Errorf("game already started")
	}
	if len(e.Players) >= e.Settings.MaxPlayers {
		return nil, fmt.Errorf("room is full")
	}
	for _, p := range e.Players {
		if p.ID == id {
			return &p, nil
		}
	}
	color := Colors[len(e.Players)%len(Colors)]
	p := Player{ID: id, Name: name, Color: color, Score: 0, Meeple: 7}
	if e.Settings.Expansions.Abbot {
		p.Abbot = 1
	}
	e.Players = append(e.Players, p)
	return &p, nil
}

func (e *Engine) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.Phase != PhaseLobby {
		return fmt.Errorf("already started")
	}
	if len(e.Players) < 2 {
		return fmt.Errorf("need at least 2 players")
	}
	cap := e.Settings.Expansions.MaxPlayersCap()
	if e.Settings.MaxPlayers > cap {
		e.Settings.MaxPlayers = cap
	}
	e.Catalog = BuildCatalogWithExpansions(e.Settings.Expansions)
	land := BuildDeckFor(e.Catalog, e.Settings.Expansions)
	e.rng.Shuffle(len(land), func(i, j int) { land[i], land[j] = land[j], land[i] })

	if e.Settings.Expansions.River {
		river := BuildRiverStack(e.Catalog)
		e.rng.Shuffle(len(river), func(i, j int) { river[i], river[j] = river[j], river[i] })
		if _, ok := e.Catalog["river_lake"]; ok {
			river = append(river, "river_lake")
		}
		e.Deck = append(river, land...)
		e.Board[key(0, 0)] = PlacedTile{DefID: "river_source", X: 0, Y: 0, Rotation: 0}
	} else {
		e.Deck = land
		e.Board[key(0, 0)] = PlacedTile{DefID: "start", X: 0, Y: 0, Rotation: 0}
	}
	e.Turn = 0
	e.drawLocked()
	e.Phase = PhasePlaceTile
	e.Message = fmt.Sprintf("%s's turn — place a tile", e.Players[e.Turn].Name)
	return nil
}

func (e *Engine) drawLocked() {
	for len(e.Deck) > 0 {
		id := e.Deck[0]
		e.Deck = e.Deck[1:]
		e.Current = &id
		if len(e.validSpotsLocked(*e.Current)) > 0 {
			return
		}
	}
	e.Current = nil
	e.endGameLocked()
}

func (e *Engine) PlaceTile(playerID string, x, y, rot int) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.Phase != PhasePlaceTile {
		return fmt.Errorf("not tile placement phase")
	}
	if e.Players[e.Turn].ID != playerID {
		return fmt.Errorf("not your turn")
	}
	if e.Current == nil {
		return fmt.Errorf("no tile to place")
	}
	if !e.canPlaceLocked(*e.Current, x, y, rot) {
		return fmt.Errorf("invalid placement")
	}
	pt := PlacedTile{DefID: *e.Current, X: x, Y: y, Rotation: rot % 4}
	e.Board[key(x, y)] = pt
	e.LastPlaced = &pt
	e.Current = nil
	slots := e.availableMeepleSlotsLocked(pt)
	if len(slots) > 0 {
		e.Phase = PhasePlaceMeeple
		msg := fmt.Sprintf("%s — place a meeple or skip", e.Players[e.Turn].Name)
		if e.Settings.Expansions.Abbot && e.Players[e.Turn].Abbot == 0 {
			for _, m := range e.Meeples {
				if m.PlayerID == e.Players[e.Turn].ID && m.IsAbbot {
					msg = fmt.Sprintf("%s — place, skip, or recall abbot", e.Players[e.Turn].Name)
					break
				}
			}
		}
		e.Message = msg
	} else {
		e.scoreAfterPlacementLocked()
		e.advanceTurnLocked()
	}
	return nil
}

func (e *Engine) availableMeepleSlotsLocked(pt PlacedTile) []MeepleSlot {
	p := e.Players[e.Turn]
	out := []MeepleSlot{}
	for _, s := range e.meepleSlotsLocked(pt) {
		if s.Abbot {
			if p.Abbot > 0 {
				out = append(out, s)
			}
			continue
		}
		if p.Meeple > 0 {
			out = append(out, s)
		}
	}
	return out
}

func (e *Engine) PlaceMeeple(playerID string, kind FeatureKind, slot int, asAbbot bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.Phase != PhasePlaceMeeple {
		return fmt.Errorf("not meeple phase")
	}
	if e.Players[e.Turn].ID != playerID {
		return fmt.Errorf("not your turn")
	}
	if e.LastPlaced == nil {
		return fmt.Errorf("no tile")
	}
	ok := false
	for _, s := range e.availableMeepleSlotsLocked(*e.LastPlaced) {
		if s.Kind == kind && s.Slot == slot && s.Abbot == asAbbot {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("invalid meeple slot")
	}
	if asAbbot {
		if !e.Settings.Expansions.Abbot || e.Players[e.Turn].Abbot <= 0 {
			return fmt.Errorf("no abbot left")
		}
		e.Players[e.Turn].Abbot--
	} else {
		if e.Players[e.Turn].Meeple <= 0 {
			return fmt.Errorf("no meeples left")
		}
		e.Players[e.Turn].Meeple--
	}
	e.Meeples = append(e.Meeples, Meeple{
		PlayerID: playerID,
		X:        e.LastPlaced.X,
		Y:        e.LastPlaced.Y,
		Kind:     kind,
		Slot:     slot,
		IsAbbot:  asAbbot,
	})
	e.scoreAfterPlacementLocked()
	e.advanceTurnLocked()
	return nil
}

func (e *Engine) SkipMeeple(playerID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.Phase != PhasePlaceMeeple {
		return fmt.Errorf("not meeple phase")
	}
	if e.Players[e.Turn].ID != playerID {
		return fmt.Errorf("not your turn")
	}
	e.scoreAfterPlacementLocked()
	e.advanceTurnLocked()
	return nil
}

func (e *Engine) RecallAbbot(playerID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.Phase != PhasePlaceMeeple {
		return fmt.Errorf("not meeple phase")
	}
	if e.Players[e.Turn].ID != playerID {
		return fmt.Errorf("not your turn")
	}
	if !e.Settings.Expansions.Abbot {
		return fmt.Errorf("abbot expansion off")
	}
	idx := -1
	for i, m := range e.Meeples {
		if m.PlayerID == playerID && m.IsAbbot {
			idx = i
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("no abbot on board")
	}
	m := e.Meeples[idx]
	pts := e.monasteryScore(m.X, m.Y)
	for i := range e.Players {
		if e.Players[i].ID == playerID {
			e.Players[i].Score += pts
			e.Players[i].Abbot++
		}
	}
	e.Meeples = append(e.Meeples[:idx], e.Meeples[idx+1:]...)
	e.scoreAfterPlacementLocked()
	e.advanceTurnLocked()
	return nil
}

func (e *Engine) advanceTurnLocked() {
	if e.Phase == PhaseGameOver {
		return
	}
	e.Turn = (e.Turn + 1) % len(e.Players)
	e.drawLocked()
	if e.Phase == PhaseGameOver {
		return
	}
	e.Phase = PhasePlaceTile
	e.Message = fmt.Sprintf("%s's turn — place a tile", e.Players[e.Turn].Name)
}

func (e *Engine) endGameLocked() {
	e.Phase = PhaseGameOver
	e.finalScoreLocked()
	best := -1
	winners := []string{}
	for _, p := range e.Players {
		if p.Score > best {
			best = p.Score
			winners = []string{p.ID}
		} else if p.Score == best {
			winners = append(winners, p.ID)
		}
	}
	e.WinnerIDs = winners
	names := []string{}
	for _, id := range winners {
		for _, p := range e.Players {
			if p.ID == id {
				names = append(names, p.Name)
			}
		}
	}
	e.Message = fmt.Sprintf("Game over — winner: %v", names)
}

func (e *Engine) canPlaceLocked(defID string, x, y, rot int) bool {
	if _, exists := e.Board[key(x, y)]; exists {
		return false
	}
	def := e.Catalog[defID]
	adjacent := false
	riverTouch := false
	for d := Dir(0); d < 4; d++ {
		dx, dy := d.Delta()
		nx, ny := x+dx, y+dy
		nb, ok := e.Board[key(nx, ny)]
		if !ok {
			continue
		}
		adjacent = true
		nbDef := e.Catalog[nb.DefID]
		myEdge := EdgeTerrain(def, rot, d)
		theirEdge := EdgeTerrain(nbDef, nb.Rotation, d.Opposite())
		if myEdge != theirEdge {
			return false
		}
		myR := hasRiverEdge(def, rot, d)
		theirR := hasRiverEdge(nbDef, nb.Rotation, d.Opposite())
		if myR != theirR {
			return false
		}
		if myR && theirR {
			riverTouch = true
		}
	}
	if !adjacent {
		return false
	}
	if def.River && !riverTouch {
		return false
	}
	return true
}

func hasRiverEdge(def TileDef, rotation int, d Dir) bool {
	for _, rd := range RotatedRiverEdges(def, rotation) {
		if rd == d {
			return true
		}
	}
	return false
}

func (e *Engine) validSpotsLocked(defID string) []Spot {
	candidates := map[string]struct{}{}
	for _, t := range e.Board {
		for d := Dir(0); d < 4; d++ {
			dx, dy := d.Delta()
			x, y := t.X+dx, t.Y+dy
			k := key(x, y)
			if _, occupied := e.Board[k]; occupied {
				continue
			}
			candidates[k] = struct{}{}
		}
	}
	spots := []Spot{}
	for k := range candidates {
		var x, y int
		fmt.Sscanf(k, "%d,%d", &x, &y)
		rots := []int{}
		for r := 0; r < 4; r++ {
			if e.canPlaceLocked(defID, x, y, r) {
				rots = append(rots, r)
			}
		}
		if len(rots) > 0 {
			spots = append(spots, Spot{X: x, Y: y, Rots: rots})
		}
	}
	return spots
}

func (e *Engine) Public(youID string) PublicState {
	e.mu.Lock()
	defer e.mu.Unlock()
	board := make([]PlacedTile, 0, len(e.Board))
	for _, t := range e.Board {
		board = append(board, t)
	}
	var spots []Spot
	var slots []MeepleSlot
	if e.Phase == PhasePlaceTile && e.Current != nil {
		spots = e.validSpotsLocked(*e.Current)
	}
	if e.Phase == PhasePlaceMeeple && e.LastPlaced != nil {
		slots = e.availableMeepleSlotsLocked(*e.LastPlaced)
	}
	players := append([]Player{}, e.Players...)
	meeples := append([]Meeple{}, e.Meeples...)
	var cur *string
	if e.Current != nil {
		c := *e.Current
		cur = &c
	}
	var last *PlacedTile
	if e.LastPlaced != nil {
		l := *e.LastPlaced
		last = &l
	}
	return PublicState{
		RoomCode:    e.RoomCode,
		Settings:    e.Settings,
		Phase:       e.Phase,
		Players:     players,
		Turn:        e.Turn,
		Board:       board,
		Meeples:     meeples,
		CurrentTile: cur,
		TilesLeft:   len(e.Deck) + boolToInt(e.Current != nil),
		ValidSpots:  spots,
		MeepleSlots: slots,
		LastPlaced:  last,
		Message:     e.Message,
		WinnerIDs:   append([]string{}, e.WinnerIDs...),
		TileCatalog: e.Catalog,
		YouID:       youID,
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
