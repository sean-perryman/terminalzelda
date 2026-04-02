package game

type Dir int

const (
	DirUp Dir = iota
	DirRight
	DirDown
	DirLeft
)

func DirDelta(d Dir) (dr, dc int) {
	switch d {
	case DirUp:
		return -1, 0
	case DirDown:
		return 1, 0
	case DirLeft:
		return 0, -1
	default:
		return 0, 1
	}
}

type Player struct {
	Row, Col      int
	Facing        Dir
	Hearts        int
	InvulnFrames  int
	SwordTimer    int
	SwordCooldown int
}

type Enemy struct {
	Row, Col      int
	MoveCooldown  int
	ShootCooldown int
}

type Projectile struct {
	Row, Col float64
	Dr, Dc   float64
	Life     int
}

type RoomEntities struct {
	Enemies     []Enemy
	Projectiles []Projectile
}
