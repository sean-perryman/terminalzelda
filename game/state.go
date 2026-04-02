package game

import (
	"fmt"
	"math"
	"math/rand/v2"
)

type State struct {
	Overworld map[RoomID]RoomData
	Dungeon   map[RoomID]RoomData
	Realm     Realm
	Room      RoomID

	Player       Player
	RoomEntities map[RoomLoc]*RoomEntities
	Runtimes     map[RoomLoc]RoomRuntime

	OWReturnRoom RoomID
	OWReturnRow  int
	OWReturnCol  int

	Tick     int
	Message  string
	GameOver bool
	Victory  bool
}

func (s *State) roomLoc() RoomLoc {
	return RoomLoc{Realm: s.Realm, Room: s.Room}
}

func (s *State) activeWorld() map[RoomID]RoomData {
	if s.Realm == RealmDungeon {
		return s.Dungeon
	}
	return s.Overworld
}

func NewState() *State {
	s := &State{
		Overworld:    LoadOverworld(),
		Dungeon:      LoadDungeon(),
		Realm:        RealmOverworld,
		Room:         RoomID{1, 1},
		RoomEntities: make(map[RoomLoc]*RoomEntities),
		Runtimes:     make(map[RoomLoc]RoomRuntime),
	}
	rd := s.Overworld[s.Room]
	pr, pc := RoomInnerH/2, RoomInnerW/2
	if TileBlocksMove(rd.Tiles[pr][pc]) {
		pr, pc = 5, 18
	}
	s.Player = Player{Row: pr, Col: pc, Facing: DirDown, Hearts: StartingHearts}
	s.ensureRoomEntities()
	return s
}

func (s *State) ensureRoomEntities() {
	loc := s.roomLoc()
	if _, ok := s.RoomEntities[loc]; ok {
		return
	}
	rd := s.activeWorld()[s.Room]
	re := &RoomEntities{
		Enemies:     make([]Enemy, 0, len(rd.EnemySpawns)),
		Projectiles: nil,
	}
	rt := s.Runtimes[loc]
	if !rt.EnemiesCleared {
		for _, sp := range rd.EnemySpawns {
			re.Enemies = append(re.Enemies, Enemy{Row: sp[0], Col: sp[1], ShootCooldown: 20})
		}
	}
	s.RoomEntities[loc] = re
}

func (s *State) CurrentRoomData() RoomData {
	return s.activeWorld()[s.Room]
}

func (s *State) baseTileAt(r, c int) Tile {
	return s.CurrentRoomData().Tiles[r][c]
}

// EffectiveTile is the logical tile after pickups and opened doors (dungeon).
func (s *State) EffectiveTile(r, c int) Tile {
	t := s.baseTileAt(r, c)
	loc := s.roomLoc()
	rt := s.Runtimes[loc]
	if t == TileRupee && rt.RupeeCollected(r, c) {
		return TileFloor
	}
	if s.Realm != RealmDungeon {
		return t
	}
	if t == TileDungeonKey && rt.KeyTaken {
		return TileFloor
	}
	if t == TileLockedDoor && rt.DoorOpen {
		return TileFloor
	}
	return t
}

func (s *State) EntitiesHere() *RoomEntities {
	s.ensureRoomEntities()
	return s.RoomEntities[s.roomLoc()]
}

func (s *State) InBounds(r, c int) bool {
	return r >= 0 && r < RoomInnerH && c >= 0 && c < RoomInnerW
}

func (s *State) blocksAt(r, c int) bool {
	if !s.InBounds(r, c) {
		return true
	}
	return TileBlocksMove(s.EffectiveTile(r, c))
}

func (s *State) enemyAt(r, c int) *Enemy {
	re := s.EntitiesHere()
	for i := range re.Enemies {
		e := &re.Enemies[i]
		if e.Row == r && e.Col == c {
			return e
		}
	}
	return nil
}

func (s *State) handleLockedDoorBump(nr, nc int) {
	if s.baseTileAt(nr, nc) != TileLockedDoor {
		return
	}
	loc := s.roomLoc()
	rt := s.Runtimes[loc]
	if rt.DoorOpen {
		return
	}
	if s.Player.HasDungeonKey {
		rt.DoorOpen = true
		s.Runtimes[loc] = rt
		s.Player.HasDungeonKey = false
		s.Message = "The door opened!"
	} else {
		s.Message = "It's locked."
	}
}

func (s *State) afterEnteringTile() {
	if s.GameOver || s.Victory {
		return
	}
	p := &s.Player
	r, c := p.Row, p.Col
	switch s.baseTileAt(r, c) {
	case TileStairsDown:
		s.enterDungeon()
	case TileStairsUp:
		s.exitDungeonFromStairs()
	case TileDungeonKey:
		if s.Realm != RealmDungeon {
			return
		}
		loc := s.roomLoc()
		rt := s.Runtimes[loc]
		if rt.KeyTaken {
			return
		}
		rt.KeyTaken = true
		s.Runtimes[loc] = rt
		s.Player.HasDungeonKey = true
		s.Message = "You got a small key!"
	case TileGoal:
		if s.Realm != RealmDungeon {
			return
		}
		s.Victory = true
		s.Message = "You claimed the triforce! Press R to play again."
	case TileRupee:
		loc := s.roomLoc()
		rt := s.Runtimes[loc]
		if rt.RupeeCollected(r, c) {
			return
		}
		s.Runtimes[loc] = rt.WithRupeeCollected(r, c)
		p.Rupees++
		s.Message = "Found a rupee!"
	}
}

