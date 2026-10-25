# Experiment protocol

Record the seed, level, neuron count, episode count, maximum steps, discount, learning
rate, Go version, and output directory for every run. Run `go test ./...` and
`go vet ./...` before training. Keep `training.csv`, `training.json`, and
`checkpoint.json` together. Evaluate with the same seed and network size used for the
checkpoint, and report success rate, reward, steps, deadlocks, timeouts, training time,
and inference time from the generated files.

No comparison agent or alternate feature baseline is part of this protocol.
