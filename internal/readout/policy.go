package readout

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sokoban-fruit-fly/internal/game"
)

type Policy struct {
	Dimension    int         `json:"dimension"`
	Weights      [][]float64 `json:"weights"`
	Bias         []float64   `json:"bias"`
	ValueWeights []float64   `json:"value_weights"`
	ValueBias    float64     `json:"value_bias"`
}

func New(d int) *Policy {
	p := &Policy{Dimension: d, Weights: make([][]float64, 4), Bias: make([]float64, 4), ValueWeights: make([]float64, d)}
	for i := range p.Weights {
		p.Weights[i] = make([]float64, d)
	}
	return p
}

func (p *Policy) logits(x []float64) []float64 {
	z := make([]float64, 4)
	for a := range z {
		z[a] = p.Bias[a]
		for i, v := range x {
			z[a] += p.Weights[a][i] * v
		}
	}
	return z
}

func (p *Policy) Probs(x []float64, legal []game.Action) ([]float64, error) {
	if len(legal) == 0 {
		return nil, fmt.Errorf("cannot compute policy with no legal actions")
	}
	z := p.logits(x)
	maxz := math.Inf(-1)
	allowed := map[game.Action]bool{}
	for _, a := range legal {
		if a < game.Up || a > game.Right {
			return nil, fmt.Errorf("invalid legal action %d", a)
		}
		allowed[a] = true
		if z[a] > maxz {
			maxz = z[a]
		}
	}
	out := make([]float64, 4)
	sum := 0.0
	for _, a := range legal {
		out[a] = math.Exp(z[a] - maxz)
		sum += out[a]
	}
	for a := range out {
		if allowed[game.Action(a)] {
			out[a] /= sum
		}
	}
	return out, nil
}

func Sample(probs []float64, legal []game.Action, r *rand.Rand) game.Action {
	x := r.Float64()
	total := 0.0
	for _, a := range legal {
		total += probs[a]
		if x <= total {
			return a
		}
	}
	return legal[len(legal)-1]
}
func (p *Policy) Value(x []float64) float64 {
	v := p.ValueBias
	for i, w := range x {
		v += w * p.ValueWeights[i]
	}
	return v
}

type Transition struct {
	Features []float64
	Action   game.Action
	Probs    []float64
	Return   float64
	Value    float64
}

func (p *Policy) Update(batch []Transition, lr, gamma float64) {
	if len(batch) == 0 {
		return
	}
	gw := make([][]float64, 4)
	for a := range gw {
		gw[a] = make([]float64, p.Dimension)
	}
	gb := make([]float64, 4)
	gv := make([]float64, p.Dimension)
	gc := 0.0
	for _, t := range batch {
		adv := t.Return - t.Value
		for a := 0; a < 4; a++ {
			if t.Probs[a] == 0 {
				continue
			}
			dz := adv * (t.Probs[a] - boolFloat(a == int(t.Action)))
			for i, x := range t.Features {
				gw[a][i] += dz * x
			}
			gb[a] += dz
		}
		dv := t.Value - t.Return
		for i, x := range t.Features {
			gv[i] += dv * x
		}
		gc += dv
	}
	scale := lr / float64(len(batch))
	for a := 0; a < 4; a++ {
		for i := range p.Weights[a] {
			p.Weights[a][i] -= scale * gw[a][i]
		}
		p.Bias[a] -= scale * gb[a]
	}
	for i := range p.ValueWeights {
		p.ValueWeights[i] -= scale * 0.5 * gv[i]
	}
	p.ValueBias -= scale * 0.5 * gc
	_ = gamma
}
func boolFloat(v bool) float64 {
	if v {
		return 1
	}
	return 0
}

func (p *Policy) Save(path string) error {
	b, e := json.MarshalIndent(p, "", "  ")
	if e != nil {
		return e
	}
	return os.WriteFile(path, b, 0644)
}
func Load(path string) (*Policy, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return nil, e
	}
	var p Policy
	if e = json.Unmarshal(b, &p); e != nil {
		return nil, e
	}
	if p.Dimension <= 0 || len(p.Weights) != 4 {
		return nil, fmt.Errorf("invalid policy checkpoint")
	}
	return &p, nil
}
