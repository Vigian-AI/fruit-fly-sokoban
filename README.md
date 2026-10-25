# Sokoban Fruit Fly

This project implements a single Sokoban agent whose observations are encoded by a
frozen Fruit-Fly-inspired spiking network and whose four-action linear policy is
trained with on-policy REINFORCE and an internal value baseline.

The lightweight network is an engineering simulation inspired by fruit-fly neural
processing; it is not a biological connectome. It uses seeded random synaptic weights
and leaky integrate-and-fire neurons. The optional connectome mode reads the original
repository's `FLYGRAPH2` graph and `meta.json`; its wiring and weights are frozen. In
both modes only the policy and its baseline learn.

## Run

```text
go test ./...
go vet ./...
go run ./cmd/fly games
go run ./cmd/fly download -raw-dir data/raw
go run ./cmd/fly features -level level01
go run ./cmd/fly train -episodes 100 -output-dir results
go run ./cmd/fly eval -output-dir results
go run ./cmd/fly play -output-dir results
```

The connectome dataset can be downloaded automatically. This downloads the same three
MaleCNS v1.0 Feather tables used by the original repository (about 1.1 GB):

```text
go run ./cmd/fly download -raw-dir data/raw
```

The downloader validates the remote size, writes to a temporary file, and only
publishes a completed file. Existing files with the expected size are skipped. It
does not run implicitly during training.

To use the same MaleCNS connectome graph format as the reference repository, build
`graph.bin` and `meta.json` from those tables with the original repository's `fly build`
(or its equivalent build step), then select the connectome explicitly:

```text
go run ./cmd/fly download -raw-dir ..\fruit-fly\data\raw
go run ..\fruit-fly\cmd\fly build
go run ./cmd/fly features -network-mode connectome -graph-dir ..\fruit-fly\data\processed
go run ./cmd/fly train -network-mode connectome -graph-dir ..\fruit-fly\data\processed
```

Training writes `training.csv`, `training.json`, and `checkpoint.json`. The command
line currently uses the built-in levels in `internal/games/sokoban/levels.go`; the
matching text copies are in `data/levels`.

## Architecture

`Sokoban state -> retina encoder -> frozen LIF network -> activity vector -> masked
linear policy -> action`. The environment validates box/target counts, supports
`# . $ @ * +`, detects corner deadlocks, and enforces a configurable episode limit.

For a transition, REINFORCE uses `A_t = G_t - V(s_t)`. The policy gradient is
`-A_t * grad(log pi(a_t|s_t))`; the value head is trained toward the discounted
return. Illegal actions are excluded before softmax.

## Limitations

The current version provides a lightweight network by default and small deterministic
levels. Connectome mode requires the downloaded tables and a built `FLYGRAPH2` graph.
Training metrics are execution outputs and are not pre-filled with experimental claims.
