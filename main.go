package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"golang.org/x/term"

	"terminalzelda/game"
)

func main() {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		fmt.Fprintln(os.Stderr, "This game needs an interactive terminal (try SSH with TTY allocation).")
		os.Exit(1)
	}

	s, err := tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "terminal: %v\n", err)
		os.Exit(1)
	}
	if err := s.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "init: %v\n", err)
		os.Exit(1)
	}
	defer s.Fini()

	s.SetStyle(tcell.StyleDefault)
	s.Clear()

	st := game.NewState()

	evCh := make(chan tcell.Event, 64)
	go func() {
		for {
			ev := s.PollEvent()
			evCh <- ev
			if ev == nil {
				return
			}
		}
	}()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	draw(s, st)
	s.Show()

	for {
		select {
		case ev := <-evCh:
			if ev == nil {
				return
			}
			switch ev := ev.(type) {
			case *tcell.EventResize:
				s.Sync()
				draw(s, st)
				s.Show()
			case *tcell.EventKey:
				if handleKey(st, ev) {
					return
				}
				draw(s, st)
				s.Show()
			}
		case <-ticker.C:
			st.Update()
			draw(s, st)
			s.Show()
		}
	}
}

func handleKey(st *game.State, ev *tcell.EventKey) bool {
	if ev.Key() == tcell.KeyCtrlC {
		return true
	}
	if ev.Key() == tcell.KeyRune {
		switch r := ev.Rune(); r {
		case 'q', 'Q':
			return true
		case 'r', 'R':
			st.Reset()
			return false
		}
	}

	if st.GameOver {
		return false
	}

	switch ev.Key() {
	case tcell.KeyUp:
		st.MovePlayer(game.DirUp)
	case tcell.KeyDown:
		st.MovePlayer(game.DirDown)
	case tcell.KeyLeft:
		st.MovePlayer(game.DirLeft)
	case tcell.KeyRight:
		st.MovePlayer(game.DirRight)
	case tcell.KeyRune:
		switch r := ev.Rune(); r {
		case 'w', 'W', 'k', 'K':
			st.MovePlayer(game.DirUp)
		case 's', 'S', 'j', 'J':
			st.MovePlayer(game.DirDown)
		case 'a', 'A', 'h', 'H':
			st.MovePlayer(game.DirLeft)
		case 'd', 'D', 'l', 'L':
			st.MovePlayer(game.DirRight)
		case 'z', 'Z', ' ':
			st.SwingSword()
		}
	}
	return false
}

func tileRune(t game.Tile) rune {
	switch t {
	case game.TileWall:
		return '#'
	case game.TileWater:
		return '~'
	case game.TileBush:
		return '*'
	default:
		return '.'
	}
}

func swordRune(d game.Dir) rune {
	if d == game.DirLeft || d == game.DirRight {
		return '-'
	}
	return '|'
}

func drawString(sc tcell.Screen, x, y int, st tcell.Style, s string, maxW int) {
	col := x
	for _, r := range s {
		if col >= maxW {
			return
		}
		if r == '\n' {
			return
		}
		sc.SetContent(col, y, r, nil, st)
		col++
	}
}

func draw(sc tcell.Screen, st *game.State) {
	w, h := sc.Size()
	needW, needH := game.RoomInnerW+4, game.RoomInnerH+5
	if w < needW || h < needH {
		sc.Clear()
		msg := fmt.Sprintf("Terminal too small (need at least %dx%d). Resize and press R.", needW, needH)
		drawString(sc, 0, 0, tcell.StyleDefault.Foreground(tcell.ColorRed), msg, w)
		return
	}

	sc.Clear()

	title := fmt.Sprintf(" Terminal Zelda — overworld room (%d, %d) ", st.Room.X, st.Room.Y)
	titleStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	drawString(sc, max(0, (w-len(title))/2), 0, titleStyle, title, w)

	p := &st.Player
	full := p.Hearts / 2
	half := p.Hearts % 2
	maxFull := (game.StartingHearts + 1) / 2
	filled := full
	if half != 0 {
		filled++
	}
	hud := " "
	for i := 0; i < full; i++ {
		hud += "\u2665"
	}
	if half != 0 {
		hud += "\u2661"
	}
	for i := 0; i < max(0, maxFull-filled); i++ {
		hud += "\u00b7"
	}
	hud += "  "
	if st.Message != "" {
		hud += st.Message
	}
	drawString(sc, 0, 1, tcell.StyleDefault.Foreground(tcell.ColorWhite), hud, w)

	offR := 3
	offC := max(0, (w-game.RoomInnerW)/2)
	rd := st.CurrentRoomData()
	re := st.EntitiesHere()

	projCells := make(map[[2]int]struct{})
	for _, pr := range re.Projectiles {
		ri := int(math.Round(pr.Row))
		ci := int(math.Round(pr.Col))
		projCells[[2]int{ri, ci}] = struct{}{}
	}

	enemyCells := make(map[[2]int]struct{})
	for _, e := range re.Enemies {
		enemyCells[[2]int{e.Row, e.Col}] = struct{}{}
	}

	swordCells := make(map[[2]int]struct{})
	if p.SwordTimer > 0 {
		dr, dc := game.DirDelta(p.Facing)
		for dist := 1; dist <= 2; dist++ {
			r, c := p.Row+dr*dist, p.Col+dc*dist
			if r >= 0 && r < game.RoomInnerH && c >= 0 && c < game.RoomInnerW {
				swordCells[[2]int{r, c}] = struct{}{}
			}
		}
	}

	green := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	cyan := tcell.StyleDefault.Foreground(tcell.ColorAqua)
	yellow := tcell.StyleDefault.Foreground(tcell.ColorYellow)
	white := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	red := tcell.StyleDefault.Foreground(tcell.ColorRed)

	for r := 0; r < game.RoomInnerH; r++ {
		for c := 0; c < game.RoomInnerW; c++ {
			t := rd.Tiles[r][c]
			ch := tileRune(t)
			stl := green
			switch t {
			case game.TileWater:
				stl = cyan
			case game.TileBush:
				stl = green
			case game.TileWall:
				stl = yellow
			}

			x, y := offC+c, offR+r
			if r == p.Row && c == p.Col {
				if p.InvulnFrames > 0 && (st.Tick/3)%2 == 0 {
					ch = tileRune(t)
					stl = green
					switch t {
					case game.TileWater:
						stl = cyan
					case game.TileWall:
						stl = yellow
					case game.TileBush:
						stl = green
					}
				} else {
					ch = '@'
					stl = yellow
				}
			} else if _, ok := enemyCells[[2]int{r, c}]; ok {
				ch = 'o'
				stl = red
			} else if _, ok := projCells[[2]int{r, c}]; ok {
				ch = '\u2022'
				stl = red
			} else if _, ok := swordCells[[2]int{r, c}]; ok {
				ch = swordRune(p.Facing)
				stl = yellow
			}

			sc.SetContent(x, y, ch, nil, stl)
		}
	}

	help := " Arrows/WASD move  Z/Space sword  R restart  Q quit "
	drawString(sc, max(0, (w-len(help))/2), offR+game.RoomInnerH, white, help, w)
	if st.GameOver {
		goMsg := " GAME OVER "
		stGo := tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true)
		drawString(sc, max(0, (w-len(goMsg))/2), offR+game.RoomInnerH+1, stGo, goMsg, w)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
