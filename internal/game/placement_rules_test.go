package game

import (
	"testing"
)

func TestEdgeTerrainRotation(t *testing.T) {
	def := T("CRFR", nil, nil, false, 0)
	cases := []struct {
		rot  int
		dir  Dir
		want Terrain
	}{
		{0, N, City},
		{0, E, Road},
		{0, S, Field},
		{0, W, Road},
		{1, N, Road},
		{1, E, City},
		{1, S, Road},
		{1, W, Field},
		{2, N, Field},
		{2, E, Road},
		{2, S, City},
		{2, W, Road},
		{3, N, Road},
		{3, E, Field},
		{3, S, Road},
		{3, W, City},
	}
	for _, tc := range cases {
		got := EdgeTerrain(def, tc.rot, tc.dir)
		if got != tc.want {
			t.Fatalf("rot=%d dir=%v: got %v want %v", tc.rot, tc.dir, got, tc.want)
		}
	}
}

func TestCanPlaceRequiresAdjacency(t *testing.T) {
	e := NewEngine("T", Settings{MaxPlayers: 2}, 1)
	e.Board[key(0, 0)] = PlacedTile{DefID: "start", X: 0, Y: 0, Rotation: 0}
	if e.canPlaceLocked("cfff", 5, 5, 0) {
		t.Fatal("floating tile must be rejected")
	}
	if e.canPlaceLocked("cfff", 1, 1, 0) {
		t.Fatal("diagonal-only contact must be rejected")
	}
}

func TestCanPlaceRejectsOccupied(t *testing.T) {
	e := NewEngine("T", Settings{MaxPlayers: 2}, 1)
	e.Board[key(0, 0)] = PlacedTile{DefID: "start", X: 0, Y: 0, Rotation: 0}
	if e.canPlaceLocked("cfff", 0, 0, 0) {
		t.Fatal("overlap must be rejected")
	}
}

func TestCanPlaceEdgeMatchingAndRotation(t *testing.T) {
	e := NewEngine("T", Settings{MaxPlayers: 2}, 1)
	e.Catalog["anchor"] = T("CRFR", [][]Dir{{N}}, [][]Dir{{E, W}}, false, 0)
	e.Board[key(0, 0)] = PlacedTile{DefID: "anchor", X: 0, Y: 0, Rotation: 0}

	if e.canPlaceLocked("cfff", 1, 0, 0) {
		t.Fatal("field against road must be illegal")
	}
	if !e.canPlaceLocked("rffr", 1, 0, 0) {
		t.Fatal("road against road must be legal")
	}
	if e.canPlaceLocked("cfff", 0, -1, 0) {
		t.Fatal("cfff rot 0 puts field south against city — illegal")
	}
	if !e.canPlaceLocked("cfff", 0, -1, 2) {
		t.Fatal("cfff rot 2 puts city south against city — legal")
	}
	if e.canPlaceLocked("cfff", 0, 1, 0) {
		t.Fatal("cfff rot 0 puts city north against field — illegal")
	}
	if !e.canPlaceLocked("cfff", 0, 1, 2) {
		t.Fatal("cfff rot 2 puts field north against field — legal")
	}
}

func TestCanPlaceMatchesAllSharedEdges(t *testing.T) {
	e := NewEngine("T", Settings{MaxPlayers: 2}, 1)
	e.Catalog["anchor"] = T("CRFR", [][]Dir{{N}}, [][]Dir{{E, W}}, false, 0)
	e.Board[key(0, 0)] = PlacedTile{DefID: "anchor", X: 0, Y: 0, Rotation: 0}
	e.Board[key(2, 0)] = PlacedTile{DefID: "cfff", X: 2, Y: 0, Rotation: 1}

	if e.canPlaceLocked("rfrf", 1, 0, 0) {
		t.Fatal("must fail when one neighbor matches and the other does not")
	}
	if !e.canPlaceLocked("rffr", 1, 0, 0) {
		t.Fatal("road curve should fit between east-road and west-field neighbors")
	}
}

func TestStartTileEdgesMatchClassicCRFR(t *testing.T) {
	want := [4]Terrain{City, Road, Field, Road}
	for _, id := range []string{"start", "cfrf"} {
		got := BuildCatalog()[id].Edges
		if got != want {
			t.Fatalf("%s edges=%v want classic start/city-road CRFR %v", id, got, want)
		}
	}
}

