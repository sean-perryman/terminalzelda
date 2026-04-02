package game

// Realm selects which room graph the player is in.
type Realm int

const (
	RealmOverworld Realm = iota
	RealmDungeon
)

// RoomLoc identifies a room for per-room runtime (key picked, door open).
type RoomLoc struct {
	Realm Realm
	Room  RoomID
}

// RoomRuntime holds mutable per-room state (pickups, door, cleared enemies).
type RoomRuntime struct {
	KeyTaken       bool
	DoorOpen       bool
	EnemiesCleared bool
	// RupeesGone lists map cells where a rupee was collected (base tile was TileRupee).
	RupeesGone [][2]int
}

func (rt RoomRuntime) RupeeCollected(r, c int) bool {
	for _, p := range rt.RupeesGone {
		if p[0] == r && p[1] == c {
			return true
		}
	}
	return false
}

// WithRupeeCollected returns a copy with this rupee cell marked taken (map stores values).
func (rt RoomRuntime) WithRupeeCollected(r, c int) RoomRuntime {
	if rt.RupeeCollected(r, c) {
		return rt
	}
	rt.RupeesGone = append(append([][2]int(nil), rt.RupeesGone...), [2]int{r, c})
	return rt
}
