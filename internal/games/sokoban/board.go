package sokoban

import (
	"fmt"
	"sort"
	"strings"

	"sokoban-fruit-fly/internal/game"
)

type Board struct {
	Width, Height  int
	Walls, Targets map[game.Position]bool
	InitialPlayer  game.Position
	InitialBoxes   []game.Position
}

func position(row, col int) game.Position { return game.Position{Row: row, Col: col} }

func Parse(text string) (Board, error) {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return Board{}, fmt.Errorf("level is empty")
	}
	b := Board{Height: len(lines), Walls: map[game.Position]bool{}, Targets: map[game.Position]bool{}}
	for r, line := range lines {
		if len(line) > b.Width {
			b.Width = len(line)
		}
		for c, ch := range line {
			p := position(r, c)
			switch ch {
			case '#':
				b.Walls[p] = true
			case '.':
				b.Targets[p] = true
			case '$':
				b.InitialBoxes = append(b.InitialBoxes, p)
			case '@':
				b.InitialPlayer = p
			case '*':
				b.Targets[p] = true
				b.InitialBoxes = append(b.InitialBoxes, p)
			case '+':
				b.Targets[p] = true
				b.InitialPlayer = p
			case ' ', '\t':
			default:
				return Board{}, fmt.Errorf("invalid level character %q at row %d col %d", ch, r, c)
			}
		}
	}
	if len(b.Targets) != len(b.InitialBoxes) {
		return Board{}, fmt.Errorf("level has %d targets but %d boxes", len(b.Targets), len(b.InitialBoxes))
	}
	if b.InitialPlayer == (game.Position{}) && !containsPosition(lines, '@') && !containsPosition(lines, '+') {
		return Board{}, fmt.Errorf("level has no player")
	}
	return b, nil
}

func containsPosition(lines []string, ch rune) bool {
	for _, l := range lines {
		if strings.ContainsRune(l, ch) {
			return true
		}
	}
	return false
}

func (b Board) Start() State {
	boxes := append([]game.Position(nil), b.InitialBoxes...)
	return State{Player: b.InitialPlayer, Boxes: boxes}
}

type State = game.State

func (b Board) Key(s State) string {
	boxes := append([]game.Position(nil), s.Boxes...)
	sort.Slice(boxes, func(i, j int) bool {
		if boxes[i].Row == boxes[j].Row {
			return boxes[i].Col < boxes[j].Col
		}
		return boxes[i].Row < boxes[j].Row
	})
	var out strings.Builder
	fmt.Fprintf(&out, "%d,%d|", s.Player.Row, s.Player.Col)
	for _, p := range boxes {
		fmt.Fprintf(&out, "%d,%d;", p.Row, p.Col)
	}
	return out.String()
}

func (b Board) Render(s State) string {
	box := map[game.Position]bool{}
	for _, p := range s.Boxes {
		box[p] = true
	}
	var out strings.Builder
	for r := 0; r < b.Height; r++ {
		for c := 0; c < b.Width; c++ {
			p := position(r, c)
			ch := ' '
			switch {
			case b.Walls[p]:
				ch = '#'
			case b.Targets[p]:
				ch = '.'
			}
			if box[p] {
				if b.Targets[p] {
					ch = '*'
				} else {
					ch = '$'
				}
			}
			if s.Player == p {
				if b.Targets[p] {
					ch = '+'
				} else {
					ch = '@'
				}
			}
			out.WriteRune(ch)
		}
		if r+1 < b.Height {
			out.WriteByte('\n')
		}
	}
	return out.String()
}
