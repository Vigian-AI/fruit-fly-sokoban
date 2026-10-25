package retina

import (
	"fmt"
	"sokoban-fruit-fly/internal/connectome"
	"sokoban-fruit-fly/internal/game"
	"sokoban-fruit-fly/internal/games/sokoban"
	"sokoban-fruit-fly/internal/sim"
)

type featureNetwork interface {
	Features([]float64) []float64
	Neurons() int
}

type Encoder struct {
	board      sokoban.Board
	network    featureNetwork
	connectome *sim.Connectome
	groups     [][]int32
}

func New(board sokoban.Board, seed int64, neurons int) *Encoder {
	return &Encoder{board: board, network: sim.New(sim.Config{InputSize: board.Width * board.Height * 4, Neurons: neurons, Ticks: 8, Seed: seed})}
}

func NewConnectome(board sokoban.Board, g *connectome.Graph, ticks int) (*Encoder, error) {
	if len(g.SensorPools) < 2 {
		return nil, fmt.Errorf("connectome has %d sensor pools; need 2 photoreceptor pools", len(g.SensorPools))
	}
	cells := board.Width * board.Height
	groups := make([][]int32, 2)
	for ch := range groups {
		per := len(g.SensorPools[ch]) / cells
		if per == 0 {
			return nil, fmt.Errorf("sensor pool %d has fewer neurons than board cells", ch)
		}
		groups[ch] = g.SensorPools[ch][:per*cells]
	}
	return &Encoder{board: board, connectome: sim.NewConnectome(g, ticks), groups: groups}, nil
}

func (e *Encoder) Neurons() int {
	if e.connectome != nil {
		return e.connectome.Neurons()
	}
	return e.network.Neurons()
}

func (e *Encoder) Encode(s sokoban.State) []float64 {
	cells := e.board.Width * e.board.Height
	input := make([]float64, cells*4)
	set := func(channel int, p game.Position) {
		if p.Row >= 0 && p.Row < e.board.Height && p.Col >= 0 && p.Col < e.board.Width {
			input[channel*cells+p.Row*e.board.Width+p.Col] = 1
		}
	}
	for p := range e.board.Walls {
		set(0, p)
	}
	for p := range e.board.Targets {
		set(1, p)
	}
	for _, p := range s.Boxes {
		set(2, p)
	}
	set(3, s.Player)
	if e.groups == nil {
		return e.network.Features(input)
	}
	stimulus := make([]sim.Input, 0)
	for ch := range e.groups {
		per := len(e.groups[ch]) / cells
		for cell := 0; cell < cells; cell++ {
			lit := input[cell] + input[cells+cell]
			if ch == 1 {
				lit = input[2*cells+cell] + input[3*cells+cell]
			}
			if lit > 0 {
				stimulus = append(stimulus, sim.Input{Neurons: e.groups[ch][cell*per : (cell+1)*per], Current: .5})
			}
		}
	}
	return e.connectome.Features(stimulus)
}
