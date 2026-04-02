"""Player, enemies, and projectiles."""

from __future__ import annotations

from dataclasses import dataclass, field
from enum import IntEnum
from typing import List, Tuple


class Dir(IntEnum):
    UP = 0
    RIGHT = 1
    DOWN = 2
    LEFT = 3


@dataclass
class Player:
    row: int
    col: int
    facing: Dir = Dir.DOWN
    hearts: int = 6  # half-hearts
    invuln_frames: int = 0
    sword_timer: int = 0
    sword_cooldown: int = 0


def dir_delta(d: Dir) -> Tuple[int, int]:
    if d == Dir.UP:
        return (-1, 0)
    if d == Dir.DOWN:
        return (1, 0)
    if d == Dir.LEFT:
        return (0, -1)
    return (0, 1)


@dataclass
class Enemy:
    row: int
    col: int
    move_cooldown: int = 0
    shoot_cooldown: int = 20


@dataclass
class Projectile:
    row: float
    col: float
    dr: float
    dc: float
    life: int = 120


@dataclass
class RoomEntities:
    enemies: List[Enemy] = field(default_factory=list)
    projectiles: List[Projectile] = field(default_factory=list)
