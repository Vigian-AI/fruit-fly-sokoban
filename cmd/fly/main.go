package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"sokoban-fruit-fly/internal/connectome"
	"sokoban-fruit-fly/internal/download"
	"sokoban-fruit-fly/internal/eval"
	"sokoban-fruit-fly/internal/features"
	"sokoban-fruit-fly/internal/games/sokoban"
	"sokoban-fruit-fly/internal/readout"
	"sokoban-fruit-fly/internal/train"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}
	switch os.Args[1] {
	case "games":
		fmt.Println("sokoban: level01, level02")
	case "download":
		downloadCommand(os.Args[2:])
	case "train":
		trainCommand(os.Args[2:])
	case "eval":
		evalCommand(os.Args[2:])
	case "features":
		featuresCommand(os.Args[2:])
	case "play":
		playCommand(os.Args[2:])
	default:
		usage()
	}
}
func usage() { fmt.Println("fly {games|download|features|train|eval|play} [options]") }
func downloadCommand(args []string) {
	fs := flag.NewFlagSet("download", flag.ExitOnError)
	dir := fs.String("raw-dir", "data/raw", "directory for the MaleCNS feather tables")
	fs.Parse(args)
	if err := download.All(*dir); err != nil {
		fatal(fmt.Errorf("download connectome dataset: %w", err))
	}
	fmt.Printf("connectome tables are ready in %s\n", *dir)
}
func common(fs *flag.FlagSet) (*string, *int, *int64, *int, *string, *string, *string, *string) {
	level := fs.String("level", "level01", "level name")
	neurons := fs.Int("neurons", 64, "number of simulated neurons")
	seed := fs.Int64("seed", 1, "random seed")
	max := fs.Int("max-steps", 100, "episode step limit")
	out := fs.String("output-dir", "results", "output directory")
	config := fs.String("config", "configs/default.json", "configuration file (reserved for experiment metadata)")
	mode := fs.String("network-mode", "lightweight", "network mode: lightweight or connectome")
	graphDir := fs.String("graph-dir", "data/connectome", "directory containing graph.bin and meta.json")
	return level, neurons, seed, max, out, config, mode, graphDir
}
func extractor(b sokoban.Board, seed int64, neurons int, mode, graphDir string) (*features.Extractor, error) {
	if mode == "lightweight" {
		return features.New(b, seed, neurons), nil
	}
	if mode != "connectome" {
		return nil, fmt.Errorf("unknown network mode %q", mode)
	}
	g, _, err := connectome.Load(graphDir)
	if err != nil {
		return nil, err
	}
	return features.NewConnectome(b, g, 48)
}
func board(name string) (sokoban.Board, error) { return sokoban.Load(name) }
func trainCommand(args []string) {
	fs := flag.NewFlagSet("train", flag.ExitOnError)
	level, n, seed, max, out, _, mode, graphDir := common(fs)
	episodes := fs.Int("episodes", 100, "training episodes")
	lr := fs.Float64("learning-rate", 0.01, "policy learning rate")
	gamma := fs.Float64("gamma", 0.99, "discount factor")
	fs.Parse(args)
	b, e := board(*level)
	if e != nil {
		fatal(e)
	}
	f, e := extractor(b, *seed, *n, *mode, *graphDir)
	if e != nil {
		fatal(e)
	}
	p := readout.New(f.Dimension())
	cfg := train.Config{Episodes: *episodes, MaxSteps: *max, LearningRate: *lr, Gamma: *gamma, Seed: *seed, Neurons: *n}
	metrics, e := (&train.Trainer{Board: b, Extractor: f.Extract, Policy: p, Config: cfg}).Run(*out)
	if e != nil {
		fatal(e)
	}
	if e = p.Save(filepath.Join(*out, "checkpoint.json")); e != nil {
		fatal(e)
	}
	fmt.Printf("trained %d episodes; checkpoint=%s\n", len(metrics), filepath.Join(*out, "checkpoint.json"))
}
func evalCommand(args []string) {
	fs := flag.NewFlagSet("eval", flag.ExitOnError)
	level, n, seed, max, out, _, mode, graphDir := common(fs)
	fs.Parse(args)
	b, e := board(*level)
	if e != nil {
		fatal(e)
	}
	f, e := extractor(b, *seed, *n, *mode, *graphDir)
	if e != nil {
		fatal(e)
	}
	p, e := readout.Load(filepath.Join(*out, "checkpoint.json"))
	if e != nil {
		fatal(e)
	}
	r, e := eval.Run(b, p, f, *max)
	if e != nil {
		fatal(e)
	}
	v, _ := json.MarshalIndent(r, "", "  ")
	fmt.Println(string(v))
}
func featuresCommand(args []string) {
	fs := flag.NewFlagSet("features", flag.ExitOnError)
	level, n, seed, _, _, _, mode, graphDir := common(fs)
	fs.Parse(args)
	b, e := board(*level)
	if e != nil {
		fatal(e)
	}
	f, e := extractor(b, *seed, *n, *mode, *graphDir)
	if e != nil {
		fatal(e)
	}
	x, e := f.Extract(b.Start())
	if e != nil {
		fatal(e)
	}
	fmt.Printf("level=%s feature-dimension=%d activity=%v\n", *level, len(x), x)
}
func playCommand(args []string) {
	fs := flag.NewFlagSet("play", flag.ExitOnError)
	level, n, seed, max, out, _, mode, graphDir := common(fs)
	fs.Parse(args)
	b, e := board(*level)
	if e != nil {
		fatal(e)
	}
	f, e := extractor(b, *seed, *n, *mode, *graphDir)
	if e != nil {
		fatal(e)
	}
	p, e := readout.Load(filepath.Join(*out, "checkpoint.json"))
	if e != nil {
		fatal(e)
	}
	s := b.Start()
	for i := 0; i < *max && !b.Won(s); i++ {
		fmt.Printf("\n%s\n", b.Render(s))
		x, _ := f.Extract(s)
		legal := train.Legal(b, s)
		if len(legal) == 0 {
			break
		}
		probs, _ := p.Probs(x, legal)
		a := legal[0]
		for _, candidate := range legal {
			if probs[candidate] > probs[a] {
				a = candidate
			}
		}
		s, _, e = b.Step(s, a)
		if e != nil {
			fatal(e)
		}
		time.Sleep(50 * time.Millisecond)
	}
	fmt.Printf("\n%s\nwon=%v\n", b.Render(s), b.Won(s))
}
func fatal(e error) { fmt.Fprintln(os.Stderr, e); os.Exit(1) }