// TryBuyHeart spends rupees when standing on a shop tile (S). Returns true if a purchase happened.
func (s *State) TryBuyHeart() bool {
	if s.GameOver || s.Victory {
		return false
	}
	p := &s.Player
	if s.baseTileAt(p.Row, p.Col) != TileShop {
		s.Message = "Stand on the shop (S) and press B to buy."
		return false
	}
	if p.Rupees < ShopHeartPrice {
		s.Message = fmt.Sprintf("You need %d rupees for a heart.", ShopHeartPrice)
		return false
	}
	if p.Hearts >= MaxHalfHearts {
		s.Message = "Your hearts are already full."
		return false
	}
	p.Rupees -= ShopHeartPrice
	p.Hearts += 2
	if p.Hearts > MaxHalfHearts {
		p.Hearts = MaxHalfHearts
	}
	s.Message = fmt.Sprintf("Bought a heart! (-%d rupees)", ShopHeartPrice)
	return true
}

func (s *State) enterDungeon() {
	if s.Realm == RealmDungeon {
		return
	}
	s.OWReturnRoom = s.Room
	s.OWReturnRow = s.Player.Row
	s.OWReturnCol = s.Player.Col
	s.Realm = RealmDungeon
	s.Room = RoomID{0, 0}
	s.Player.Row = 5
	s.Player.Col = 12
	s.Player.Facing = DirDown
	s.ensureRoomEntities()
}

func (s *State) exitDungeonFromStairs() {
	if s.Realm != RealmDungeon {
		return
	}
	s.Realm = RealmOverworld
	s.Room = s.OWReturnRoom
	s.Player.Row = s.OWReturnRow
	s.Player.Col = s.OWReturnCol
	s.Player.Facing = DirDown
	s.ensureRoomEntities()
}

func (s *State) MovePlayer(d Dir) {
	if s.GameOver || s.Victory {
		return
	}
	p := &s.Player
	p.Facing = d
	dr, dc := DirDelta(d)
	nr, nc := p.Row+dr, p.Col+dc
	if !s.InBounds(nr, nc) {
		s.transitionRoom(d)
		return
	}
	if s.baseTileAt(nr, nc) == TileLockedDoor {
		rt := s.Runtimes[s.roomLoc()]
		if !rt.DoorOpen {
			s.handleLockedDoorBump(nr, nc)
			return
		}
	}
	if s.blocksAt(nr, nc) {
		return
	}
	if s.enemyAt(nr, nc) != nil {
		return
	}
	p.Row, p.Col = nr, nc
	s.afterEnteringTile()
}

func (s *State) transitionRoom(exitDir Dir) {
	rx, ry := s.Room.X, s.Room.Y
	p := &s.Player
	w := s.activeWorld()
	var next RoomID
	switch exitDir {
	case DirUp:
		next = RoomID{rx, ry - 1}
	case DirDown:
		next = RoomID{rx, ry + 1}
	case DirLeft:
		next = RoomID{rx - 1, ry}
	case DirRight:
		next = RoomID{rx + 1, ry}
	default:
		return
	}
	if _, ok := w[next]; !ok {
		return
	}
	s.Room = next
	switch exitDir {
	case DirUp:
		p.Row = RoomInnerH - 2
	case DirDown:
		p.Row = 1
	case DirLeft:
		p.Col = RoomInnerW - 2
	case DirRight:
		p.Col = 1
	}
	s.ensureRoomEntities()
}

func (s *State) SwingSword() {
	if s.GameOver || s.Victory {
		return
	}
	p := &s.Player
	if p.SwordCooldown > 0 || p.SwordTimer > 0 {
		return
	}
	p.SwordTimer = SwordActiveFrames
	p.SwordCooldown = SwordCooldownFrames
}

func (s *State) swordCells() [][2]int {
	p := &s.Player
	if p.SwordTimer <= 0 {
		return nil
	}
	dr, dc := DirDelta(p.Facing)
	var out [][2]int
	for dist := 1; dist <= 2; dist++ {
		r, c := p.Row+dr*dist, p.Col+dc*dist
		if s.InBounds(r, c) {
			out = append(out, [2]int{r, c})
		}
	}
	return out
}

