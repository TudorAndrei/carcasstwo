package game

type Expansions struct {
	Farmers             bool `json:"farmers"`
	River               bool `json:"river"`
	Abbot               bool `json:"abbot"`
	InnsAndCathedrals   bool `json:"innsAndCathedrals"`
	TradersAndBuilders  bool `json:"tradersAndBuilders"`
	PrincessAndDragon   bool `json:"princessAndDragon"`
	Tower               bool `json:"tower"`
	AbbeyAndMayor       bool `json:"abbeyAndMayor"`
	BridgesCastlesBaz   bool `json:"bridgesCastlesBazaars"`
	HillsAndSheep       bool `json:"hillsAndSheep"`
}

func (x Expansions) Labels() []string {
	out := []string{}
	if x.Farmers {
		out = append(out, "Farmers")
	}
	if x.River {
		out = append(out, "The River")
	}
	if x.Abbot {
		out = append(out, "The Abbot")
	}
	if x.InnsAndCathedrals {
		out = append(out, "Inns & Cathedrals")
	}
	if x.TradersAndBuilders {
		out = append(out, "Traders & Builders")
	}
	if x.PrincessAndDragon {
		out = append(out, "Princess & Dragon")
	}
	if x.Tower {
		out = append(out, "The Tower")
	}
	if x.AbbeyAndMayor {
		out = append(out, "Abbey & Mayor")
	}
	if x.BridgesCastlesBaz {
		out = append(out, "Bridges, Castles & Bazaars")
	}
	if x.HillsAndSheep {
		out = append(out, "Hills & Sheep")
	}
	return out
}

func (x Expansions) MaxPlayersCap() int {
	if x.InnsAndCathedrals {
		return 6
	}
	return 5
}
