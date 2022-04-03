package modules

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/d5/tengo/v2"
	"github.com/sbaeurle/comb/metrics/config"
	"github.com/sbaeurle/comb/metrics/outputs"
)

type Script struct {
	mu      sync.Mutex
	log     config.Logger
	cfg     config.EndpointConfig
	input   chan []byte
	storage map[string][]float64
	outputz []outputs.Output
}

func init() {
	Modules["SCRIPT"] = NewScript
}

func NewScript(log config.Logger, cfg config.EndpointConfig, input chan []byte) Module {
	return &Script{log: log, cfg: cfg, input: input}
}

func (s *Script) StartMeasurement(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.storage = make(map[string][]float64)

	s.outputz = make([]outputs.Output, 0)
	for _, v := range s.cfg.Outputs {
		out := outputs.Outputz[filepath.Ext(v)]
		tmp, err := out(s.log, v, path, s.cfg.Fields, s.cfg.Header)
		if err != nil {
			return err
		}
		s.outputz = append(s.outputz, tmp)
	}
	return nil
}

func (s *Script) AddMeasurements() {
	for v := range s.input {
		var r map[string]interface{}
		err := json.Unmarshal(v, &r)
		if err != nil {
			s.log.Error(err)
			continue
		}

		file, err := os.Open(s.cfg.Config["ScriptPath"])
		if err != nil {
			s.log.Error(err)
			continue
		}

		tmp, err := io.ReadAll(file)
		if err != nil {
			s.log.Error(err)
			continue
		}
		file.Close()

		script := tengo.NewScript(tmp)
		script.Add("input", r)

		compiled, err := script.Run()
		if err != nil {
			s.log.Error(err)
			continue
		}
		tmp2 := compiled.Get("output").Map()
		output := make(map[string]float64)
		for key, value := range tmp2 {
			switch value := value.(type) {
			case float64:
				output[key] = value
			}
		}

		for _, out := range s.outputz {
			out.WriteResult(output)
		}

	}
}

func (s *Script) CollectMetrics() (map[string]float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make(map[string]float64)

	for k, v := range s.storage {
		tmp := calculateAggregations(v, k, s.cfg.Metrics[k])
		for m, a := range tmp {
			out[m] = a
		}
	}
	return out, nil
}
