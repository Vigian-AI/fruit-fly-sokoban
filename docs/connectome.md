# Connectome-backed mode

The `connectome` network mode consumes the exact frozen graph format produced by the
reference `fruit-fly` repository: `graph.bin` with magic `FLYGRAPH2` plus `meta.json`.
The graph contains the signed sparse edges, in-degrees, photoreceptor pools, and the
512 descending plus 512 VNC motor readout neurons selected by the reference builder.

The Sokoban encoder maps walls and targets to the reference R1-R6 pool and boxes and
the player to the R8 pool. The wiring and weights remain frozen; only the four-action
readout policy is trained.

Build the graph with the original repository's `fly build`, then copy or point this
project at that directory:

```text
go run ./cmd/fly features -network-mode connectome -graph-dir ..\fruit-fly\data\processed
go run ./cmd/fly train -network-mode connectome -graph-dir ..\fruit-fly\data\processed
```

No connectome dataset is downloaded automatically. If `graph.bin` or `meta.json` is
missing, the command fails with an explicit error.
