package game

import "testing"

func TestAddBotAndPlay(t *testing.T) {
	e := NewEngine("BOTS", Settings{MaxPlayers: 3, Expansions: Expansions{Farmers: true}, Name: "t"}, 99)
	if _, err := e.AddPlayer("human", "You"); err != nil {
		t.Fatal(err)
	}
	if _, err := e.AddBot(); err != nil {
		t.Fatal(err)
	}
	st := e.Public("human")
	if len(st.Players) != 2 {
		t.Fatalf("players=%d", len(st.Players))
	}
	if !st.Players[1].IsBot {
		t.Fatal("second player should be bot")
	}
	if err := e.Start(); err != nil {
		t.Fatal(err)
	}
	spot := e.Public("human").ValidSpots[0]
	if err := e.PlaceTile("human", spot.X, spot.Y, spot.Rots[0]); err != nil {
		t.Fatal(err)
	}
	if e.Public("human").Phase == PhasePlaceMeeple {
		if err := e.SkipMeeple("human"); err != nil {
			t.Fatal(err)
		}
	}
	if !e.CurrentPlayerIsBot() {
		t.Fatal("expected bot turn")
	}
	for i := 0; i < 8; i++ {
		if !e.CurrentPlayerIsBot() {
			break
		}
		acted, err := e.BotAct()
		if err != nil {
			t.Fatal(err)
		}
		if !acted {
			t.Fatal("bot should act")
		}
	}
	if e.CurrentPlayerIsBot() {
		t.Fatal("bot should have finished its turn")
	}
}

func TestBotFillsRoom(t *testing.T) {
	e := NewEngine("FULL", Settings{MaxPlayers: 2, Name: "t"}, 1)
	e.AddPlayer("human", "You")
	if _, err := e.AddBot(); err != nil {
		t.Fatal(err)
	}
	if _, err := e.AddBot(); err == nil {
		t.Fatal("expected full room")
	}
}

func TestBotsHideSpotsFromOthers(t *testing.T) {
	e := NewEngine("HIDE", Settings{MaxPlayers: 3, Name: "t"}, 7)
	e.AddPlayer("human", "You")
	bot, err := e.AddBot()
	if err != nil {
		t.Fatal(err)
	}
	e.Start()
	if len(e.Public("human").ValidSpots) == 0 {
		t.Fatal("current player should see spots")
	}
	if len(e.Public(bot.ID).ValidSpots) != 0 {
		t.Fatal("non-current player must not see placement spots")
	}
}

func TestTwoBotsPlayToGameOver(t *testing.T) {
	e := NewEngine("DUEL", Settings{MaxPlayers: 2, Expansions: Expansions{Farmers: true}, Name: "t"}, 12345)
	if _, err := e.AddBot(); err != nil {
		t.Fatal(err)
	}
	if _, err := e.AddBot(); err != nil {
		t.Fatal(err)
	}
	if err := e.Start(); err != nil {
		t.Fatal(err)
	}
	for steps := 0; steps < 5000; steps++ {
		st := e.Public("")
		if st.Phase == PhaseGameOver {
			if len(st.WinnerIDs) == 0 {
				t.Fatal("game over without winners")
			}
			if len(st.Board) < 2 {
				t.Fatalf("board too small: %d", len(st.Board))
			}
			return
		}
		if !e.CurrentPlayerIsBot() {
			t.Fatalf("expected bot turn at step %d phase=%s", steps, st.Phase)
		}
		acted, err := e.BotAct()
		if err != nil {
			t.Fatalf("step %d: %v", steps, err)
		}
		if !acted {
			t.Fatalf("step %d: bot did not act (phase=%s)", steps, st.Phase)
		}
	}
	t.Fatal("game did not finish")
}

func TestHumanAndBotPlayManyTurns(t *testing.T) {
	e := NewEngine("MIX", Settings{MaxPlayers: 2, Expansions: Expansions{Farmers: false}, Name: "t"}, 42)
	e.AddPlayer("human", "You")
	e.AddBot()
	if err := e.Start(); err != nil {
		t.Fatal(err)
	}
	for turns := 0; turns < 80; turns++ {
		st := e.Public("human")
		if st.Phase == PhaseGameOver {
			return
		}
		if e.CurrentPlayerIsBot() {
			acted, err := e.BotAct()
			if err != nil {
				t.Fatal(err)
			}
			if !acted {
				t.Fatal("bot stalled")
			}
			continue
		}
		st = e.Public("human")
		if st.Phase == PhasePlaceTile {
			if len(st.ValidSpots) == 0 {
				t.Fatal("human has no spots")
			}
			spot := st.ValidSpots[0]
			if err := e.PlaceTile("human", spot.X, spot.Y, spot.Rots[0]); err != nil {
				t.Fatal(err)
			}
		}
		st = e.Public("human")
		if st.Phase == PhasePlaceMeeple {
			if err := e.SkipMeeple("human"); err != nil {
				t.Fatal(err)
			}
		}
	}
}
