package game

import (
	"testing"
)

func twoPlayerEngine(farmers bool) *Engine {
	e := NewEngine("TEST", Settings{MaxPlayers: 5, Expansions: Expansions{Farmers: farmers}, Name: "t"}, 1)
	e.AddPlayer("p1", "A")
	e.AddPlayer("p2", "B")
	e.Start()
	e.Deck = []string{"ffff_m", "ffff_m", "ffff_m", "ffff_m", "cfff", "cfff"}
	e.Current = nil
	e.Phase = PhasePlaceTile
	e.Turn = 0
	e.Meeples = nil
	e.LastPlaced = nil
	e.Players[0].Meeple = 7
	e.Players[1].Meeple = 7
	e.Players[0].Score = 0
	e.Players[1].Score = 0
	e.Board = map[string]PlacedTile{
		key(0, 0): {DefID: "start", X: 0, Y: 0, Rotation: 0},
	}
	return e
}

func forceCurrent(e *Engine, defID string) {
	id := defID
	e.Current = &id
	e.Phase = PhasePlaceTile
}

func slotPresent(slots []MeepleSlot, kind FeatureKind, slot int) bool {
	for _, s := range slots {
		if s.Kind == kind && s.Slot == slot {
			return true
		}
	}
	return false
}

func hasKind(slots []MeepleSlot, kind FeatureKind) bool {
	for _, s := range slots {
		if s.Kind == kind {
			return true
		}
	}
	return false
}

func TestCannotPlaceOnOccupiedCityNetwork(t *testing.T) {
	e := twoPlayerEngine(true)
	e.Board[key(0, -1)] = PlacedTile{DefID: "ccff", X: 0, Y: -1, Rotation: 2}
	e.Meeples = []Meeple{{PlayerID: "p2", X: 0, Y: -1, Kind: FeatCity, Slot: 0}}
	e.Players[1].Meeple = 6

	forceCurrent(e, "cfff")
	if err := e.PlaceTile("p1", -1, -1, 1); err != nil {
		t.Fatalf("place city onto open west arm: %v", err)
	}
	if e.Phase != PhasePlaceMeeple {
		t.Fatalf("expected meeple phase, got %s", e.Phase)
	}
	slots := e.Public("p1").MeepleSlots
	if slotPresent(slots, FeatCity, 0) {
		t.Fatalf("city must be blocked by opponent meeple on connected city; slots=%v", slots)
	}
	if err := e.PlaceMeeple("p1", FeatCity, 0, false); err == nil {
		t.Fatal("expected PlaceMeeple on occupied city to fail")
	}
	if !hasKind(slots, FeatField) {
		t.Fatalf("field should remain available when farmers on; slots=%v", slots)
	}
}

func TestCannotPlaceOnOccupiedRoadNetworkAcrossTiles(t *testing.T) {
	e := twoPlayerEngine(true)
	e.Board = map[string]PlacedTile{
		key(0, 0): {DefID: "rfrf", X: 0, Y: 0, Rotation: 0},
	}
	e.Meeples = []Meeple{{PlayerID: "p2", X: 0, Y: 0, Kind: FeatRoad, Slot: 0}}
	e.Players[1].Meeple = 6

	forceCurrent(e, "rfrf")
	if err := e.PlaceTile("p1", 0, 1, 0); err != nil {
		t.Fatalf("place connecting road: %v", err)
	}
	if e.Phase != PhasePlaceMeeple {
		t.Fatalf("expected meeple phase (fields keep slots alive), got %s", e.Phase)
	}
	slots := e.Public("p1").MeepleSlots
	if slotPresent(slots, FeatRoad, 0) {
		t.Fatalf("occupied road must not be offered as a slot; slots=%v", slots)
	}
	if err := e.PlaceMeeple("p1", FeatRoad, 0, false); err == nil {
		t.Fatal("expected PlaceMeeple on occupied road to fail")
	}
}

func TestRoadCompleteOpenStraightRoad(t *testing.T) {
	e := twoPlayerEngine(false)
	e.Board = map[string]PlacedTile{
		key(0, 0): {DefID: "rfrf", X: 0, Y: 0, Rotation: 0},
		key(0, 1): {DefID: "rfrf", X: 0, Y: 1, Rotation: 0},
	}
	e.mu.Lock()
	segs := e.collectFeature(0, 0, FeatRoad, 0)
	complete := e.roadComplete(segs)
	e.mu.Unlock()
	if complete {
		t.Fatal("open straight road must not be complete")
	}
}

