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

// RoomRuntime holds mutable dungeon room state (initial layout is in RoomData).
type RoomRuntime struct {
	KeyTaken bool
	DoorOpen bool
}
