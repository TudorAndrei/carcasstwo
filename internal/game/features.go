package game

type segmentRef struct {
	x, y int
	kind FeatureKind
	slot int
}

func (e *Engine) meepleSlotsLocked(pt PlacedTile) []MeepleSlot {
	def := e.Catalog[pt.DefID]
	slots := []MeepleSlot{}
	cities := RotatedCityGroups(def, pt.Rotation)
	for i := range cities {
		if !e.featureHasMeepleLocked(pt.X, pt.Y, FeatCity, i) {
			slots = append(slots, MeepleSlot{Kind: FeatCity, Slot: i, Label: "City"})
		}
	}
	roads := RotatedRoadGroups(def, pt.Rotation)
	for i := range roads {
		if !e.featureHasMeepleLocked(pt.X, pt.Y, FeatRoad, i) {
			slots = append(slots, MeepleSlot{Kind: FeatRoad, Slot: i, Label: "Road"})
		}
	}
	if def.Monastery {
		slots = append(slots, MeepleSlot{Kind: FeatMonastery, Slot: 0, Label: "Monastery"})
		if e.Settings.Expansions.Abbot {
			slots = append(slots, MeepleSlot{Kind: FeatMonastery, Slot: 0, Label: "Abbot", Abbot: true})
		}
	}
	if def.Garden && e.Settings.Expansions.Abbot {
		slots = append(slots, MeepleSlot{Kind: FeatGarden, Slot: 0, Label: "Abbot (garden)", Abbot: true})
	}
	if e.Settings.Expansions.Farmers {
		fieldSlots := e.fieldSlotsOnTile(pt)
		for _, fs := range fieldSlots {
			if !e.featureHasMeepleLocked(pt.X, pt.Y, FeatField, fs) {
				slots = append(slots, MeepleSlot{Kind: FeatField, Slot: fs, Label: "Field"})
			}
		}
	}
	return slots
}

func (e *Engine) featureHasMeepleLocked(x, y int, kind FeatureKind, slot int) bool {
	visited := map[string]bool{}
	queue := []segmentRef{{x, y, kind, slot}}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		vk := fmtSeg(cur)
		if visited[vk] {
			continue
		}
		visited[vk] = true
		for _, m := range e.Meeples {
			if m.X == cur.x && m.Y == cur.y && m.Kind == cur.kind && m.Slot == cur.slot {
				return true
			}
		}
		queue = append(queue, e.neighborsOfSegment(cur)...)
	}
	return false
}

func fmtSeg(s segmentRef) string {
	return key(s.x, s.y) + ":" + string(s.kind) + ":" + itoa(s.slot)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		return "-" + string(digits)
	}
	return string(digits)
}

func (e *Engine) neighborsOfSegment(s segmentRef) []segmentRef {
	pt, ok := e.Board[key(s.x, s.y)]
	if !ok {
		return nil
	}
	def := e.Catalog[pt.DefID]
	out := []segmentRef{}
	switch s.kind {
	case FeatCity:
		groups := RotatedCityGroups(def, pt.Rotation)
		if s.slot < 0 || s.slot >= len(groups) {
			return nil
		}
		for _, d := range groups[s.slot] {
			dx, dy := d.Delta()
			nx, ny := s.x+dx, s.y+dy
			nbt, ok := e.Board[key(nx, ny)]
			if !ok {
				continue
			}
			nbDef := e.Catalog[nbt.DefID]
			nbGroups := RotatedCityGroups(nbDef, nbt.Rotation)
			opp := d.Opposite()
			for i, g := range nbGroups {
				for _, ed := range g {
					if ed == opp {
						out = append(out, segmentRef{nx, ny, FeatCity, i})
					}
				}
			}
		}
	case FeatRoad:
		groups := RotatedRoadGroups(def, pt.Rotation)
		if s.slot < 0 || s.slot >= len(groups) {
			return nil
		}
		for _, d := range groups[s.slot] {
			dx, dy := d.Delta()
			nx, ny := s.x+dx, s.y+dy
			nbt, ok := e.Board[key(nx, ny)]
			if !ok {
				continue
			}
			nbDef := e.Catalog[nbt.DefID]
			nbGroups := RotatedRoadGroups(nbDef, nbt.Rotation)
			opp := d.Opposite()
			for i, g := range nbGroups {
				for _, ed := range g {
					if ed == opp {
						out = append(out, segmentRef{nx, ny, FeatRoad, i})
					}
				}
			}
		}
	case FeatField:
		for _, n := range e.fieldNeighbors(s) {
			out = append(out, n)
		}
	}
	return out
}

