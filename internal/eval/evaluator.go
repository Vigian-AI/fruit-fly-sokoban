package eval

import (
	"sokoban-fruit-fly/internal/features"
	"sokoban-fruit-fly/internal/games/sokoban"
	"sokoban-fruit-fly/internal/readout"
	"sokoban-fruit-fly/internal/train"
	"time"
)

type Result struct {
	Success     bool    `json:"success"`
	Reward      float64 `json:"reward"`
	Steps       int     `json:"steps"`
	InferenceMS int64   `json:"inference_ms"`
}

func Run(board sokoban.Board, p *readout.Policy, f *features.Extractor, max int) (Result, error) {
	s := board.Start()
	reward := 0.0
	start := time.Now()
	for step := 0; step < max; step++ {
		x, e := f.Extract(s)
		if e != nil {
			return Result{}, e
		}
		legal := train.Legal(board, s)
		if len(legal) == 0 {
			return Result{Reward: reward - 10, Steps: step, InferenceMS: time.Since(start).Milliseconds()}, nil
		}
		probs, e := p.Probs(x, legal)
		if e != nil {
			return Result{}, e
		}
		a := legal[0]
		for _, candidate := range legal {
			if probs[candidate] > probs[a] {
				a = candidate
			}
		}
		next, moved, e := board.Step(s, a)
		if e != nil {
			return Result{}, e
		}
		if !moved {
			reward -= 2
		} else {
			reward--
		}
		s = next
		if board.Won(s) {
			return Result{true, reward + 100, step + 1, time.Since(start).Milliseconds()}, nil
		}
		if board.Deadlocked(s) {
			return Result{false, reward - 10, step + 1, time.Since(start).Milliseconds()}, nil
		}
	}
	return Result{Reward: reward, Steps: max, InferenceMS: time.Since(start).Milliseconds()}, nil
}
