package handlers_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/tudor/carcasstwo/internal/game"
	"github.com/tudor/carcasstwo/internal/handlers"
	"github.com/tudor/carcasstwo/internal/hub"
	"github.com/tudor/carcasstwo/internal/lobby"
)

func setup() (*lobby.Manager, http.Handler) {
	l := lobby.NewManager()
	h := hub.New()
	s := handlers.New(l, h)
	r := chi.NewRouter()
	s.Routes(r)
	return l, r
}

func TestHTTPAddBotStartAndBotActs(t *testing.T) {
	l, r := setup()
	room := l.Create(game.Settings{MaxPlayers: 3, Name: "Solo", Expansions: game.Expansions{Farmers: true}})
	if _, err := room.Engine.AddPlayer("human-1", "Alice"); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/rooms/"+room.Code+"/add-bot", nil)
	req.AddCookie(&http.Cookie{Name: "player_id", Value: "human-1"})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("add-bot status %d body=%s", rec.Code, rec.Body.String())
	}
	st := room.Engine.Public("human-1")
	if len(st.Players) != 2 || !st.Players[1].IsBot {
		t.Fatalf("expected human+bot, got %+v", st.Players)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/rooms/"+room.Code+"/start", nil)
	req.AddCookie(&http.Cookie{Name: "player_id", Value: "human-1"})
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("start status %d body=%s", rec.Code, rec.Body.String())
	}

	st = room.Engine.Public("human-1")
	if st.Phase != game.PhasePlaceTile {
		t.Fatalf("phase %s", st.Phase)
	}
	if len(st.ValidSpots) == 0 {
		t.Fatal("no spots")
	}
	spot := st.ValidSpots[0]
	body, _ := json.Marshal(map[string]int{"x": spot.X, "y": spot.Y, "rot": spot.Rots[0]})
	req = httptest.NewRequest(http.MethodPost, "/api/rooms/"+room.Code+"/place-tile", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "player_id", Value: "human-1"})
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("place-tile status %d body=%s", rec.Code, rec.Body.String())
	}

	st = room.Engine.Public("human-1")
	if st.Phase == game.PhasePlaceMeeple {
		req = httptest.NewRequest(http.MethodPost, "/api/rooms/"+room.Code+"/skip-meeple", nil)
		req.AddCookie(&http.Cookie{Name: "player_id", Value: "human-1"})
		rec = httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != 200 {
			t.Fatalf("skip status %d body=%s", rec.Code, rec.Body.String())
		}
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		st = room.Engine.Public("human-1")
		if st.Phase == game.PhaseGameOver {
			t.Fatal("game ended too early")
		}
		if st.Players[st.Turn].ID == "human-1" && st.Phase == game.PhasePlaceTile {
			if len(st.Board) < 3 {
				t.Fatalf("bot should have placed a tile; board=%d", len(st.Board))
			}
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	st = room.Engine.Public("human-1")
	t.Fatalf("bot did not finish turn; phase=%s turn=%d board=%d msg=%s", st.Phase, st.Turn, len(st.Board), st.Message)
}

func TestHTTPAddBotRejectedWhenFull(t *testing.T) {
	l, r := setup()
	room := l.Create(game.Settings{MaxPlayers: 2, Name: "Full"})
	room.Engine.AddPlayer("human-1", "Alice")
	room.Engine.AddBot()

	req := httptest.NewRequest(http.MethodPost, "/api/rooms/"+room.Code+"/add-bot", nil)
	req.AddCookie(&http.Cookie{Name: "player_id", Value: "human-1"})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("want 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	b, _ := io.ReadAll(rec.Body)
	if !strings.Contains(string(b), "full") {
		t.Fatalf("body=%s", b)
	}
}