func TestDrawDiscardsUnplaceableAndRedraws(t *testing.T) {
	e := NewEngine("T", Settings{MaxPlayers: 2}, 1)
	e.Board[key(0, 0)] = PlacedTile{DefID: "cfff", X: 0, Y: 0, Rotation: 0}
	e.Deck = []string{"rrrr", "ffff_m"}
	e.drawLocked()
	if e.Current == nil {
		t.Fatal("expected a drawable tile")
	}
	if *e.Current != "ffff_m" {
		t.Fatalf("unplaceable rrrr must be discarded; got %s", *e.Current)
	}
	if len(e.Deck) != 0 {
		t.Fatalf("deck should be empty after draw, got %d", len(e.Deck))
	}
	if len(e.validSpotsLocked(*e.Current)) == 0 {
		t.Fatal("drawn tile must have at least one legal spot")
	}
}

func TestDrawEndsGameWhenNoPlaceableTilesRemain(t *testing.T) {
	e := NewEngine("T", Settings{MaxPlayers: 2}, 1)
	e.Players = []Player{{ID: "p1", Name: "A"}, {ID: "p2", Name: "B"}}
	e.Board[key(0, 0)] = PlacedTile{DefID: "cfff", X: 0, Y: 0, Rotation: 0}
	e.Deck = []string{"rrrr", "rrrr"}
	e.Phase = PhasePlaceTile
	e.drawLocked()
	if e.Current != nil {
		t.Fatal("no placeable tile should leave Current nil")
	}
	if e.Phase != PhaseGameOver {
		t.Fatalf("expected game over, got %s", e.Phase)
	}
}

func TestStartPlacesStartTileAndDraws(t *testing.T) {
	e := NewEngine("T", Settings{MaxPlayers: 2}, 7)
	e.AddPlayer("p1", "A")
	e.AddPlayer("p2", "B")
	if err := e.Start(); err != nil {
		t.Fatal(err)
	}
	start, ok := e.Board[key(0, 0)]
	if !ok || start.DefID != "start" {
		t.Fatal("starting tile must be on the board at 0,0")
	}
	for _, id := range e.Deck {
		if id == "start" {
			t.Fatal("starting tile must not remain in the draw deck")
		}
	}
	if e.Current == nil {
		t.Fatal("first player must receive a drawn tile")
	}
	if e.Phase != PhasePlaceTile {
		t.Fatalf("phase %s", e.Phase)
	}
	if len(e.validSpotsLocked(*e.Current)) == 0 {
		t.Fatal("drawn opening tile must be placeable")
	}
}

func TestClassicBaseDeckSize(t *testing.T) {
	catalog := BuildCatalog()
	deck := BuildDeck(catalog)
	if len(deck) != 71 {
		t.Fatalf("classic base set draws from 71 tiles (72 including start); got deck=%d", len(deck))
	}
}

func TestCloisterWithRoadHasSingleRoadEdge(t *testing.T) {
	def, ok := BuildCatalog()["frfr_m"]
	if !ok {
		t.Fatal("missing cloister+road tile")
	}
	roads := 0
	for _, terr := range def.Edges {
		if terr == Road {
			roads++
		}
	}
	if roads != 1 {
		t.Fatalf("classic cloister+road has exactly 1 road edge; got %d (edges=%v)", roads, def.Edges)
	}
}

func TestBaseSetIncludesFourSidedCity(t *testing.T) {
	catalog := BuildCatalog()
	found := false
	for _, def := range catalog {
		if def.Edges == [4]Terrain{City, City, City, City} {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("classic base set includes the 4-sided city tile (CCCC)")
	}
}

func TestValidSpotsAgreeWithCanPlace(t *testing.T) {
	e := NewEngine("T", Settings{MaxPlayers: 2}, 1)
	e.Board[key(0, 0)] = PlacedTile{DefID: "start", X: 0, Y: 0, Rotation: 0}
	spots := e.validSpotsLocked("rffr")
	if len(spots) == 0 {
		t.Fatal("expected valid spots for road curve next to start")
	}
	for _, s := range spots {
		for _, r := range s.Rots {
			if !e.canPlaceLocked("rffr", s.X, s.Y, r) {
				t.Fatalf("validSpots listed illegal rot %d at %d,%d", r, s.X, s.Y)
			}
		}
		for r := 0; r < 4; r++ {
			listed := false
			for _, lr := range s.Rots {
				if lr == r {
					listed = true
					break
				}
			}
			ok := e.canPlaceLocked("rffr", s.X, s.Y, r)
			if ok != listed {
				t.Fatalf("spot %d,%d rot %d: canPlace=%v listed=%v", s.X, s.Y, r, ok, listed)
			}
		}
	}
}