func (e *Engine) collectFeature(x, y int, kind FeatureKind, slot int) []segmentRef {
	visited := map[string]bool{}
	var segs []segmentRef
	queue := []segmentRef{{x, y, kind, slot}}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		vk := fmtSeg(cur)
		if visited[vk] {
			continue
		}
		visited[vk] = true
		segs = append(segs, cur)
		queue = append(queue, e.neighborsOfSegment(cur)...)
	}
	return segs
}

func (e *Engine) scoreAfterPlacementLocked() {
	if e.LastPlaced == nil {
		return
	}
	pt := *e.LastPlaced
	def := e.Catalog[pt.DefID]

	cities := RotatedCityGroups(def, pt.Rotation)
	for i := range cities {
		segs := e.collectFeature(pt.X, pt.Y, FeatCity, i)
		if e.cityComplete(segs) {
			e.scoreFeature(segs, FeatCity, true)
		}
	}
	roads := RotatedRoadGroups(def, pt.Rotation)
	for i := range roads {
		segs := e.collectFeature(pt.X, pt.Y, FeatRoad, i)
		if e.roadComplete(segs) {
			e.scoreFeature(segs, FeatRoad, true)
		}
	}

	for y := pt.Y - 1; y <= pt.Y+1; y++ {
		for x := pt.X - 1; x <= pt.X+1; x++ {
			t, ok := e.Board[key(x, y)]
			if !ok {
				continue
			}
			td := e.Catalog[t.DefID]
			if td.Monastery || td.Garden {
				if e.monasteryComplete(x, y) {
					e.scoreMonastery(x, y, true)
				}
			}
		}
	}
}

func (e *Engine) cityComplete(segs []segmentRef) bool {
	for _, s := range segs {
		pt := e.Board[key(s.x, s.y)]
		def := e.Catalog[pt.DefID]
		groups := RotatedCityGroups(def, pt.Rotation)
		if s.slot >= len(groups) {
			return false
		}
		for _, d := range groups[s.slot] {
			dx, dy := d.Delta()
			if _, ok := e.Board[key(s.x+dx, s.y+dy)]; !ok {
				return false
			}
		}
	}
	return true
}

func (e *Engine) roadComplete(segs []segmentRef) bool {
	for _, s := range segs {
		pt := e.Board[key(s.x, s.y)]
		def := e.Catalog[pt.DefID]
		groups := RotatedRoadGroups(def, pt.Rotation)
		if s.slot < 0 || s.slot >= len(groups) {
			return false
		}
		for _, d := range groups[s.slot] {
			dx, dy := d.Delta()
			if _, ok := e.Board[key(s.x+dx, s.y+dy)]; !ok {
				return false
			}
		}
	}
	return len(segs) > 0
}

func (e *Engine) monasteryComplete(x, y int) bool {
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			if _, ok := e.Board[key(x+dx, y+dy)]; !ok {
				return false
			}
		}
	}
	return true
}

func (e *Engine) monasteryScore(x, y int) int {
	n := 1
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			if _, ok := e.Board[key(x+dx, y+dy)]; ok {
				n++
			}
		}
	}
	return n
}

func (e *Engine) scoreMonastery(x, y int, remove bool) {
	points := e.monasteryScore(x, y)
	owners := map[string]int{}
	keep := e.Meeples[:0]
	for _, m := range e.Meeples {
		if m.X == x && m.Y == y && (m.Kind == FeatMonastery || m.Kind == FeatGarden) {
			owners[m.PlayerID]++
			if remove {
				e.returnFollower(m)
				continue
			}
		}
		keep = append(keep, m)
	}
	if remove {
		e.Meeples = keep
	}
	e.award(owners, points)
}

