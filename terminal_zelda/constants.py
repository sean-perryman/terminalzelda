"""Shared sizes and tuning (canonical room fits 80x24 terminals)."""

# Inner playable area (walls sit outside this in the padded grid).
ROOM_INNER_W = 36
ROOM_INNER_H = 10

# Sword attack duration and cooldown (frames at ~20 FPS logic ticks).
SWORD_ACTIVE_FRAMES = 4
SWORD_COOLDOWN_FRAMES = 12

# Enemy tuning.
ENEMY_MOVE_INTERVAL = 8
ENEMY_SHOOT_INTERVAL = 45
PROJECTILE_SPEED = 1

# Player starts with N half-hearts (display as ♥ pairs).
STARTING_HEARTS = 6
