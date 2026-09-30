package game

import "fmt"

var botNames = []string{"Acolyte", "Miller", "Shepherd", "Squire", "Friar", "Peddler"}

func (e *Engine) AddBot() (*Player, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.Phase != PhaseLobby {
		return nil, fmt.Errorf("game already started")
	}
	if len(e.Players) >= e.Settings.MaxPlayers {
		return nil, fmt.Errorf("room is full")
	}
	name := botNames[len(e.Players)%len(botNames)]
	for _, p := range e.Players {
		if p.Name == name {
			name = fmt.Sprintf("%s %d", name, len(e.Players)+1)
			break
		}
	}
	id := fmt.Sprintf("bot-%d-%d", len(e.Players), e.rng.Int63())
	color := Colors[len(e.Players)%len(Colors)]
	p := Player{ID: id, Name: name, Color: color, Score: 0, Meeple: 7, IsBot: true}
	if e.Settings.Expansions.Abbot {
		p.Abbot = 1
	}
	e.Players = append(e.Players, p)
	return &p, nil
}

func (e *Engine) CurrentPlayerIsBot() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.currentPlayerIsBotLocked()
}

func (e *Engine) currentPlayerIsBotLocked() bool {
	if e.Phase == PhaseLobby || e.Phase == PhaseGameOver {
		return false
	}
	if e.Turn < 0 || e.Turn >= len(e.Players) {
		return false
	}
	return e.Players[e.Turn].IsBot
}

func (e *Engine) BotAct() (bool, error) {
	e.mu.Lock()
	if !e.currentPlayerIsBotLocked() {
		e.mu.Unlock()
		return false, nil
	}
	playerID := e.Players[e.Turn].ID
	phase := e.Phase
	meepleLeft := e.Players[e.Turn].Meeple
	var spotX, spotY, spotRot int
	var meepleKind FeatureKind
	var meepleSlot int
	var asAbbot bool
	var skip bool
	switch phase {
	case PhasePlaceTile:
		if e.Current == nil {
			e.mu.Unlock()
			return false, fmt.Errorf("no tile to place")
		}
		spots := e.validSpotsLocked(*e.Current)
		if len(spots) == 0 {
			e.mu.Unlock()
			return false, fmt.Errorf("no valid spots")
		}
		spot := spots[e.rng.Intn(len(spots))]
		spotX, spotY = spot.X, spot.Y
		spotRot = spot.Rots[e.rng.Intn(len(spot.Rots))]
	case PhasePlaceMeeple:
		if e.LastPlaced == nil {
			e.mu.Unlock()
			return false, fmt.Errorf("no tile")
		}
		slots := e.availableMeepleSlotsLocked(*e.LastPlaced)
		choice := pickBotMeeple(slots, e.Players[e.Turn], e.rng.Intn(100))
		if choice == nil || (meepleLeft <= 2 && e.rng.Intn(100) < 35 && !choice.Abbot) {
			skip = true
		} else {
			meepleKind = choice.Kind
			meepleSlot = choice.Slot
			asAbbot = choice.Abbot
		}
	default:
		e.mu.Unlock()
		return false, nil
	}
	e.mu.Unlock()

	switch phase {
	case PhasePlaceTile:
		if err := e.PlaceTile(playerID, spotX, spotY, spotRot); err != nil {
			return false, err
		}
		return true, nil
	case PhasePlaceMeeple:
		if skip {
			if err := e.SkipMeeple(playerID); err != nil {
				return false, err
			}
			return true, nil
		}
		if err := e.PlaceMeeple(playerID, meepleKind, meepleSlot, asAbbot); err != nil {
			if err2 := e.SkipMeeple(playerID); err2 != nil {
				return false, err
			}
			return true, nil
		}
		return true, nil
	default:
		return false, nil
	}
}

func pickBotMeeple(slots []MeepleSlot, p Player, roll int) *MeepleSlot {
	if len(slots) == 0 {
		return nil
	}
	priority := func(s MeepleSlot) int {
		if s.Abbot {
			if s.Kind == FeatGarden {
				return 3
			}
			return 1
		}
		switch s.Kind {
		case FeatMonastery:
			return 5
		case FeatCity:
			return 4
		case FeatRoad:
			return 3
		case FeatField:
			if p.Meeple >= 4 {
				return 2
			}
			return 0
		default:
			return 0
		}
	}
	var best *MeepleSlot
	bestScore := 0
	for i := range slots {
		s := &slots[i]
		sc := priority(*s)
		if sc > bestScore {
			bestScore = sc
			best = s
		}
	}
	if bestScore == 0 {
		return nil
	}
	if roll < 15 {
		return nil
	}
	return best
}
