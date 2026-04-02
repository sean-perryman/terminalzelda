package game

const (
	RoomInnerW = 36
	RoomInnerH = 10

	SwordActiveFrames   = 4
	SwordCooldownFrames = 12

	EnemyMoveInterval  = 8
	EnemyShootInterval = 45
	ProjectileSpeed    = 1.0

	StartingHearts = 6
	// MaxHalfHearts caps heart containers (each ♥ is two half-hearts).
	MaxHalfHearts = 16
	// ShopHeartPrice is rupees for +1 full heart (two half-hearts).
	ShopHeartPrice = 10
)
