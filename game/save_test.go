package game

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	s := NewState()
	s.Player.Rupees = 42
	s.Player.Hearts = 8
	s.Room = RoomID{2, 1}
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(tmp, "terminalzelda", "save.json")
	if _, err := os.Stat(p); err != nil {
		t.Fatal(err)
	}
	s2, err := LoadGame()
	if err != nil {
		t.Fatal(err)
	}
	if s2.Player.Rupees != 42 || s2.Player.Hearts != 8 {
		t.Fatalf("got rupees=%d hearts=%d", s2.Player.Rupees, s2.Player.Hearts)
	}
	if s2.Room != (RoomID{2, 1}) {
		t.Fatalf("room %+v", s2.Room)
	}
}