func (e *Engine) scoreFeature(segs []segmentRef, kind FeatureKind, remove bool) {
	tileSet := map[string]bool{}
	shields := 0
	for _, s := range segs {
		k := key(s.x, s.y)
		if tileSet[k] {
			continue
		}
		tileSet[k] = true
		if kind == FeatCity {
			pt := e.Board[k]
			shields += e.Catalog[pt.DefID].ShieldsForCitySlot(s.slot)
		}
	}
	points := len(tileSet)
	if kind == FeatCity {
		cathedral := false
		for _, s := range segs {
			if e.Catalog[e.Board[key(s.x, s.y)].DefID].Cathedral {
				cathedral = true
				break
			}
		}
		if cathedral {
			points = len(tileSet)*3 + shields*2
		} else {
			points = len(tileSet)*2 + shields*2
		}
	}
	if kind == FeatRoad {
		inn := false
		for _, s := range segs {
			if e.Catalog[e.Board[key(s.x, s.y)].DefID].Inn {
				inn = true
				break
			}
		}
		if inn {
			points = len(tileSet) * 2
		}
	}
	owners := map[string]int{}
	segSet := map[string]bool{}
	for _, s := range segs {
		segSet[fmtSeg(s)] = true
	}
	keep := e.Meeples[:0]
	for _, m := range e.Meeples {
		if m.Kind == kind && segSet[fmtSeg(segmentRef{m.X, m.Y, m.Kind, m.Slot})] {
			owners[m.PlayerID]++
			if remove {
				e.returnFollower(m)
				continue
			}
		}
		keep = append(keep, m)
	}
	if remove {
		e.Meeples = keep
	}
	e.award(owners, points)
}

func (e *Engine) award(owners map[string]int, points int) {
	if len(owners) == 0 || points == 0 {
		return
	}
	max := 0
	for _, n := range owners {
		if n > max {
			max = n
		}
	}
	for id, n := range owners {
		if n == max {
			for i := range e.Players {
				if e.Players[i].ID == id {
					e.Players[i].Score += points
				}
			}
		}
	}
}

func (e *Engine) returnFollower(m Meeple) {
	for i := range e.Players {
		if e.Players[i].ID == m.PlayerID {
			if m.IsAbbot {
				e.Players[i].Abbot++
			} else {
				e.Players[i].Meeple++
			}
			return
		}
	}
}

func (e *Engine) returnMeeple(playerID string) {
	e.returnFollower(Meeple{PlayerID: playerID})
}

func (e *Engine) finalScoreLocked() {
	seenCity := map[string]bool{}
	seenRoad := map[string]bool{}
	seenMon := map[string]bool{}
	seenField := map[string]bool{}

	for _, m := range append([]Meeple{}, e.Meeples...) {
		switch m.Kind {
		case FeatCity:
			sk := fmtSeg(segmentRef{m.X, m.Y, m.Kind, m.Slot})
			if seenCity[sk] {
				continue
			}
			segs := e.collectFeature(m.X, m.Y, FeatCity, m.Slot)
			for _, s := range segs {
				seenCity[fmtSeg(s)] = true
			}
			tileSet := map[string]bool{}
			shields := 0
			for _, s := range segs {
				k := key(s.x, s.y)
				if tileSet[k] {
					continue
				}
				tileSet[k] = true
				shields += e.Catalog[e.Board[k].DefID].ShieldsForCitySlot(s.slot)
			}
			points := len(tileSet) + shields
			for _, s := range segs {
				if e.Catalog[e.Board[key(s.x, s.y)].DefID].Cathedral {
					points = 0
					break
				}
			}
			owners := e.ownersOn(segs)
			e.award(owners, points)
		case FeatRoad:
			sk := fmtSeg(segmentRef{m.X, m.Y, m.Kind, m.Slot})
			if seenRoad[sk] {
				continue
			}
			segs := e.collectFeature(m.X, m.Y, FeatRoad, m.Slot)
			for _, s := range segs {
				seenRoad[fmtSeg(s)] = true
			}
			tileSet := map[string]bool{}
			inn := false
			for _, s := range segs {
				tileSet[key(s.x, s.y)] = true
				if e.Catalog[e.Board[key(s.x, s.y)].DefID].Inn {
					inn = true
				}
			}
			pts := len(tileSet)
			if inn {
				pts = 0
			}
			e.award(e.ownersOn(segs), pts)
		case FeatMonastery, FeatGarden:
			k := key(m.X, m.Y)
			if seenMon[k] {
				continue
			}
			seenMon[k] = true
			e.scoreMonastery(m.X, m.Y, false)
		case FeatField:
			if !e.Settings.Expansions.Farmers {
				continue
			}
			sk := fmtSeg(segmentRef{m.X, m.Y, m.Kind, m.Slot})
			if seenField[sk] {
				continue
			}
			segs := e.collectFeature(m.X, m.Y, FeatField, m.Slot)
			for _, s := range segs {
				seenField[fmtSeg(s)] = true
			}
			cities := e.completedCitiesTouchingField(segs)
			e.award(e.ownersOn(segs), cities*3)
		}
	}
}

