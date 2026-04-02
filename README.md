# Terminal Zelda

NES **Zelda** vibes in your terminal: linked overworld screens, sword arc, enemies that shoot back. Meant for **SSH**—one binary, no browser, no GUI.

```bash
go run .
# or
go build -o zelda . && ./zelda
```

**Controls:** arrows · WASD · HJKL · **Z** / space (sword) · **R** restart · **Q** quit · **Ctrl+C** bail

You need a real TTY (`ssh -t user@host`). UTF-8 locale helps for hearts and bullets. Window about **40×14** or larger.

**Stack:** Go 1.22+, [tcell](https://github.com/gdamore/tcell).

---

*Fan work—not affiliated with Nintendo.*
