package features

import (
	"fmt"
	"sokoban-fruit-fly/internal/connectome"
	"sokoban-fruit-fly/internal/games/sokoban"
	"sokoban-fruit-fly/internal/retina"
)

type Extractor struct {
	encoder   *retina.Encoder
	dimension int
}

func NewConnectome(board sokoban.Board, g *connectome.Graph, ticks int) (*Extractor, error) {
	e, err := retina.NewConnectome(board, g, ticks)
	if err != nil {
		return nil, err
	}
	return &Extractor{encoder: e, dimension: e.Neurons()}, nil
}

func New(board sokoban.Board, seed int64, neurons int) *Extractor {
	return &Extractor{encoder: retina.New(board, seed, neurons), dimension: neurons}
}

func (e *Extractor) Extract(s sokoban.State) ([]float64, error) {
	x := e.encoder.Encode(s)
	if len(x) != e.dimension {
		return nil, fmt.Errorf("feature dimension %d, want %d", len(x), e.dimension)
	}
	return x, nil
}

func (e *Extractor) Dimension() int { return e.dimension }
