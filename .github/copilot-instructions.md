# Project Guidelines

## Build and Test

- This is a Go 1.21 module (`sokoban-fruit-fly`). Run commands from the repository root.
- Format changed Go files with `gofmt -w <files>`.
- Run the full test suite with `go test ./...`; run static checks with `go vet ./...`.
- Run one package's tests with `go test ./internal/games/sokoban` or
  `go test ./internal/readout`.
- Run one test by name with, for example,
  `go test ./internal/games/sokoban -run '^TestCornerDeadlock$'`.
- The executable is driven through `go run ./cmd/fly <command>`. Available commands are
  `games`, `download`, `features`, `train`, `eval`, and `play`.
- Before an experiment, use `go test ./...` and `go vet ./...`. The full run protocol and
  required recorded parameters are in [docs/experiment_protocol.md](../docs/experiment_protocol.md).

## Architecture

The runtime pipeline is:

`Sokoban board/state -> retina encoder -> frozen spiking feature network -> linear policy`.

- [cmd/fly/main.go](../cmd/fly/main.go) is the CLI composition root. It loads one of the
  built-in levels, selects the network mode, creates the feature extractor, and dispatches
  training, evaluation, feature inspection, or greedy play.
- [internal/games/sokoban](../internal/games/sokoban) parses and validates levels, owns
  board transitions, legal-action checks, win detection, rendering, and corner deadlocks.
  The CLI uses the levels embedded in `internal/games/sokoban/levels.go`; `data/levels`
  contains matching text copies.
- [internal/retina](../internal/retina) converts four board channels (walls, targets,
  boxes, player) into network input. [internal/features](../internal/features) exposes a
  fixed-dimensional extractor to the rest of the system.
- Lightweight mode uses the seeded LIF simulation in [internal/sim/network.go](../internal/sim/network.go).
  Connectome mode uses [internal/sim/connectome.go](../internal/sim/connectome.go) and a
  frozen `FLYGRAPH2` graph loaded by [internal/connectome](../internal/connectome). See
  [docs/connectome.md](../docs/connectome.md) before changing graph or sensor-pool
  handling.
- [internal/readout](../internal/readout) contains the only trainable model: a four-action
  masked softmax policy and a separate linear value baseline. The feature network and
  connectome wiring/weights are not trained.
- [internal/train](../internal/train) runs deterministic on-policy REINFORCE episodes,
  masks illegal actions, applies rewards for movement/win/deadlock, and writes
  `training.csv` and `training.json`. The CLI saves the policy as `checkpoint.json`;
  [internal/eval](../internal/eval) consumes that checkpoint.

## Conventions

- Preserve deterministic experiment control: pass the configured seed to the lightweight
  network and training RNG, and use the same seed, level, neuron count, and network mode
  when evaluating a checkpoint.
- Keep action handling through `game.Actions`, `train.Legal`, and `readout.Policy.Probs`;
  do not bypass legal-action masking or introduce a second action ordering.
- Network mode is explicit: `lightweight` is the default; `connectome` requires an
  existing directory containing `graph.bin` and `meta.json` and must not trigger an
  implicit download. The connectome download/build workflow is documented in
  [README.md](../README.md) and [docs/connectome.md](../docs/connectome.md).
- Keep the feature dimension consistent between the extractor and policy checkpoint.
  Lightweight dimensions come from `-neurons`; connectome dimensions come from the graph's
  readout neurons.
- Treat `configs/default.json` as experiment metadata/reference for now: the current CLI
  uses its flag defaults and does not load that JSON automatically.
- Keep `training.csv`, `training.json`, and `checkpoint.json` together in one output
  directory. Do not commit generated results or downloaded connectome data unless a task
  explicitly requires them.
- When changing the learning objective, state encoding, rewards, or network semantics,
  update [docs/methodology.md](../docs/methodology.md) and relevant tests along with the
  implementation.
