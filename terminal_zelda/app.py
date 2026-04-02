"""curses front-end: runs cleanly over SSH with UTF-8 locale."""

from __future__ import annotations

import curses
import sys

from .constants import ROOM_INNER_H, ROOM_INNER_W, STARTING_HEARTS
from .entities import Dir
from .game import GameState
from .world import Tile


def _tile_char(t: Tile) -> str:
    if t == Tile.WALL:
        return "#"
    if t == Tile.WATER:
        return "~"
    if t == Tile.BUSH:
        return "*"
    return "."


def _sword_glyph(facing: Dir) -> str:
    if facing in (Dir.LEFT, Dir.RIGHT):
        return "-"
    return "|"


def draw(stdscr: "curses._CursesWindow", state: GameState) -> None:
    stdscr.erase()
    rows, cols = stdscr.getmaxyx()
    need_h = ROOM_INNER_H + 5
    need_w = ROOM_INNER_W + 4
    if rows < need_h or cols < need_w:
        msg = f"Terminal too small (need at least {need_w}x{need_h}). Resize and press R."
        try:
            stdscr.addstr(0, 0, msg[: max(0, cols - 1)])
        except curses.error:
            pass
        stdscr.refresh()
        return

    rx, ry = state.room_coord
    title = f" Terminal Zelda — overworld room ({rx}, {ry}) "
    try:
        stdscr.attron(curses.color_pair(4))
        stdscr.addstr(0, max(0, (cols - len(title)) // 2), title[:cols])
        stdscr.attroff(curses.color_pair(4))
    except curses.error:
        pass

    p = state.player
    full = p.hearts // 2
    half = p.hearts % 2
    max_full = (STARTING_HEARTS + 1) // 2
    filled = full + (1 if half else 0)
    heart_bar = "\u2665" * full + ("\u2661" if half else "")
    heart_bar += "\u00b7" * max(0, max_full - filled)
    hud = f" {heart_bar}  "
    if state.message:
        hud += state.message
    try:
        stdscr.addstr(1, 0, hud[: cols - 1])
    except curses.error:
        pass

    off_r = 3
    off_c = max(0, (cols - ROOM_INNER_W) // 2)
    room = state.current_room()
    re = state.entities_here()

    proj_cells = {(int(round(pr.row)), int(round(pr.col))) for pr in re.projectiles}
    enemy_cells = {(e.row, e.col) for e in re.enemies}
    sword_cells: set[tuple[int, int]] = set()
    if p.sword_timer > 0:
        dr, dc = [( -1, 0), (0, 1), (1, 0), (0, -1)][int(p.facing)]
        for dist in (1, 2):
            r, c = p.row + dr * dist, p.col + dc * dist
            if 0 <= r < ROOM_INNER_H and 0 <= c < ROOM_INNER_W:
                sword_cells.add((r, c))

    for r in range(ROOM_INNER_H):
        for c in range(ROOM_INNER_W):
            ch = _tile_char(room.tiles[r][c])
            pair = 1
            if room.tiles[r][c] == Tile.WATER:
                pair = 2
            elif room.tiles[r][c] == Tile.BUSH:
                pair = 5
            elif room.tiles[r][c] == Tile.WALL:
                pair = 3

            pr, pc = r + off_r, c + off_c
            if r == p.row and c == p.col:
                if p.invuln_frames > 0 and (state.tick // 3) % 2 == 0:
                    ch = _tile_char(room.tiles[r][c])
                else:
                    ch = "@"
                    pair = 6
            elif (r, c) in enemy_cells:
                ch = "o"
                pair = 7
            elif (r, c) in proj_cells:
                ch = "\u2022"
                pair = 7
            elif (r, c) in sword_cells:
                ch = _sword_glyph(p.facing)
                pair = 6

            try:
                stdscr.attron(curses.color_pair(pair))
                stdscr.addstr(pr, pc, ch)
                stdscr.attroff(curses.color_pair(pair))
            except curses.error:
                pass

    help_line = " Arrows/WASD move  Z/Space sword  R restart  Q quit "
    try:
        stdscr.attron(curses.color_pair(4))
        y = off_r + ROOM_INNER_H
        stdscr.addstr(y, max(0, (cols - len(help_line)) // 2), help_line[:cols])
        stdscr.attroff(curses.color_pair(4))
        if state.game_over:
            go = " GAME OVER "
            stdscr.attron(curses.A_BOLD | curses.color_pair(7))
            stdscr.addstr(y + 1, max(0, (cols - len(go)) // 2), go[:cols])
            stdscr.attroff(curses.A_BOLD | curses.color_pair(7))
    except curses.error:
        pass

    stdscr.refresh()


def _handle_key(state: GameState, ch: int) -> bool:
    """Return True to exit."""
    if ch in (ord("q"), ord("Q")):
        return True
    if ch in (ord("r"), ord("R")):
        # reassign by replacing core fields
        new = GameState.new()
        state.world = new.world
        state.room_coord = new.room_coord
        state.player = new.player
        state.room_entities = new.room_entities
        state.tick = 0
        state.message = ""
        state.game_over = False
        return False

    if state.game_over:
        return False

    if ch in (curses.KEY_UP, ord("k"), ord("K"), ord("w"), ord("W")):
        state.move_player(Dir.UP)
    elif ch in (curses.KEY_DOWN, ord("j"), ord("J"), ord("s"), ord("S")):
        state.move_player(Dir.DOWN)
    elif ch in (curses.KEY_LEFT, ord("h"), ord("H"), ord("a"), ord("A")):
        state.move_player(Dir.LEFT)
    elif ch in (curses.KEY_RIGHT, ord("l"), ord("L"), ord("d"), ord("D")):
        state.move_player(Dir.RIGHT)
    elif ch in (ord(" "), ord("z"), ord("Z")):
        state.swing_sword()
    return False


def _init_colors() -> None:
    curses.start_color()
    curses.use_default_colors()
    # pair -> (fg, bg) -1 is default
    curses.init_pair(1, curses.COLOR_GREEN, -1)  # grass
    curses.init_pair(2, curses.COLOR_CYAN, -1)  # water
    curses.init_pair(3, curses.COLOR_YELLOW, -1)  # walls
    curses.init_pair(4, curses.COLOR_WHITE, -1)  # ui
    curses.init_pair(5, curses.COLOR_GREEN, -1)  # bush (bold later)
    curses.init_pair(6, curses.COLOR_YELLOW, -1)  # player / sword
    curses.init_pair(7, curses.COLOR_RED, -1)  # enemies


def run(stdscr: "curses._CursesWindow") -> None:
    curses.curs_set(0)
    stdscr.nodelay(True)
    stdscr.timeout(50)
    _init_colors()

    state = GameState.new()
    while True:
        ch = stdscr.getch()
        if _handle_key(state, ch):
            break
        state.update()
        draw(stdscr, state)


def main() -> None:
    if not sys.stdout.isatty():
        print("This game needs an interactive terminal (try SSH with TTY allocation).", file=sys.stderr)
        sys.exit(1)
    try:
        curses.wrapper(run)
    except KeyboardInterrupt:
        pass
