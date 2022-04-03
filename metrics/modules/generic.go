package modules

import (
	"encoding/json"
	"path/filepath"
	"sync"

	"github.com/sbaeurle/comb/metrics/config"
	"github.com/sbaeurle/comb/metrics/outputs"
)

type Generic struct {
	log     config.Logger
	mu      sync.Mutex
	cfg     config.EndpointConfig
	input   chan []byte
	storage map[string][]float64
	outputz []outputs.Output
}

func init() {
	Modules["GENERIC"] = NewGeneric
}

func NewGeneric(log config.Logger, cfg config.EndpointConfig, input chan []byte) Module {
	return &Generic{log: log, cfg: cfg, input: input}
}

func (g *Generic) StartMeasurement(path string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.storage = make(map[string][]float64)

	g.outputz = make([]outputs.Output, 0)
	for _, v := range g.cfg.Outputs {
		out := outputs.Outputz[filepath.Ext(v)]
		tmp, err := out(g.log, v, path, g.cfg.Fields, g.cfg.Header)
		if err != nil {
			return err
		}
		g.outputz = append(g.outputz, tmp)
	}
	return nil
}

func (g *Generic) AddMeasurements() {
	for v := range g.input {
		var r results
		err := json.Unmarshal(v, &r)
		if err != nil {
			g.log.Error(err)
			continue
		}

		g.mu.Lock()
		for k, v := range r {
			g.storage[k] = append(g.storage[k], v)
		}
		g.mu.Unlock()

		for _, v := range g.outputz {
			v.WriteResult(r)
		}
	}
}

func (g *Generic) CollectMetrics() (map[string]float64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	out := make(map[string]float64)

	for k, v := range g.storage {
		tmp := calculateAggregations(v, k, g.cfg.Metrics[k])
		for m, a := range tmp {
			out[m] = a
		}
	}
	return out, nil
}
