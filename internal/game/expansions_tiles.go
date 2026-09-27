package game

func BuildCatalogWithExpansions(x Expansions) map[string]TileDef {
	cat := BuildCatalog()
	if x.River {
		for id, def := range riverTiles() {
			cat[id] = def
		}
	}
	if x.Abbot {
		for id, def := range gardenTiles() {
			cat[id] = def
		}
	}
	if x.InnsAndCathedrals {
		for id, def := range innsCathedralTiles() {
			cat[id] = def
		}
	}
	return cat
}

func BuildDeckFor(cat map[string]TileDef, x Expansions) []string {
	deck := BuildDeck(cat)
	if x.Abbot {
		for id, n := range map[string]int{"garden_a": 1, "garden_b": 1, "garden_c": 1, "garden_d": 1, "garden_e": 1} {
			if _, ok := cat[id]; !ok {
				continue
			}
			for i := 0; i < n; i++ {
				deck = append(deck, id)
			}
		}
	}
	if x.InnsAndCathedrals {
		for id, n := range map[string]int{
			"inn_ns": 2, "inn_ew": 2, "inn_curve": 2,
			"cathedral": 2,
		} {
			if _, ok := cat[id]; !ok {
				continue
			}
			for i := 0; i < n; i++ {
				deck = append(deck, id)
			}
		}
	}
	return deck
}

func BuildRiverStack(cat map[string]TileDef) []string {
	mids := []string{"river_straight", "river_straight", "river_curve", "river_curve", "river_curve", "river_city", "river_road", "river_road", "river_cloister", "river_fork"}
	out := make([]string, 0, 12)
	for _, id := range mids {
		if _, ok := cat[id]; ok {
			out = append(out, id)
		}
	}
	return out
}

func riverTiles() map[string]TileDef {
	return map[string]TileDef{
		"river_source": {
			ID: "river_source", Edges: [4]Terrain{Field, Field, Field, Field},
			River: true, RiverEdges: []Dir{S},
		},
		"river_lake": {
			ID: "river_lake", Edges: [4]Terrain{Field, Field, Field, Field},
			River: true, RiverEdges: []Dir{N},
		},
		"river_straight": {
			ID: "river_straight", Edges: [4]Terrain{Field, Field, Field, Field},
			River: true, RiverEdges: []Dir{N, S},
		},
		"river_curve": {
			ID: "river_curve", Edges: [4]Terrain{Field, Field, Field, Field},
			River: true, RiverEdges: []Dir{N, E},
		},
		"river_city": {
			ID: "river_city", Edges: [4]Terrain{City, Field, Field, Field},
			CityGroups: [][]Dir{{N}}, River: true, RiverEdges: []Dir{S, W},
		},
		"river_road": {
			ID: "river_road", Edges: [4]Terrain{Field, Road, Field, Road},
			RoadGroups: [][]Dir{{E, W}}, River: true, RiverEdges: []Dir{N, S},
		},
		"river_cloister": {
			ID: "river_cloister", Edges: [4]Terrain{Field, Field, Field, Field},
			Monastery: true, River: true, RiverEdges: []Dir{N, S},
		},
		"river_fork": {
			ID: "river_fork", Edges: [4]Terrain{Field, Field, Field, Field},
			River: true, RiverEdges: []Dir{N, E, W},
		},
	}
}

func gardenTiles() map[string]TileDef {
	return map[string]TileDef{
		"garden_a": {ID: "garden_a", Edges: [4]Terrain{Field, Field, Field, Field}, Garden: true},
		"garden_b": {ID: "garden_b", Edges: [4]Terrain{Field, Road, Field, Road}, RoadGroups: [][]Dir{{E, W}}, Garden: true},
		"garden_c": {ID: "garden_c", Edges: [4]Terrain{City, Field, Field, Field}, CityGroups: [][]Dir{{N}}, Garden: true},
		"garden_d": {ID: "garden_d", Edges: [4]Terrain{Field, Field, Road, Field}, RoadGroups: [][]Dir{{S}}, Garden: true},
		"garden_e": {ID: "garden_e", Edges: [4]Terrain{City, Field, Road, Field}, CityGroups: [][]Dir{{N}}, RoadGroups: [][]Dir{{S}}, Garden: true},
	}
}

func innsCathedralTiles() map[string]TileDef {
	return map[string]TileDef{
		"inn_ns": {
			ID: "inn_ns", Edges: [4]Terrain{Road, Field, Road, Field},
			RoadGroups: [][]Dir{{N, S}}, Inn: true,
		},
		"inn_ew": {
			ID: "inn_ew", Edges: [4]Terrain{Field, Road, Field, Road},
			RoadGroups: [][]Dir{{E, W}}, Inn: true,
		},
		"inn_curve": {
			ID: "inn_curve", Edges: [4]Terrain{Road, Field, Field, Road},
			RoadGroups: [][]Dir{{N, W}}, Inn: true,
		},
		"cathedral": {
			ID: "cathedral", Edges: [4]Terrain{City, City, City, City},
			CityGroups: [][]Dir{{N, E, S, W}}, Cathedral: true,
		},
	}
}

func RotatedRiverEdges(def TileDef, rotation int) []Dir {
	out := make([]Dir, len(def.RiverEdges))
	for i, d := range def.RiverEdges {
		out[i] = RotateDir(d, rotation)
	}
	return out
}
