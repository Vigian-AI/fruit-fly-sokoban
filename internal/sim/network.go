package sim

import (
	"math"
	"math/rand"
)

type Config struct {
	InputSize int   `json:"input_size"`
	Neurons   int   `json:"neurons"`
	Ticks     int   `json:"ticks"`
	Seed      int64 `json:"seed"`
}

type Network struct {
	cfg     Config
	weights [][]float64
	neurons []Neuron
}

func New(c Config) *Network {
	if c.Neurons <= 0 {
		c.Neurons = 64
	}
	if c.Ticks <= 0 {
		c.Ticks = 8
	}
	r := rand.New(rand.NewSource(c.Seed))
	n := &Network{cfg: c, weights: make([][]float64, c.Neurons), neurons: make([]Neuron, c.Neurons)}
	for i := range n.weights {
		n.weights[i] = make([]float64, c.InputSize)
		for j := range n.weights[i] {
			n.weights[i][j] = r.NormFloat64() / math.Sqrt(float64(c.InputSize))
		}
		n.neurons[i] = Neuron{Threshold: 0.45, Decay: 0.85}
	}
	return n
}

func (n *Network) Features(input []float64) []float64 {
	if len(input) != n.cfg.InputSize {
		panic("sim: input size mismatch")
	}
	out := make([]float64, len(n.neurons))
	for t := 0; t < n.cfg.Ticks; t++ {
		for i, w := range n.weights {
			current := 0.0
			for j, x := range input {
				current += w[j] * x
			}
			if n.neurons[i].Step(current) {
				out[i]++
			}
		}
	}
	for i := range out {
		out[i] /= float64(n.cfg.Ticks)
	}
	return out
}

func (n *Network) Neurons() int { return len(n.neurons) }
