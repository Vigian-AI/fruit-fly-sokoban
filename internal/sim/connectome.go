package sim

import "sokoban-fruit-fly/internal/connectome"

type Connectome struct {
	graph            *connectome.Graph
	decay, threshold float32
	ticks            int
}

type Input struct {
	Neurons []int32
	Current float32
}

func NewConnectome(g *connectome.Graph, ticks int) *Connectome {
	if ticks <= 0 {
		ticks = 48
	}
	return &Connectome{graph: g, decay: .85, threshold: .15, ticks: ticks}
}

func (c *Connectome) Features(input []Input) []float64 {
	n := c.graph.N
	v := make([]float32, n)
	current := make([]float32, n)
	fired := make([]bool, n)
	for _, in := range input {
		for _, i := range in.Neurons {
			if i >= 0 && int(i) < n {
				current[i] += in.Current
			}
		}
	}
	counts := make([]float64, len(c.graph.Readout))
	for t := 0; t < c.ticks; t++ {
		for i := 0; i < n; i++ {
			v[i] = c.decay*v[i] + current[i]
			current[i] = 0
			fired[i] = v[i] >= c.threshold
			if fired[i] {
				v[i] = 0
			}
		}
		for i := 0; i < n; i++ {
			if !fired[i] {
				continue
			}
			for k := c.graph.Start[i]; k < c.graph.Start[i+1]; k++ {
				current[c.graph.Target[k]] += c.graph.Weight[k]
			}
		}
		for k, i := range c.graph.Readout {
			if fired[i] {
				counts[k]++
			}
		}
	}
	for i := range counts {
		counts[i] /= float64(c.ticks)
	}
	return counts
}

func (c *Connectome) Neurons() int           { return len(c.graph.Readout) }
func (c *Connectome) SensorPools() [][]int32 { return c.graph.SensorPools }
