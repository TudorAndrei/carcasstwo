package game

import "testing"

func scoringEngine(farmers bool) *Engine {
	e := NewEngine("SCORE", Settings{MaxPlayers: 5, Expansions: Expansions{Farmers: farmers}, Name: "t"}, 1)
	e.AddPlayer("p1", "A")
	e.AddPlayer("p2", "B")
	e.Phase = PhasePlaceTile
	e.Turn = 0
	e.Deck = nil
	e.Current = nil
	e.Meeples = nil
	e.LastPlaced = nil
	e.Players[0].Meeple = 7
	e.Players[1].Meeple = 7
	e.Players[0].Score = 0
	e.Players[1].Score = 0
	e.Board = map[string]PlacedTile{}
	return e
}

func place(e *Engine, defID string, x, y, rot int) {
	e.Board[key(x, y)] = PlacedTile{DefID: defID, X: x, Y: y, Rotation: rot}
}

func TestScoreCompletedThreeTileRoad(t *testing.T) {
	e := scoringEngine(false)
	place(e, "rrrf", 0, -1, 0)
	place(e, "rfrf", 0, 0, 0)
	place(e, "rrrf", 0, 1, 0)
	e.Meeples = []Meeple{{PlayerID: "p1", X: 0, Y: 0, Kind: FeatRoad, Slot: 0}}
	e.Players[0].Meeple = 6
	e.LastPlaced = &PlacedTile{DefID: "rrrf", X: 0, Y: 1, Rotation: 0}

	segs := e.collectFeature(0, 0, FeatRoad, 0)
	if !e.roadComplete(segs) {
		t.Fatal("expected 3-tile road closed by junctions to be complete")
	}
	e.scoreAfterPlacementLocked()
	if e.Players[0].Score != 3 {
		t.Fatalf("completed 3-tile road: want 3 points, got %d", e.Players[0].Score)
	}
	if e.Players[0].Meeple != 7 {
		t.Fatalf("meeple should return after road score, supply=%d", e.Players[0].Meeple)
	}
	if len(e.Meeples) != 0 {
		t.Fatalf("scored road meeple should be removed, got %v", e.Meeples)
	}
}

func TestScoreCompletedCityWithShield(t *testing.T) {
	e := scoringEngine(false)
	e.Catalog["cfff_s"] = func() TileDef {
		td := T("CFFF", [][]Dir{{N}}, nil, false, 1)
		td.ID = "cfff_s"
		return td
	}()
	place(e, "cfff", 0, 0, 0)
	place(e, "cfff_s", 0, -1, 2)
	e.Meeples = []Meeple{{PlayerID: "p1", X: 0, Y: 0, Kind: FeatCity, Slot: 0}}
	e.Players[0].Meeple = 6
	e.LastPlaced = &PlacedTile{DefID: "cfff_s", X: 0, Y: -1, Rotation: 2}

	segs := e.collectFeature(0, 0, FeatCity, 0)
	if !e.cityComplete(segs) {
		t.Fatal("expected 2-tile city to be complete")
	}
	e.scoreAfterPlacementLocked()
	want := 2*2 + 1*2
	if e.Players[0].Score != want {
		t.Fatalf("completed city with shield: want %d (2*tiles+2*shields), got %d", want, e.Players[0].Score)
	}
	if e.Players[0].Meeple != 7 {
		t.Fatalf("meeple should return after city score, supply=%d", e.Players[0].Meeple)
	}
}

func TestScoreCompletedMonasteryNine(t *testing.T) {
	e := scoringEngine(false)
	place(e, "ffff_m", 0, 0, 0)
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			place(e, "ffff_m", dx, dy, 0)
		}
	}
	e.Meeples = []Meeple{{PlayerID: "p1", X: 0, Y: 0, Kind: FeatMonastery, Slot: 0}}
	e.Players[0].Meeple = 6
	e.LastPlaced = &PlacedTile{DefID: "ffff_m", X: 1, Y: 1, Rotation: 0}

	if !e.monasteryComplete(0, 0) {
		t.Fatal("monastery should be complete with 8 neighbors")
	}
	e.scoreAfterPlacementLocked()
	if e.Players[0].Score != 9 {
		t.Fatalf("completed monastery: want 9, got %d", e.Players[0].Score)
	}
	if e.Players[0].Meeple != 7 {
		t.Fatalf("meeple should return after monastery score, supply=%d", e.Players[0].Meeple)
	}
}