func TestFeatureHasMeepleWalksConnectedRoad(t *testing.T) {
	e := twoPlayerEngine(false)
	e.Board = map[string]PlacedTile{
		key(0, 0): {DefID: "rfrf", X: 0, Y: 0, Rotation: 0},
		key(0, 1): {DefID: "rfrf", X: 0, Y: 1, Rotation: 0},
	}
	e.Meeples = []Meeple{{PlayerID: "p2", X: 0, Y: 0, Kind: FeatRoad, Slot: 0}}
	e.mu.Lock()
	occupied := e.featureHasMeepleLocked(0, 1, FeatRoad, 0)
	e.mu.Unlock()
	if !occupied {
		t.Fatal("connected road across tiles must count as occupied")
	}
}

func TestFeatureHasMeepleWalksConnectedCity(t *testing.T) {
	e := twoPlayerEngine(false)
	e.Board = map[string]PlacedTile{
		key(0, 0):  {DefID: "start", X: 0, Y: 0, Rotation: 0},
		key(0, -1): {DefID: "ccff", X: 0, Y: -1, Rotation: 2},
		key(-1, -1): {DefID: "cfff", X: -1, Y: -1, Rotation: 1},
	}
	e.Meeples = []Meeple{{PlayerID: "p1", X: 0, Y: -1, Kind: FeatCity, Slot: 0}}
	e.mu.Lock()
	occupied := e.featureHasMeepleLocked(-1, -1, FeatCity, 0)
	e.mu.Unlock()
	if !occupied {
		t.Fatal("connected city across tiles must count as occupied")
	}
}

func TestCanPlaceOnSameTileDifferentUnoccupiedFeatures(t *testing.T) {
	e := twoPlayerEngine(false)
	forceCurrent(e, "crrr")
	if err := e.PlaceTile("p1", 0, -1, 2); err != nil {
		t.Fatalf("place crrr north of start: %v", err)
	}
	if e.Phase != PhasePlaceMeeple {
		t.Fatalf("expected meeple phase, got %s", e.Phase)
	}
	slots := e.Public("p1").MeepleSlots
	if !hasKind(slots, FeatCity) {
		t.Fatalf("expected city slot; slots=%v", slots)
	}
	roadSlots := 0
	for _, s := range slots {
		if s.Kind == FeatRoad {
			roadSlots++
		}
	}
	if roadSlots == 0 {
		t.Fatalf("expected road slot(s) on same tile; slots=%v", slots)
	}
	before := e.Players[0].Meeple
	if err := e.PlaceMeeple("p1", FeatRoad, 0, false); err != nil {
		t.Fatalf("place on road while city also free: %v", err)
	}
	if e.Players[0].Meeple > before {
		t.Fatal("supply should not increase on place")
	}
	if e.Players[0].Meeple < before-1 {
		t.Fatalf("supply should drop by at most 1; before=%d after=%d", before, e.Players[0].Meeple)
	}
}

func TestSkipMeepleAllowed(t *testing.T) {
	e := twoPlayerEngine(false)
	forceCurrent(e, "cfff")
	if err := e.PlaceTile("p1", 0, -1, 2); err != nil {
		t.Fatalf("place: %v", err)
	}
	if e.Phase != PhasePlaceMeeple {
		t.Fatalf("expected meeple phase, got %s", e.Phase)
	}
	supply := e.Players[0].Meeple
	boardN := len(e.Meeples)
	if err := e.SkipMeeple("p1"); err != nil {
		t.Fatalf("skip: %v", err)
	}
	if e.Players[0].Meeple != supply {
		t.Fatalf("skip must not change supply; got %d want %d", e.Players[0].Meeple, supply)
	}
	if len(e.Meeples) != boardN {
		t.Fatal("skip must not place meeple")
	}
	if e.Turn != 1 {
		t.Fatalf("turn should advance to p2, got %d", e.Turn)
	}
}

func TestFarmersGatedBySettings(t *testing.T) {
	off := twoPlayerEngine(false)
	forceCurrent(off, "cfff")
	if err := off.PlaceTile("p1", 0, -1, 2); err != nil {
		t.Fatalf("place: %v", err)
	}
	if hasKind(off.Public("p1").MeepleSlots, FeatField) {
		t.Fatal("field slots must be absent when farmers disabled")
	}
	if err := off.PlaceMeeple("p1", FeatField, 0, false); err == nil {
		t.Fatal("PlaceMeeple field must fail when farmers disabled")
	}

	on := twoPlayerEngine(true)
	forceCurrent(on, "cfff")
	if err := on.PlaceTile("p1", 0, -1, 2); err != nil {
		t.Fatalf("place: %v", err)
	}
	if on.Phase != PhasePlaceMeeple {
		t.Fatalf("expected meeple phase, got %s", on.Phase)
	}
	if !hasKind(on.Public("p1").MeepleSlots, FeatField) {
		t.Fatalf("field slots must be present when farmers enabled; slots=%v", on.Public("p1").MeepleSlots)
	}
}

