package lobby

import (
	"testing"

	"github.com/tudor/carcasstwo/internal/game"
)

func TestListPublicHidesPrivateRooms(t *testing.T) {
	m := NewManager()
	pub := m.Create(game.Settings{Name: "Open", MaxPlayers: 4, Private: false})
	priv := m.Create(game.Settings{Name: "Secret", MaxPlayers: 4, Private: true})
	_ = pub
	_ = priv

	listed := m.ListPublic()
	if len(listed) != 1 {
		t.Fatalf("ListPublic len = %d, want 1", len(listed))
	}
	if listed[0].Code != pub.Code {
		t.Fatalf("listed %s, want public %s", listed[0].Code, pub.Code)
	}
	if _, ok := m.Get(priv.Code); !ok {
		t.Fatal("private room should still be joinable by code")
	}
}
