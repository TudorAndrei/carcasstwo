package game

import "testing"

func TestPlaceAndScoreRoad(t *testing.T) {
	e := NewEngine("TEST", Settings{MaxPlayers: 2, Expansions: Expansions{Farmers: false}, Name: "t"}, 42)
	e.AddPlayer("p1", "A")
	e.AddPlayer("p2", "B")
	if err := e.Start(); err != nil {
		t.Fatal(err)
	}
	st := e.Public("p1")
	if st.Phase != PhasePlaceTile {
		t.Fatalf("phase %s", st.Phase)
	}
	if st.CurrentTile == nil {
		t.Fatal("no tile")
	}
	if len(st.ValidSpots) == 0 {
		t.Fatal("no valid spots")
	}
	spot := st.ValidSpots[0]
	if err := e.PlaceTile("p1", spot.X, spot.Y, spot.Rots[0]); err != nil {
		t.Fatal(err)
	}
	st = e.Public("p1")
	if st.Phase != PhasePlaceMeeple && st.Phase != PhasePlaceTile {
		t.Fatalf("unexpected phase %s", st.Phase)
	}
}

func TestInvalidOverlap(t *testing.T) {
	e := NewEngine("TEST", Settings{MaxPlayers: 2}, 1)
	e.AddPlayer("p1", "A")
	e.AddPlayer("p2", "B")
	e.Start()
	err := e.PlaceTile("p1", 0, 0, 0)
	if err == nil {
		t.Fatal("expected overlap error")
	}
}