func TestMeepleSupplyDecrementsAndReturnsOnScore(t *testing.T) {
	e := twoPlayerEngine(false)
	forceCurrent(e, "cfff")
	if err := e.PlaceTile("p1", 0, -1, 2); err != nil {
		t.Fatalf("place completing city tile: %v", err)
	}
	if e.Phase != PhasePlaceMeeple {
		t.Fatalf("expected meeple phase, got %s", e.Phase)
	}
	before := e.Players[0].Meeple
	if err := e.PlaceMeeple("p1", FeatCity, 0, false); err != nil {
		t.Fatalf("place meeple: %v", err)
	}
	if e.Players[0].Meeple != before {
		t.Fatalf("completed city must return meeple; before=%d after=%d score=%d meeples=%v",
			before, e.Players[0].Meeple, e.Players[0].Score, e.Meeples)
	}
	if e.Players[0].Score != 4 {
		t.Fatalf("2-tile city should score 4, got %d", e.Players[0].Score)
	}
	for _, m := range e.Meeples {
		if m.Kind == FeatCity {
			t.Fatalf("scored city meeple should be removed: %+v", m)
		}
	}
}

func TestMeepleSupplyDecrementsWhenFeatureIncomplete(t *testing.T) {
	e := twoPlayerEngine(false)
	forceCurrent(e, "ccff")
	if err := e.PlaceTile("p1", 0, -1, 2); err != nil {
		t.Fatalf("place incomplete city: %v", err)
	}
	before := e.Players[0].Meeple
	if err := e.PlaceMeeple("p1", FeatCity, 0, false); err != nil {
		t.Fatalf("place meeple: %v", err)
	}
	if e.Players[0].Meeple != before-1 {
		t.Fatalf("incomplete feature must keep meeple out of supply; before=%d after=%d", before, e.Players[0].Meeple)
	}
	if len(e.Meeples) != 1 {
		t.Fatalf("meeple should remain on board, got %d", len(e.Meeples))
	}
}

func TestPlaceOnlyOnJustPlacedTile(t *testing.T) {
	e := twoPlayerEngine(false)
	forceCurrent(e, "rfrf")
	if err := e.PlaceTile("p1", 1, 0, 1); err != nil {
		t.Fatalf("place road east: %v", err)
	}
	if e.Phase != PhasePlaceMeeple {
		t.Fatalf("expected meeple phase, got %s", e.Phase)
	}
	if err := e.PlaceMeeple("p1", FeatCity, 0, false); err == nil {
		t.Fatal("must not place on a city that is not a slot of the just-placed tile")
	}
}

func TestSevenMeeplesInSupply(t *testing.T) {
	e := NewEngine("TEST", Settings{MaxPlayers: 2}, 1)
	p, err := e.AddPlayer("p1", "A")
	if err != nil {
		t.Fatal(err)
	}
	if p.Meeple != 7 {
		t.Fatalf("want 7 meeples, got %d", p.Meeple)
	}
}

func TestStartAndCfrfTileDataInconsistency(t *testing.T) {
	cat := BuildCatalog()
	for _, id := range []string{"start", "cfrf"} {
		td := cat[id]
		for _, g := range td.RoadGroups {
			for _, d := range g {
				if td.Edges[d] != Road {
					t.Errorf("%s: RoadGroups includes dir %v but edge is %v", id, d, td.Edges[d])
				}
			}
		}
		for d := Dir(0); d < 4; d++ {
			if td.Edges[d] != Road {
				continue
			}
			found := false
			for _, g := range td.RoadGroups {
				for _, ed := range g {
					if ed == d {
						found = true
					}
				}
			}
			if !found {
				t.Errorf("%s: Road edge %v not in any RoadGroup", id, d)
			}
		}
	}
}

func TestStartEastRoadOccupationPropagates(t *testing.T) {
	e := twoPlayerEngine(false)
	e.Board = map[string]PlacedTile{
		key(0, 0): {DefID: "start", X: 0, Y: 0, Rotation: 0},
		key(1, 0): {DefID: "rfrf", X: 1, Y: 0, Rotation: 1},
	}
	e.Meeples = []Meeple{{PlayerID: "p1", X: 0, Y: 0, Kind: FeatRoad, Slot: 0}}
	e.mu.Lock()
	joined := e.featureHasMeepleLocked(1, 0, FeatRoad, 0)
	e.mu.Unlock()
	if !joined {
		t.Fatal("start east-west road must block meeple on connecting east road")
	}
}
