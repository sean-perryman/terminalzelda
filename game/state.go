package game

import (
	"math"
	"math/rand/v2"
)

type State struct {
	World        map[RoomID]RoomData
	Room         RoomID
	Player       Player
	RoomEntities map[RoomID]*RoomEntities
	Tick         int
	Message      string
	GameOver     bool
}

func NewState() *State {
	world := LoadWorld()
	start := RoomID{1, 1}
	rd := world[start]
	pr, pc := RoomInnerH/2, RoomInnerW/2
	if TileBlocksMove(rd.Tiles[pr][pc]) {
		pr, pc = 5, 18
	}
	s := &State{
		World:        world,
		Room:         start,
		Player:       Player{Row: pr, Col: pc, Facing: DirDown, Hearts: StartingHearts},
		RoomEntities: make(map[RoomID]*RoomEntities),
	}
	s.ensureRoomEntities(start)
	return s
}

func (s *State) ensureRoomEntities(id RoomID) {
	if _, ok := s.RoomEntities[id]; ok {
		return
	}
	rd := s.World[id]
	re := &RoomEntities{
		Enemies:     make([]Enemy, 0, len(rd.EnemySpawns)),
		Projectiles: nil,
	}
	for _, sp := range rd.EnemySpawns {
		re.Enemies = append(re.Enemies, Enemy{Row: sp[0], Col: sp[1], ShootCooldown: 20})
	}
	s.RoomEntities[id] = re
}

func (s *State) CurrentRoomData() RoomData {
	return s.World[s.Room]
}

func (s *State) EntitiesHere() *RoomEntities {
	s.ensureRoomEntities(s.Room)
	return s.RoomEntities[s.Room]
}

func (s *State) InBounds(r, c int) bool {
	return r >= 0 && r < RoomInnerH && c >= 0 && c < RoomInnerW
}

func (s *State) blocksAt(r, c int) bool {
	if !s.InBounds(r, c) {
		return true
	}
	rd := s.CurrentRoomData()
	return TileBlocksMove(rd.Tiles[r][c])
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

func (s *State) MovePlayer(d Dir) {
	if s.GameOver {
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
	if s.blocksAt(nr, nc) {
		return
	}
	if s.enemyAt(nr, nc) != nil {
		return
	}
	p.Row, p.Col = nr, nc
}

func (s *State) transitionRoom(exitDir Dir) {
	rx, ry := s.Room.X, s.Room.Y
	p := &s.Player
	next := s.Room
	switch exitDir {
	case DirUp:
		next = RoomID{rx, ry - 1}
		if _, ok := s.World[next]; ok {
			s.Room = next
			p.Row = RoomInnerH - 2
			s.ensureRoomEntities(next)
		}
	case DirDown:
		next = RoomID{rx, ry + 1}
		if _, ok := s.World[next]; ok {
			s.Room = next
			p.Row = 1
			s.ensureRoomEntities(next)
		}
	case DirLeft:
		next = RoomID{rx - 1, ry}
		if _, ok := s.World[next]; ok {
			s.Room = next
			p.Col = RoomInnerW - 2
			s.ensureRoomEntities(next)
		}
	case DirRight:
		next = RoomID{rx + 1, ry}
		if _, ok := s.World[next]; ok {
			s.Room = next
			p.Col = 1
			s.ensureRoomEntities(next)
		}
	}
}

func (s *State) SwingSword() {
	if s.GameOver {
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
	if s.GameOver {
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
	if p.SwordTimer > 0 {
		hit := make(map[[2]int]struct{})
		for _, cell := range s.swordCells() {
			hit[[2]int{cell[0], cell[1]}] = struct{}{}
		}
		alive := re.Enemies[:0]
		for i := range re.Enemies {
			e := re.Enemies[i]
			if _, ok := hit[[2]int{e.Row, e.Col}]; ok {
				s.Message = "Enemy defeated!"
				continue
			}
			alive = append(alive, e)
		}
		re.Enemies = alive
		p.SwordTimer--
	}

	s.updateEnemies(re)
	s.updateProjectiles(re)

	if len(re.Enemies) == 0 && s.Message == "Enemy defeated!" && s.Tick%30 == 0 {
		s.Message = ""
	}
}

func (s *State) updateEnemies(re *RoomEntities) {
	rd := s.CurrentRoomData()
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
				if TileBlocksMove(rd.Tiles[nr][nc]) {
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
	rd := s.CurrentRoomData()
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
		if TileBlocksMove(rd.Tiles[ri][ci]) {
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
	s.World = n.World
	s.Room = n.Room
	s.Player = n.Player
	s.RoomEntities = n.RoomEntities
	s.Tick = 0
	s.Message = ""
	s.GameOver = false
}
