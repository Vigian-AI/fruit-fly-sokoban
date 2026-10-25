package readout

import (
	"sokoban-fruit-fly/internal/game"
	"testing"
)

func TestProbsMaskAndNormalize(t *testing.T) {
	p := New(2)
	got, err := p.Probs([]float64{1, 2}, []game.Action{game.Left, game.Right})
	if err != nil {
		t.Fatal(err)
	}
	if got[game.Up] != 0 || got[game.Down] != 0 {
		t.Fatal("illegal action was not masked")
	}
	if got[game.Left]+got[game.Right] < .999 || got[game.Left]+got[game.Right] > 1.001 {
		t.Fatal("probabilities do not sum to one")
	}
}
