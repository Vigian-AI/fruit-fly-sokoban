package train

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"sokoban-fruit-fly/internal/game"
	"sokoban-fruit-fly/internal/games/sokoban"
	"sokoban-fruit-fly/internal/readout"
)

type Config struct {
	Episodes     int     `json:"episodes"`
	MaxSteps     int     `json:"max_steps"`
	LearningRate float64 `json:"learning_rate"`
	Gamma        float64 `json:"gamma"`
	Seed         int64   `json:"seed"`
	Neurons      int     `json:"neurons"`
}
type Metric struct {
	Episode    int     `json:"episode"`
	Reward     float64 `json:"reward"`
	Steps      int     `json:"steps"`
	Success    bool    `json:"success"`
	Deadlock   bool    `json:"deadlock"`
	Timeout    bool    `json:"timeout"`
	DurationMS int64   `json:"duration_ms"`
}

type Trainer struct {
	Board     sokoban.Board
	Extractor func(sokoban.State) ([]float64, error)
	Policy    *readout.Policy
	Config    Config
}

func Legal(b sokoban.Board, s sokoban.State) []game.Action {
	out := []game.Action{}
	for _, a := range game.Actions() {
		if b.Legal(s, a) {
			out = append(out, a)
		}
	}
	return out
}

func (t *Trainer) Run(output string) ([]Metric, error) {
	if t.Config.Episodes <= 0 {
		return nil, fmt.Errorf("episodes must be positive")
	}
	if t.Config.MaxSteps <= 0 {
		return nil, fmt.Errorf("max_steps must be positive")
	}
	rng := randSource(t.Config.Seed)
	metrics := make([]Metric, 0, t.Config.Episodes)
	for ep := 1; ep <= t.Config.Episodes; ep++ {
		start := time.Now()
		state := t.Board.Start()
		transitions := []readout.Transition{}
		reward := 0.0
		steps := 0
		success, deadlock, timeout := false, false, false
		for step := 0; step < t.Config.MaxSteps; step++ {
			x, e := t.Extractor(state)
			if e != nil {
				return nil, e
			}
			legal := Legal(t.Board, state)
			if len(legal) == 0 {
				deadlock = true
				reward -= 10
				break
			}
			probs, e := t.Policy.Probs(x, legal)
			if e != nil {
				return nil, e
			}
			a := readout.Sample(probs, legal, rng)
			transitions = append(transitions, readout.Transition{Features: x, Action: a, Probs: probs, Value: t.Policy.Value(x)})
			next, moved, e := t.Board.Step(state, a)
			if e != nil {
				return nil, e
			}
			steps++
			if !moved {
				reward -= 2
			} else {
				reward -= 1
			}
			if t.Board.Won(next) {
				reward += 100
				success = true
				state = next
				break
			}
			state = next
			if t.Board.Deadlocked(state) {
				reward -= 10
				deadlock = true
				break
			}
		}
		if !success && !deadlock {
			timeout = true
		}
		ret := reward
		for i := len(transitions) - 1; i >= 0; i-- {
			transitions[i].Return = ret
			ret *= t.Config.Gamma
		}
		t.Policy.Update(transitions, t.Config.LearningRate, t.Config.Gamma)
		metrics = append(metrics, Metric{ep, reward, steps, success, deadlock, timeout, time.Since(start).Milliseconds()})
	}
	if err := os.MkdirAll(output, 0755); err != nil {
		return nil, err
	}
	f, e := os.Create(filepath.Join(output, "training.csv"))
	if e != nil {
		return nil, e
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"episode", "reward", "steps", "success", "deadlock", "timeout", "duration_ms"})
	for _, m := range metrics {
		_ = w.Write([]string{fmt.Sprint(m.Episode), fmt.Sprint(m.Reward), fmt.Sprint(m.Steps), fmt.Sprint(m.Success), fmt.Sprint(m.Deadlock), fmt.Sprint(m.Timeout), fmt.Sprint(m.DurationMS)})
	}
	w.Flush()
	if e = w.Error(); e != nil {
		return nil, e
	}
	b, e := json.MarshalIndent(metrics, "", "  ")
	if e != nil {
		return nil, e
	}
	if e = os.WriteFile(filepath.Join(output, "training.json"), b, 0644); e != nil {
		return nil, e
	}
	return metrics, nil
}

func randSource(seed int64) *rand.Rand { return rand.New(rand.NewSource(seed)) }
