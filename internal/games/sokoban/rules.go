package sokoban

import (
	"fmt"
	"sokoban-fruit-fly/internal/game"
)

func delta(a game.Action) game.Position {
	switch a {
	case game.Up:
		return position(-1, 0)
	case game.Down:
		return position(1, 0)
	case game.Left:
		return position(0, -1)
	default:
		return position(0, 1)
	}
}

func (b Board) Legal(s State, a game.Action) bool {
	d := delta(a)
	next := position(s.Player.Row+d.Row, s.Player.Col+d.Col)
	if b.Walls[next] {
		return false
	}
	i := boxAt(s, next)
	if i < 0 {
		return true
	}
	beyond := position(next.Row+d.Row, next.Col+d.Col)
	return !b.Walls[beyond] && boxAt(s, beyond) < 0
}

func boxAt(s State, p game.Position) int {
	for i, b := range s.Boxes {
		if b == p {
			return i
		}
	}
	return -1
}

func (b Board) Step(s State, a game.Action) (State, bool, error) {
	if a < game.Up || a > game.Right {
		return s, false, fmt.Errorf("invalid action %d", a)
	}
	if !b.Legal(s, a) {
		return s, false, nil
	}
	d := delta(a)
	next := position(s.Player.Row+d.Row, s.Player.Col+d.Col)
	out := State{Player: next, Boxes: append([]game.Position(nil), s.Boxes...)}
	if i := boxAt(out, next); i >= 0 {
		out.Boxes[i] = position(next.Row+d.Row, next.Col+d.Col)
	}
	return out, true, nil
}

func (b Board) Won(s State) bool {
	for _, p := range s.Boxes {
		if !b.Targets[p] {
			return false
		}
	}
	return len(s.Boxes) > 0
}

func (b Board) Deadlocked(s State) bool {
	for _, p := range s.Boxes {
		if b.Targets[p] {
			continue
		}
		if (b.Walls[position(p.Row-1, p.Col)] || b.Walls[position(p.Row+1, p.Col)]) &&
			(b.Walls[position(p.Row, p.Col-1)] || b.Walls[position(p.Row, p.Col+1)]) {
			return true
		}
	}
	return false
}
