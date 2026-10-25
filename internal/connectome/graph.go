package connectome

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const graphMagic = "FLYGRAPH2"

type Graph struct {
	N           int
	Start       []int64
	Target      []int32
	Weight      []float32
	InDegree    []float32
	Region      []uint8
	SensorPools [][]int32
	Readout     []int32
}

type Meta struct {
	Neurons        int      `json:"neurons"`
	Connections    int      `json:"connections"`
	ReadoutClasses []string `json:"readout_classes"`
	ReadoutTypes   []string `json:"readout_types"`
	ReadoutBodyIDs []int64  `json:"readout_body_ids"`
}

func Load(dir string) (*Graph, *Meta, error) {
	f, err := os.Open(filepath.Join(dir, "graph.bin"))
	if err != nil {
		return nil, nil, fmt.Errorf("open connectome graph: %w", err)
	}
	defer f.Close()
	r := bufio.NewReaderSize(f, 1<<20)
	magic := make([]byte, len(graphMagic))
	if _, err := io.ReadFull(r, magic); err != nil || string(magic) != graphMagic {
		return nil, nil, fmt.Errorf("graph.bin: bad header")
	}
	var readErr error
	get := func(v any) {
		if readErr == nil {
			readErr = binary.Read(r, binary.LittleEndian, v)
		}
	}
	length := func() int {
		var n int64
		get(&n)
		if n < 0 || n > 1<<31 {
			readErr = fmt.Errorf("invalid array length %d", n)
			return 0
		}
		return int(n)
	}
	var n int64
	get(&n)
	if n <= 0 || n > 1<<31 {
		return nil, nil, fmt.Errorf("invalid neuron count %d", n)
	}
	g := &Graph{N: int(n)}
	g.Start = make([]int64, length())
	get(g.Start)
	g.Target = make([]int32, length())
	get(g.Target)
	g.Weight = make([]float32, length())
	get(g.Weight)
	g.InDegree = make([]float32, length())
	get(g.InDegree)
	g.Region = make([]uint8, length())
	get(g.Region)
	pools := length()
	g.SensorPools = make([][]int32, pools)
	for i := range g.SensorPools {
		g.SensorPools[i] = make([]int32, length())
		get(g.SensorPools[i])
	}
	g.Readout = make([]int32, length())
	get(g.Readout)
	if readErr != nil {
		return nil, nil, fmt.Errorf("graph.bin: %w", readErr)
	}
	b, err := os.ReadFile(filepath.Join(dir, "meta.json"))
	if err != nil {
		return nil, nil, fmt.Errorf("read connectome metadata: %w", err)
	}
	var meta Meta
	if err = json.Unmarshal(b, &meta); err != nil {
		return nil, nil, fmt.Errorf("parse connectome metadata: %w", err)
	}
	if len(g.Start) != g.N+1 || len(g.Target) != len(g.Weight) {
		return nil, nil, fmt.Errorf("invalid graph dimensions")
	}
	return g, &meta, nil
}
