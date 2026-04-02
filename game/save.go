package game

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const saveVersion = 1

type saveFile struct {
	V      int  `json:"v"`
	Realm  int  `json:"realm"`
	RX     int  `json:"rx"`
	RY     int  `json:"ry"`
	PR     int  `json:"pr"`
	PC     int  `json:"pc"`
	Facing int  `json:"facing"`
	Hearts int  `json:"hearts"`
	Rupees int  `json:"rupees"`
	HasKey bool `json:"has_key"`

	OWX  int `json:"ow_rx"`
	OWY  int `json:"ow_ry"`
	OWPR int `json:"ow_pr"`
	OWPC int `json:"ow_pc"`

	Rooms []saveRoom `json:"rooms"`
}

type saveRoom struct {
	Realm          int      `json:"realm"`
	RX             int      `json:"rx"`
	RY             int      `json:"ry"`
	KeyTaken       bool     `json:"key_taken"`
	DoorOpen       bool     `json:"door_open"`
	EnemiesCleared bool     `json:"enemies_cleared"`
	RupeesGone     [][2]int `json:"rupees_gone"`
}

// SavePath returns the path to save.json under the user config directory.
func SavePath() (string, error) {
	d, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "terminalzelda", "save.json"), nil
}

// DeleteSave removes the save file if it exists.
func DeleteSave() error {
	p, err := SavePath()
	if err != nil {
		return err
	}
	return os.Remove(p)
}

// Save writes the current game to disk (call before quitting).
func (s *State) Save() error {
	p, err := SavePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	sf := saveFile{
		V:      saveVersion,
		Realm:  int(s.Realm),
		RX:     s.Room.X,
		RY:     s.Room.Y,
		PR:     s.Player.Row,
		PC:     s.Player.Col,
		Facing: int(s.Player.Facing),
		Hearts: s.Player.Hearts,
		Rupees: s.Player.Rupees,
		HasKey: s.Player.HasDungeonKey,
		OWX:    s.OWReturnRoom.X,
		OWY:    s.OWReturnRoom.Y,
		OWPR:   s.OWReturnRow,
		OWPC:   s.OWReturnCol,
	}
	for loc, rt := range s.Runtimes {
		sf.Rooms = append(sf.Rooms, saveRoom{
			Realm:          int(loc.Realm),
			RX:             loc.Room.X,
			RY:             loc.Room.Y,
			KeyTaken:       rt.KeyTaken,
			DoorOpen:       rt.DoorOpen,
			EnemiesCleared: rt.EnemiesCleared,
			RupeesGone:     append([][2]int(nil), rt.RupeesGone...),
		})
	}
	data, err := json.MarshalIndent(sf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o600)
}

// LoadGame restores from save, or returns an error if missing/invalid.
func LoadGame() (*State, error) {
	p, err := SavePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		return nil, err
	}
	var sf saveFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return nil, err
	}
	if sf.V != saveVersion || sf.PR < 0 || sf.PC < 0 {
		return nil, errors.New("unsupported or corrupt save")
	}

	s := &State{
		Overworld:    LoadOverworld(),
		Dungeon:      LoadDungeon(),
		Realm:        Realm(sf.Realm),
		Room:         RoomID{sf.RX, sf.RY},
		RoomEntities: make(map[RoomLoc]*RoomEntities),
		Runtimes:     make(map[RoomLoc]RoomRuntime),
		OWReturnRoom: RoomID{sf.OWX, sf.OWY},
		OWReturnRow:  sf.OWPR,
		OWReturnCol:  sf.OWPC,
		Player: Player{
			Row:           sf.PR,
			Col:           sf.PC,
			Facing:        Dir(sf.Facing),
			Hearts:        sf.Hearts,
			Rupees:        sf.Rupees,
			HasDungeonKey: sf.HasKey,
		},
	}
	if s.Player.Hearts < 1 {
		s.Player.Hearts = StartingHearts
	}
	if s.Player.Hearts > MaxHalfHearts {
		s.Player.Hearts = MaxHalfHearts
	}
	if s.Player.Rupees < 0 {
		s.Player.Rupees = 0
	}
	if s.Player.Facing < DirUp || s.Player.Facing > DirLeft {
		s.Player.Facing = DirDown
	}
	if _, ok := s.activeWorld()[s.Room]; !ok {
		return nil, errors.New("save room out of bounds")
	}
	for _, r := range sf.Rooms {
		loc := RoomLoc{Realm: Realm(r.Realm), Room: RoomID{r.RX, r.RY}}
		s.Runtimes[loc] = RoomRuntime{
			KeyTaken:       r.KeyTaken,
			DoorOpen:       r.DoorOpen,
			EnemiesCleared: r.EnemiesCleared,
			RupeesGone:     append([][2]int(nil), r.RupeesGone...),
		}
	}
	s.ensureRoomEntities()
	return s, nil
}