func (e *Engine) ownersOn(segs []segmentRef) map[string]int {
	segSet := map[string]bool{}
	for _, s := range segs {
		segSet[fmtSeg(s)] = true
	}
	owners := map[string]int{}
	for _, m := range e.Meeples {
		if segSet[fmtSeg(segmentRef{m.X, m.Y, m.Kind, m.Slot})] {
			owners[m.PlayerID]++
		}
	}
	return owners
}

func (e *Engine) fieldSlotsOnTile(pt PlacedTile) []int {
	def := e.Catalog[pt.DefID]
	n := 0
	for d := Dir(0); d < 4; d++ {
		if EdgeTerrain(def, pt.Rotation, d) == Field {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	slots := []int{}
	for i := 0; i < e.countFieldRegions(pt); i++ {
		slots = append(slots, i)
	}
	return slots
}

func (e *Engine) countFieldRegions(pt PlacedTile) int {
	def := e.Catalog[pt.DefID]
	fieldEdges := 0
	for d := Dir(0); d < 4; d++ {
		if EdgeTerrain(def, pt.Rotation, d) == Field {
			fieldEdges++
		}
	}
	if fieldEdges == 0 {
		return 0
	}
	roads := RotatedRoadGroups(def, pt.Rotation)
	if len(roads) == 0 {
		return 1
	}
	if len(roads) == 1 && len(roads[0]) == 2 {
		a, b := roads[0][0], roads[0][1]
		if (int(a)+2)%4 == int(b) {
			return 2
		}
		return 1
	}
	ends := 0
	for _, g := range roads {
		ends += len(g)
	}
	if ends > 0 {
		return ends
	}
	return 1
}

func (e *Engine) fieldNeighbors(s segmentRef) []segmentRef {
	pt, ok := e.Board[key(s.x, s.y)]
	if !ok {
		return nil
	}
	edges := e.fieldEdgesForSlot(pt, s.slot)
	out := []segmentRef{}
	for _, d := range edges {
		dx, dy := d.Delta()
		nx, ny := s.x+dx, s.y+dy
		nbt, ok := e.Board[key(nx, ny)]
		if !ok {
			continue
		}
		opp := d.Opposite()
		if EdgeTerrain(e.Catalog[nbt.DefID], nbt.Rotation, opp) != Field {
			continue
		}
		for slot := 0; slot < e.countFieldRegions(nbt); slot++ {
			nbEdges := e.fieldEdgesForSlot(nbt, slot)
			for _, ed := range nbEdges {
				if ed == opp {
					out = append(out, segmentRef{nx, ny, FeatField, slot})
				}
			}
		}
	}
	return out
}

func (e *Engine) fieldEdgesForSlot(pt PlacedTile, slot int) []Dir {
	def := e.Catalog[pt.DefID]
	fieldEdges := []Dir{}
	for d := Dir(0); d < 4; d++ {
		if EdgeTerrain(def, pt.Rotation, d) == Field {
			fieldEdges = append(fieldEdges, d)
		}
	}
	regions := e.countFieldRegions(pt)
	if regions <= 1 {
		return fieldEdges
	}
	if slot >= 0 && slot < len(fieldEdges) {
		return []Dir{fieldEdges[slot]}
	}
	return nil
}

func (e *Engine) completedCitiesTouchingField(fieldSegs []segmentRef) int {
	cityKeys := map[string]bool{}
	count := 0
	seenCityFeat := map[string]bool{}
	for _, fs := range fieldSegs {
		pt := e.Board[key(fs.x, fs.y)]
		def := e.Catalog[pt.DefID]
		cities := RotatedCityGroups(def, pt.Rotation)
		for i := range cities {
			ck := key(fs.x, fs.y) + ":c:" + itoa(i)
			if seenCityFeat[ck] {
				continue
			}
			segs := e.collectFeature(fs.x, fs.y, FeatCity, i)
			if !e.cityComplete(segs) {
				continue
			}
			rep := ""
			for _, s := range segs {
				k := key(s.x, s.y)
				if rep == "" || k < rep {
					rep = k
				}
			}
			rep += ":" + itoa(len(segs))
			if cityKeys[rep] {
				continue
			}
			cityKeys[rep] = true
			for _, s := range segs {
				seenCityFeat[key(s.x, s.y)+":c:"+itoa(s.slot)] = true
			}
			count++
		}
	}
	return count
}
