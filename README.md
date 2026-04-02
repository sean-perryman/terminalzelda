# Terminal Zelda

NES **Zelda** vibes in your terminal: linked overworld screens, sword arc, enemies that shoot back. Meant for **SSH**—one binary, no browser, no GUI.

```bash
go run .
# or
go build -o zelda . && ./zelda
```

**Controls:** arrows · WASD · HJKL · **Z** / space (sword) · **B** buy a heart on **`S`** (shop) · **Q** save and quit · **R** new game (deletes save) · **Ctrl+C** quits without saving

**Leaving a screen:** walk through the **middle of any wall** — there is a **two-tile gap** (shown as **`·`** on the border). Same idea as the NES game: you step off the edge into the next area.

You need a real TTY (`ssh -t user@host`). UTF-8 locale helps for hearts and bullets. Window about **40×14** or larger.

**Dungeon:** In the southeast overworld screen, step on **`>`** to enter a three-room cave. Find **`K`**, use the key on **`+`**, defeat the critters, then stand on **`%`** (triforce). **`<`** takes you back to the entrance.

**Rupees:** **`$`** on the ground (and **1 rupee** per enemy you defeat). Stand on **`S`** in the central-south overworld and press **B** to buy a heart for **10** rupees (up to **8** full hearts).

**Save:** Progress is stored under your config directory, e.g. `~/.config/terminalzelda/save.json` (see [XDG](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html)). The game loads it automatically on startup.

**Stack:** Go 1.22+, [tcell](https://github.com/gdamore/tcell).

---

*Fan work—not affiliated with Nintendo.*