func TestFinalIncompleteCityHalfPoints(t *testing.T) {
	e := scoringEngine(false)
	place(e, "ccff_s", 0, 0, 0)
	place(e, "cfff", 0, -1, 2)
	e.Meeples = []Meeple{{PlayerID: "p1", X: 0, Y: 0, Kind: FeatCity, Slot: 0}}
	e.Players[0].Meeple = 6

	segs := e.collectFeature(0, 0, FeatCity, 0)
	if e.cityComplete(segs) {
		t.Fatal("city should be incomplete (east edge still open)")
	}
	e.finalScoreLocked()
	want := 2 + 1
	if e.Players[0].Score != want {
		t.Fatalf("final incomplete city: want %d (1*tiles+1*shields), got %d", want, e.Players[0].Score)
	}
}

func TestTieForMostMeeplesBothScore(t *testing.T) {
	e := scoringEngine(false)
	place(e, "cfff", 0, 0, 0)
	place(e, "cfff", 0, -1, 2)
	e.Meeples = []Meeple{
		{PlayerID: "p1", X: 0, Y: 0, Kind: FeatCity, Slot: 0},
		{PlayerID: "p2", X: 0, Y: -1, Kind: FeatCity, Slot: 0},
	}
	e.Players[0].Meeple = 6
	e.Players[1].Meeple = 6
	e.LastPlaced = &PlacedTile{DefID: "cfff", X: 0, Y: -1, Rotation: 2}

	e.scoreAfterPlacementLocked()
	want := 4
	if e.Players[0].Score != want || e.Players[1].Score != want {
		t.Fatalf("tie should award full points to both: p1=%d p2=%d want %d each",
			e.Players[0].Score, e.Players[1].Score, want)
	}
}

func TestFinalFarmerScoresCompletedCities(t *testing.T) {
	e := scoringEngine(true)
	place(e, "cfff", 0, 0, 0)
	place(e, "cfff", 0, -1, 2)
	place(e, "ffff_m", 1, 0, 0)
	e.Meeples = []Meeple{{PlayerID: "p1", X: 1, Y: 0, Kind: FeatField, Slot: 0}}
	e.Players[0].Meeple = 6

	citySegs := e.collectFeature(0, 0, FeatCity, 0)
	if !e.cityComplete(citySegs) {
		t.Fatal("setup city must be complete for farmer scoring")
	}
	e.finalScoreLocked()
	if e.Players[0].Score != 3 {
		t.Fatalf("farmer: want 3 per completed adjacent city, got %d", e.Players[0].Score)
	}
}

func TestRoadCompleteRequiresClosedEndsOnStraightSegments(t *testing.T) {
	e := scoringEngine(false)
	place(e, "rfrf", 0, 0, 0)
	place(e, "rfrf", 0, 1, 0)
	segs := e.collectFeature(0, 0, FeatRoad, 0)
	if e.roadComplete(segs) {
		t.Fatal("incomplete straight road with open ends must not be complete")
	}
}

func TestFinalIncompleteRoadOnePerTile(t *testing.T) {
	e := scoringEngine(false)
	place(e, "rfrf", 0, 0, 0)
	place(e, "rfrf", 0, 1, 0)
	e.Meeples = []Meeple{{PlayerID: "p1", X: 0, Y: 0, Kind: FeatRoad, Slot: 0}}
	e.finalScoreLocked()
	if e.Players[0].Score != 2 {
		t.Fatalf("final incomplete road: want 2, got %d", e.Players[0].Score)
	}
}

func TestFinalIncompleteMonasteryPartial(t *testing.T) {
	e := scoringEngine(false)
	place(e, "ffff_m", 0, 0, 0)
	place(e, "ffff_m", 1, 0, 0)
	place(e, "ffff_m", 0, 1, 0)
	e.Meeples = []Meeple{{PlayerID: "p1", X: 0, Y: 0, Kind: FeatMonastery, Slot: 0}}
	e.finalScoreLocked()
	if e.Players[0].Score != 3 {
		t.Fatalf("final incomplete monastery with 2 neighbors: want 3, got %d", e.Players[0].Score)
	}
}

func TestMajorityMeeplesOnlyWinnerScores(t *testing.T) {
	e := scoringEngine(false)
	owners := map[string]int{"p1": 2, "p2": 1}
	e.award(owners, 4)
	if e.Players[0].Score != 4 {
		t.Fatalf("majority holder want 4, got %d", e.Players[0].Score)
	}
	if e.Players[1].Score != 0 {
		t.Fatalf("minority holder want 0, got %d", e.Players[1].Score)
	}
}