func (s *State) Update() {
	if s.GameOver || s.Victory {
		return
	}
	s.Tick++
	p := &s.Player
	if p.InvulnFrames > 0 {
		p.InvulnFrames--
	}
	if p.SwordCooldown > 0 {
		p.SwordCooldown--
	}

	re := s.EntitiesHere()
	rd := s.CurrentRoomData()
	hadSpawns := len(rd.EnemySpawns) > 0
	if p.SwordTimer > 0 {
		hit := make(map[[2]int]struct{})
		for _, cell := range s.swordCells() {
			hit[[2]int{cell[0], cell[1]}] = struct{}{}
		}
		killed := 0
		alive := re.Enemies[:0]
		for i := range re.Enemies {
			e := re.Enemies[i]
			if _, ok := hit[[2]int{e.Row, e.Col}]; ok {
				killed++
				continue
			}
			alive = append(alive, e)
		}
		re.Enemies = alive
		p.Rupees += killed
		if killed > 0 {
			s.Message = "Enemy defeated!"
		}
		p.SwordTimer--
		if hadSpawns && len(re.Enemies) == 0 {
			loc := s.roomLoc()
			rt := s.Runtimes[loc]
			rt.EnemiesCleared = true
			s.Runtimes[loc] = rt
		}
	}

	s.updateEnemies(re)
	s.updateProjectiles(re)

	if len(re.Enemies) == 0 && s.Message == "Enemy defeated!" && s.Tick%30 == 0 {
		s.Message = ""
	}
}

func (s *State) updateEnemies(re *RoomEntities) {
	p := &s.Player
	for i := range re.Enemies {
		e := &re.Enemies[i]
		if e.MoveCooldown > 0 {
			e.MoveCooldown--
		} else {
			e.MoveCooldown = EnemyMoveInterval + rand.IntN(5)
			dirs := []Dir{DirUp, DirDown, DirLeft, DirRight}
			rand.Shuffle(len(dirs), func(a, b int) { dirs[a], dirs[b] = dirs[b], dirs[a] })
			for _, d := range dirs {
				dr, dc := DirDelta(d)
				nr, nc := e.Row+dr, e.Col+dc
				if !s.InBounds(nr, nc) {
					continue
				}
				if s.blocksAt(nr, nc) {
					continue
				}
				if nr == p.Row && nc == p.Col {
					continue
				}
				blocked := false
				for j := range re.Enemies {
					if j == i {
						continue
					}
					o := &re.Enemies[j]
					if o.Row == nr && o.Col == nc {
						blocked = true
						break
					}
				}
				if blocked {
					continue
				}
				e.Row, e.Col = nr, nc
				break
			}
		}

		if e.ShootCooldown > 0 {
			e.ShootCooldown--
		} else {
			e.ShootCooldown = EnemyShootInterval + rand.IntN(31) - 10
			if rand.Float64() < 0.35 {
				s.enemyShoot(e, re)
			}
		}
	}
}

func (s *State) enemyShoot(e *Enemy, re *RoomEntities) {
	pr, pc := s.Player.Row, s.Player.Col
	dr := pr - e.Row
	dc := pc - e.Col
	dist := abs(dr) + abs(dc)
	if dist < 2 || dist > 14 {
		return
	}
	var sdr, sdc float64
	if abs(dr) > abs(dc) {
		if dr > 0 {
			sdr = 1
		} else {
			sdr = -1
		}
		sdc = 0
	} else {
		sdr = 0
		if dc > 0 {
			sdc = 1
		} else {
			sdc = -1
		}
	}
	re.Projectiles = append(re.Projectiles, Projectile{
		Row:  float64(e.Row) + sdr*0.5,
		Col:  float64(e.Col) + sdc*0.5,
		Dr:   sdr * ProjectileSpeed,
		Dc:   sdc * ProjectileSpeed,
		Life: 120,
	})
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (s *State) updateProjectiles(re *RoomEntities) {
	p := &s.Player
	kept := re.Projectiles[:0]
	for i := range re.Projectiles {
		proj := re.Projectiles[i]
		proj.Life--
		if proj.Life <= 0 {
			continue
		}
		nr := proj.Row + proj.Dr
		nc := proj.Col + proj.Dc
		ri := int(math.Round(nr))
		ci := int(math.Round(nc))
		if !s.InBounds(ri, ci) {
			continue
		}
		if s.blocksAt(ri, ci) {
			continue
		}
		if ri == p.Row && ci == p.Col && p.InvulnFrames <= 0 {
			p.Hearts--
			p.InvulnFrames = 25
			s.Message = "Ouch!"
			if p.Hearts <= 0 {
				s.GameOver = true
				s.Message = "Game over — press R to retry, Q to quit."
			}
			continue
		}
		proj.Row, proj.Col = nr, nc
		kept = append(kept, proj)
	}
	re.Projectiles = kept
}

func (s *State) Reset() {
	n := NewState()
	s.Overworld = n.Overworld
	s.Dungeon = n.Dungeon
	s.Realm = n.Realm
	s.Room = n.Room
	s.Player = n.Player
	s.RoomEntities = n.RoomEntities
	s.Runtimes = n.Runtimes
	s.OWReturnRoom = n.OWReturnRoom
	s.OWReturnRow = n.OWReturnRow
	s.OWReturnCol = n.OWReturnCol
	s.Tick = 0
	s.Message = ""
	s.GameOver = false
	s.Victory = false
}
