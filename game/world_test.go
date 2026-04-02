package game

import "testing"

func TestLoadWorld(t *testing.T) {
	w := LoadWorld()
	if len(w) != 9 {
		t.Fatalf("expected 9 rooms, got %d", len(w))
	}
	id := RoomID{1, 1}
	rd, ok := w[id]
	if !ok {
		t.Fatal("missing start room")
	}
	if len(rd.Tiles) != RoomInnerH || len(rd.Tiles[0]) != RoomInnerW {
		t.Fatalf("bad dimensions %dx%d", len(rd.Tiles[0]), len(rd.Tiles))
	}
}

func TestLoadDungeon(t *testing.T) {
	d := LoadDungeon()
	if len(d) != 3 {
		t.Fatalf("expected 3 dungeon rooms, got %d", len(d))
	}
	for _, id := range []RoomID{{0, 0}, {1, 0}, {2, 0}} {
		rd, ok := d[id]
		if !ok {
			t.Fatalf("missing dungeon room %v", id)
		}
		if len(rd.Tiles) != RoomInnerH || len(rd.Tiles[0]) != RoomInnerW {
			t.Fatalf("dungeon %v bad size", id)
		}
	}
}