func TestTileCatalogRoadGroupsMatchEdges(t *testing.T) {
	cat := BuildCatalog()
	var mismatches []string
	for id, def := range cat {
		for _, g := range def.RoadGroups {
			for _, d := range g {
				if def.Edges[d] != Road {
					mismatches = append(mismatches, id+": road group dir not Road on edge")
				}
			}
		}
		for d := Dir(0); d < 4; d++ {
			if def.Edges[d] != Road {
				continue
			}
			found := false
			for _, g := range def.RoadGroups {
				for _, ed := range g {
					if ed == d {
						found = true
					}
				}
			}
			if !found {
				mismatches = append(mismatches, id+": Road edge orphan not in RoadGroups")
			}
		}
	}
	if len(mismatches) > 0 {
		t.Fatalf("tile catalog road/edge mismatches: %v", mismatches)
	}
}

func TestRoadLoopCompletes(t *testing.T) {
	e := scoringEngine(false)
	place(e, "rffr", 0, 0, 2)
	place(e, "rffr", 1, 0, 3)
	place(e, "rffr", 1, 1, 0)
	place(e, "rffr", 0, 1, 1)
	e.Meeples = []Meeple{{PlayerID: "p1", X: 0, Y: 0, Kind: FeatRoad, Slot: 0}}
	e.Players[0].Meeple = 6
	e.LastPlaced = &PlacedTile{DefID: "rffr", X: 0, Y: 1, Rotation: 1}
	segs := e.collectFeature(0, 0, FeatRoad, 0)
	if !e.roadComplete(segs) {
		t.Fatal("4-curve road loop should be complete")
	}
	e.scoreAfterPlacementLocked()
	if e.Players[0].Score != 4 {
		t.Fatalf("loop road: want 4, got %d", e.Players[0].Score)
	}
}

func TestFarmerIgnoresIncompleteCity(t *testing.T) {
	e := scoringEngine(true)
	place(e, "cfff", 0, 0, 0)
	place(e, "ffff_m", 1, 0, 0)
	e.Meeples = []Meeple{{PlayerID: "p1", X: 1, Y: 0, Kind: FeatField, Slot: 0}}
	e.finalScoreLocked()
	if e.Players[0].Score != 0 {
		t.Fatalf("farmer must not score incomplete city, got %d", e.Players[0].Score)
	}
}

func TestFarmerTwoSeparatedCitiesOnSameTile(t *testing.T) {
	e := scoringEngine(true)
	place(e, "ccff_sep", 0, 0, 0)
	place(e, "cfff", 0, -1, 2)
	place(e, "cfff", 1, 0, 3)
	place(e, "ffff_m", 0, 1, 0)
	e.Meeples = []Meeple{{PlayerID: "p1", X: 0, Y: 1, Kind: FeatField, Slot: 0}}
	e.Players[0].Meeple = 6
	n := e.completedCitiesTouchingField(e.collectFeature(0, 1, FeatField, 0))
	if n != 2 {
		t.Fatalf("field touching two completed sep cities on one tile: want 2 cities, got %d", n)
	}
	e.finalScoreLocked()
	if e.Players[0].Score != 6 {
		t.Fatalf("farmer 2 completed cities: want 6, got %d", e.Players[0].Score)
	}
}

func TestSepShieldOnlyOnPennantSegment(t *testing.T) {
	e := scoringEngine(false)
	place(e, "ccff_sep_s", 0, 0, 0)
	place(e, "cfff", 1, 0, 3)
	e.Meeples = []Meeple{{PlayerID: "p1", X: 0, Y: 0, Kind: FeatCity, Slot: 1}}
	e.LastPlaced = &PlacedTile{DefID: "cfff", X: 1, Y: 0, Rotation: 3}
	e.scoreAfterPlacementLocked()
	if e.Players[0].Score != 4 {
		t.Fatalf("east sep city without pennant: want 4, got %d", e.Players[0].Score)
	}
}

func TestCurveTileFieldIsSingleRegion(t *testing.T) {
	e := scoringEngine(true)
	place(e, "rffr", 0, 0, 0)
	n := e.countFieldRegions(e.Board[key(0, 0)])
	if n != 1 {
		t.Fatalf("required: curve tile has one continuous field; actual countFieldRegions=%d (features.go:467-470 splits by field edge count when roads present)", n)
	}
}
