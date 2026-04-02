"""Core simulation: movement, combat, room changes."""

from __future__ import annotations

import random
from dataclasses import dataclass, field
from typing import Dict, List, Tuple

from .constants import (
    ENEMY_MOVE_INTERVAL,
    ENEMY_SHOOT_INTERVAL,
    PROJECTILE_SPEED,
    ROOM_INNER_H,
    ROOM_INNER_W,
    STARTING_HEARTS,
    SWORD_ACTIVE_FRAMES,
    SWORD_COOLDOWN_FRAMES,
)
from .entities import Dir, Enemy, Player, Projectile, RoomEntities, dir_delta
from .world import RoomData, Tile, load_world, tile_blocks_move


@dataclass
class GameState:
    world: Dict[Tuple[int, int], RoomData]
    room_coord: Tuple[int, int]
    player: Player
    room_entities: Dict[Tuple[int, int], RoomEntities] = field(default_factory=dict)
    tick: int = 0
    message: str = ""
    game_over: bool = False

    @classmethod
    def new(cls) -> GameState:
        world = load_world()
        start_room = (1, 1)
        rd = world[start_room]
        # Find a floor tile near center for spawn
        pr, pc = ROOM_INNER_H // 2, ROOM_INNER_W // 2
        if tile_blocks_move(rd.tiles[pr][pc]):
            pr, pc = 5, 18
        p = Player(row=pr, col=pc, hearts=STARTING_HEARTS)
        state = cls(world=world, room_coord=start_room, player=p)
        state._ensure_room_entities(start_room)
        return state

    def _ensure_room_entities(self, coord: Tuple[int, int]) -> None:
        if coord in self.room_entities:
            return
        rd = self.world[coord]
        re = RoomEntities(
            enemies=[Enemy(row=r, col=c) for r, c in rd.enemy_spawns],
            projectiles=[],
        )
        self.room_entities[coord] = re

    def current_room(self) -> RoomData:
        return self.world[self.room_coord]

    def entities_here(self) -> RoomEntities:
        self._ensure_room_entities(self.room_coord)
        return self.room_entities[self.room_coord]

    def in_bounds(self, r: int, c: int) -> bool:
        return 0 <= r < ROOM_INNER_H and 0 <= c < ROOM_INNER_W

    def blocks_at(self, r: int, c: int) -> bool:
        if not self.in_bounds(r, c):
            return True
        return tile_blocks_move(self.current_room().tiles[r][c])

    def enemy_at(self, r: int, c: int) -> Enemy | None:
        for e in self.entities_here().enemies:
            if e.row == r and e.col == c:
                return e
        return None

    def move_player(self, d: Dir) -> None:
        if self.game_over:
            return
        p = self.player
        p.facing = d
        dr, dc = dir_delta(d)
        nr, nc = p.row + dr, p.col + dc
        if not self.in_bounds(nr, nc):
            self._transition_room(d)
            return
        if self.blocks_at(nr, nc):
            return
        if self.enemy_at(nr, nc) is not None:
            return
        p.row, p.col = nr, nc

    def _transition_room(self, exit_dir: Dir) -> None:
        rx, ry = self.room_coord
        p = self.player
        if exit_dir == Dir.UP and (rx, ry - 1) in self.world:
            self.room_coord = (rx, ry - 1)
            p.row = ROOM_INNER_H - 2
            self._ensure_room_entities(self.room_coord)
        elif exit_dir == Dir.DOWN and (rx, ry + 1) in self.world:
            self.room_coord = (rx, ry + 1)
            p.row = 1
            self._ensure_room_entities(self.room_coord)
        elif exit_dir == Dir.LEFT and (rx - 1, ry) in self.world:
            self.room_coord = (rx - 1, ry)
            p.col = ROOM_INNER_W - 2
            self._ensure_room_entities(self.room_coord)
        elif exit_dir == Dir.RIGHT and (rx + 1, ry) in self.world:
            self.room_coord = (rx + 1, ry)
            p.col = 1
            self._ensure_room_entities(self.room_coord)

    def swing_sword(self) -> None:
        if self.game_over:
            return
        p = self.player
        if p.sword_cooldown > 0 or p.sword_timer > 0:
            return
        p.sword_timer = SWORD_ACTIVE_FRAMES
        p.sword_cooldown = SWORD_COOLDOWN_FRAMES

    def _sword_cells(self) -> List[Tuple[int, int]]:
        p = self.player
        if p.sword_timer <= 0:
            return []
        dr, dc = dir_delta(p.facing)
        cells = [(p.row + dr, p.col + dc)]
        # Second tile for longer reach (classic Zelda feel)
        cells.append((p.row + 2 * dr, p.col + 2 * dc))
        return [(r, c) for r, c in cells if self.in_bounds(r, c)]

    def update(self) -> None:
        if self.game_over:
            return
        self.tick += 1
        p = self.player
        if p.invuln_frames > 0:
            p.invuln_frames -= 1
        if p.sword_cooldown > 0:
            p.sword_cooldown -= 1

        re = self.entities_here()
        # Sword hits (apply before tick down so the swing frame connects)
        if p.sword_timer > 0:
            hit = set(self._sword_cells())
            alive: List[Enemy] = []
            for e in re.enemies:
                if (e.row, e.col) in hit:
                    self.message = "Enemy defeated!"
                    continue
                alive.append(e)
            re.enemies = alive
            p.sword_timer -= 1

        self._update_enemies(re)
        self._update_projectiles(re)

        if not re.enemies and self.message == "Enemy defeated!":
            # clear stale message after a moment
            if self.tick % 30 == 0:
                self.message = ""

    def _update_enemies(self, re: RoomEntities) -> None:
        room = self.current_room()
        for e in re.enemies:
            if e.move_cooldown > 0:
                e.move_cooldown -= 1
            else:
                e.move_cooldown = ENEMY_MOVE_INTERVAL + random.randint(0, 4)
                choices = [Dir.UP, Dir.DOWN, Dir.LEFT, Dir.RIGHT]
                random.shuffle(choices)
                for d in choices:
                    dr, dc = dir_delta(d)
                    nr, nc = e.row + dr, e.col + dc
                    if not self.in_bounds(nr, nc):
                        continue
                    if tile_blocks_move(room.tiles[nr][nc]):
                        continue
                    if nr == self.player.row and nc == self.player.col:
                        continue
                    if any(x.row == nr and x.col == nc for x in re.enemies if x is not e):
                        continue
                    e.row, e.col = nr, nc
                    break

            if e.shoot_cooldown > 0:
                e.shoot_cooldown -= 1
            else:
                e.shoot_cooldown = ENEMY_SHOOT_INTERVAL + random.randint(-10, 20)
                if random.random() < 0.35:
                    self._enemy_shoot(e, re)

    def _enemy_shoot(self, e: Enemy, re: RoomEntities) -> None:
        pr, pc = self.player.row, self.player.col
        dr = pr - e.row
        dc = pc - e.col
        dist = abs(dr) + abs(dc)
        if dist < 2 or dist > 14:
            return
        # NES octoroks: axis-aligned shots
        if abs(dr) > abs(dc):
            sdr, sdc = (1 if dr > 0 else -1), 0.0
        else:
            sdr, sdc = 0.0, (1 if dc > 0 else -1)
        re.projectiles.append(
            Projectile(row=float(e.row) + sdr * 0.5, col=float(e.col) + sdc * 0.5, dr=sdr * PROJECTILE_SPEED, dc=sdc * PROJECTILE_SPEED)
        )

    def _update_projectiles(self, re: RoomEntities) -> None:
        room = self.current_room()
        kept: List[Projectile] = []
        p = self.player
        for proj in re.projectiles:
            proj.life -= 1
            if proj.life <= 0:
                continue
            nr = proj.row + proj.dr
            nc = proj.col + proj.dc
            ri, ci = int(round(nr)), int(round(nc))
            if not self.in_bounds(ri, ci):
                continue
            if tile_blocks_move(room.tiles[ri][ci]):
                continue
            # Hit player
            if ri == p.row and ci == p.col and p.invuln_frames <= 0:
                p.hearts -= 1
                p.invuln_frames = 25
                self.message = "Ouch!"
                if p.hearts <= 0:
                    self.game_over = True
                    self.message = "Game over — press R to retry, Q to quit."
                continue
            proj.row, proj.col = nr, nc
            kept.append(proj)
        re.projectiles = kept
