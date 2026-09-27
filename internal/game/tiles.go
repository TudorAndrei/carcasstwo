package game

func T(edges string, cityGroups [][]Dir, roadGroups [][]Dir, monastery bool, shields int) TileDef {
	var e [4]Terrain
	for i, c := range edges {
		switch c {
		case 'C':
			e[i] = City
		case 'R':
			e[i] = Road
		default:
			e[i] = Field
		}
	}
	var cityShields []int
	if shields > 0 && len(cityGroups) > 0 {
		cityShields = make([]int, len(cityGroups))
		cityShields[0] = shields
	}
	return TileDef{
		Edges:       e,
		Monastery:   monastery,
		Shields:     shields,
		CityShields: cityShields,
		CityGroups:  cityGroups,
		RoadGroups:  roadGroups,
	}
}

func BuildCatalog() map[string]TileDef {
	defs := []struct {
		id    string
		count int
		def   TileDef
	}{
		{"start", 1, T("CRFR", [][]Dir{{N}}, [][]Dir{{E, W}}, false, 0)},
		{"ffff_m", 4, T("FFFF", nil, nil, true, 0)},
		{"frfr_m", 2, T("RFFF", nil, [][]Dir{{N}}, true, 0)},
		{"cccc", 1, T("CCCC", [][]Dir{{N, E, S, W}}, nil, false, 0)},
		{"cccf_s", 3, T("CCCF", [][]Dir{{N, E, S}}, nil, false, 1)},
		{"cccf", 1, T("CCCF", [][]Dir{{N, E, S}}, nil, false, 0)},
		{"cccr_s", 2, T("CCCR", [][]Dir{{N, E, S}}, [][]Dir{{W}}, false, 1)},
		{"cccr", 3, T("CCCR", [][]Dir{{N, E, S}}, [][]Dir{{W}}, false, 0)},
		{"ccff_s", 2, T("CCFF", [][]Dir{{N, E}}, nil, false, 1)},
		{"ccff", 3, T("CCFF", [][]Dir{{N, E}}, nil, false, 0)},
		{"ccff_sep_s", 2, T("CCFF", [][]Dir{{N}, {E}}, nil, false, 1)},
		{"ccff_sep", 2, T("CCFF", [][]Dir{{N}, {E}}, nil, false, 0)},
		{"cfcf_s", 2, T("CFCF", [][]Dir{{N, S}}, nil, false, 1)},
		{"cfcf", 1, T("CFCF", [][]Dir{{N, S}}, nil, false, 0)},
		{"cfcf_sep", 3, T("CFCF", [][]Dir{{N}, {S}}, nil, false, 0)},
		{"cffc", 2, T("CFFC", [][]Dir{{N, W}}, nil, false, 0)},
		{"cfff", 5, T("CFFF", [][]Dir{{N}}, nil, false, 0)},
		{"cfrf", 3, T("CRFR", [][]Dir{{N}}, [][]Dir{{E, W}}, false, 0)},
		{"cfrr", 3, T("CFRR", [][]Dir{{N}}, [][]Dir{{S}, {W}}, false, 0)},
		{"crrf", 3, T("CRRF", [][]Dir{{N}}, [][]Dir{{E}, {S}}, false, 0)},
		{"crrr", 3, T("CRRR", [][]Dir{{N}}, [][]Dir{{E}, {S}, {W}}, false, 0)},
		{"rfrf", 8, T("RFRF", nil, [][]Dir{{N, S}}, false, 0)},
		{"rffr", 9, T("RFFR", nil, [][]Dir{{N, W}}, false, 0)},
		{"rrrf", 4, T("RRRF", nil, [][]Dir{{N}, {E}, {S}}, false, 0)},
		{"rrrr", 1, T("RRRR", nil, [][]Dir{{N}, {E}, {S}, {W}}, false, 0)},
		{"ccrr_s", 2, T("CCRR", [][]Dir{{N, E}}, [][]Dir{{S, W}}, false, 1)},
		{"ccrr", 3, T("CCRR", [][]Dir{{N, E}}, [][]Dir{{S, W}}, false, 0)},
	}

	out := make(map[string]TileDef, len(defs))
	for _, d := range defs {
		td := d.def
		td.ID = d.id
		out[d.id] = td
	}
	return out
}

func BuildDeck(catalog map[string]TileDef) []string {
	counts := map[string]int{
		"ffff_m": 4, "frfr_m": 2, "cccc": 1,
		"cccf_s": 3, "cccf": 1, "cccr_s": 1, "cccr": 3,
		"ccff_s": 2, "ccff": 3, "ccff_sep_s": 2, "ccff_sep": 2,
		"cfcf_s": 1, "cfcf": 1, "cfcf_sep": 2,
		"cffc": 2, "cfff": 4, "cfrf": 3, "cfrr": 3, "crrf": 3, "crrr": 3,
		"rfrf": 7, "rffr": 8, "rrrf": 4, "rrrr": 1,
		"ccrr_s": 2, "ccrr": 3,
	}
	deck := make([]string, 0, 72)
	for id, n := range counts {
		if _, ok := catalog[id]; !ok {
			continue
		}
		for i := 0; i < n; i++ {
			deck = append(deck, id)
		}
	}
	return deck
}

func EdgeTerrain(def TileDef, rotation int, d Dir) Terrain {
	src := Dir((int(d) - rotation + 4) % 4)
	return def.Edges[src]
}

func RotateDir(d Dir, rotation int) Dir {
	return Dir((int(d) + rotation) % 4)
}

func UnrotateDir(d Dir, rotation int) Dir {
	return Dir((int(d) - rotation + 4) % 4)
}

func RotatedCityGroups(def TileDef, rotation int) [][]Dir {
	out := make([][]Dir, len(def.CityGroups))
	for i, g := range def.CityGroups {
		ng := make([]Dir, len(g))
		for j, d := range g {
			ng[j] = RotateDir(d, rotation)
		}
		out[i] = ng
	}
	return out
}

func RotatedRoadGroups(def TileDef, rotation int) [][]Dir {
	out := make([][]Dir, len(def.RoadGroups))
	for i, g := range def.RoadGroups {
		ng := make([]Dir, len(g))
		for j, d := range g {
			ng[j] = RotateDir(d, rotation)
		}
		out[i] = ng
	}
	return out
}
