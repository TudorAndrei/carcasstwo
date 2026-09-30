package game

type Terrain int

const (
	Field Terrain = iota
	Road
	City
)

func (t Terrain) String() string {
	switch t {
	case Field:
		return "field"
	case Road:
		return "road"
	case City:
		return "city"
	default:
		return "?"
	}
}

type Dir int

const (
	N Dir = iota
	E
	S
	W
)

func (d Dir) Opposite() Dir {
	return Dir((int(d) + 2) % 4)
}

func (d Dir) Delta() (int, int) {
	switch d {
	case N:
		return 0, -1
	case E:
		return 1, 0
	case S:
		return 0, 1
	case W:
		return -1, 0
	}
	return 0, 0
}

type FeatureKind string

const (
	FeatRoad      FeatureKind = "road"
	FeatCity      FeatureKind = "city"
	FeatMonastery FeatureKind = "monastery"
	FeatGarden    FeatureKind = "garden"
	FeatField     FeatureKind = "field"
)

type PlayerColor string

var Colors = []PlayerColor{"red", "blue", "green", "yellow", "black", "gray"}

type Player struct {
	ID     string      `json:"id"`
	Name   string      `json:"name"`
	Color  PlayerColor `json:"color"`
	Score  int         `json:"score"`
	Meeple int         `json:"meeple"`
	Abbot  int         `json:"abbot"`
	IsBot  bool        `json:"isBot"`
}

type TileDef struct {
	ID          string     `json:"id"`
	Edges       [4]Terrain `json:"edges"`
	Monastery   bool       `json:"monastery"`
	Garden      bool       `json:"garden"`
	Inn         bool       `json:"inn"`
	Cathedral   bool       `json:"cathedral"`
	River       bool       `json:"river"`
	RiverEdges  []Dir      `json:"riverEdges"`
	Shields     int        `json:"shields"`
	CityShields []int      `json:"cityShields"`
	CityGroups  [][]Dir    `json:"cityGroups"`
	RoadGroups  [][]Dir    `json:"roadGroups"`
}

func (def TileDef) ShieldsForCitySlot(slot int) int {
	if slot >= 0 && slot < len(def.CityShields) {
		return def.CityShields[slot]
	}
	if len(def.CityGroups) <= 1 && slot == 0 {
		return def.Shields
	}
	return 0
}

type PlacedTile struct {
	DefID    string `json:"defId"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
	Rotation int    `json:"rotation"`
}

type Meeple struct {
	PlayerID string      `json:"playerId"`
	X        int         `json:"x"`
	Y        int         `json:"y"`
	Kind     FeatureKind `json:"kind"`
	Slot     int         `json:"slot"`
	IsAbbot  bool        `json:"isAbbot"`
}

type Phase string

const (
	PhaseLobby        Phase = "lobby"
	PhasePlaceTile    Phase = "placeTile"
	PhasePlaceMeeple  Phase = "placeMeeple"
	PhaseGameOver     Phase = "gameOver"
)

type Settings struct {
	MaxPlayers int        `json:"maxPlayers"`
	Private    bool       `json:"private"`
	Name       string     `json:"name"`
	Expansions Expansions `json:"expansions"`
}

func (s Settings) FarmersEnabled() bool {
	return s.Expansions.Farmers
}

type PublicState struct {
	RoomCode       string            `json:"roomCode"`
	Settings       Settings          `json:"settings"`
	Phase          Phase             `json:"phase"`
	Players        []Player          `json:"players"`
	Turn           int               `json:"turn"`
	Board          []PlacedTile      `json:"board"`
	Meeples        []Meeple          `json:"meeples"`
	CurrentTile    *string           `json:"currentTile"`
	TilesLeft      int               `json:"tilesLeft"`
	ValidSpots     []Spot            `json:"validSpots"`
	MeepleSlots    []MeepleSlot      `json:"meepleSlots"`
	LastPlaced     *PlacedTile       `json:"lastPlaced"`
	Message        string            `json:"message"`
	WinnerIDs      []string          `json:"winnerIds"`
	TileCatalog    map[string]TileDef `json:"tileCatalog"`
	YouID          string            `json:"youId,omitempty"`
}

type Spot struct {
	X        int   `json:"x"`
	Y        int   `json:"y"`
	Rots     []int `json:"rots"`
}

type MeepleSlot struct {
	Kind  FeatureKind `json:"kind"`
	Slot  int         `json:"slot"`
	Label string      `json:"label"`
	Abbot bool        `json:"abbot"`
}
