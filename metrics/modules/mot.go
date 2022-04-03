package modules

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/sbaeurle/comb/metrics/config"
	"github.com/sbaeurle/comb/metrics/outputs"
)

type MOT struct {
	mu      sync.Mutex
	log     config.Logger
	cfg     config.EndpointConfig
	path    string
	input   chan []byte
	outputz []outputs.Output
}

type mot struct {
	Count      int         `json:"count"`
	Detections []detection `json:"detections"`
}

type detection struct {
	ID        int     `json:"id"`
	Class     string  `json:"class"`
	Conf      float64 `json:"conf"`
	BB_left   int     `json:"bb_left"`
	BB_top    int     `json:"bb_top"`
	BB_width  int     `json:"bb_width"`
	BB_height int     `json:"bb_height"`
}

func init() {
	Modules["MOT"] = NewMOT
}

func NewMOT(log config.Logger, cfg config.EndpointConfig, input chan []byte) Module {
	return &MOT{
		log:   log,
		cfg:   cfg,
		input: input,
	}
}

func (m *MOT) StartMeasurement(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.outputz = make([]outputs.Output, 0)
	m.path = path
	for _, v := range m.cfg.Outputs {
		out := outputs.Outputz[filepath.Ext(v)]
		tmp, err := out(m.log, v, path, m.cfg.Fields, m.cfg.Header)
		if err != nil {
			return err
		}
		m.outputz = append(m.outputz, tmp)
	}
	return nil
}

func (m *MOT) AddMeasurements() {
	for v := range m.input {
		var r mot
		err := json.Unmarshal(v, &r)
		if err != nil {
			m.log.Error(err)
			continue
		}

		m.mu.Lock()
		for _, det := range r.Detections {
			for _, out := range m.outputz {
				tmp := map[string]float64{"frame-number": float64(r.Count), "id": float64(det.ID), "bb_left": float64(det.BB_left), "bb_top": float64(det.BB_top), "bb_width": float64(det.BB_width), "bb_height": float64(det.BB_height), "conf": det.Conf, "x": -1.0, "y": -1.0, "z": -1.0}
				out.WriteResult(tmp)
			}
		}
		m.mu.Unlock()
	}
}

func (m *MOT) CollectMetrics() (map[string]float64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	out := make(map[string]float64)

	cmd := exec.Command("python3", m.cfg.Config["MotScript"], "--BENCHMARK", m.cfg.Config["Benchmark"], "--SPLIT_TO_EVAL", m.cfg.Config["SplitToEval"], "--GT_FOLDER", m.cfg.Config["GTFolder"], "--TRACKERS_FOLDER", m.path, "--PRINT_RESULTS", "False", "--OUTPUT_SUMMARY", "True", "--OUTPUT_DETAILED", "False", "--PLOT_CURVES", "False", "--TIME_PROGRESS", "False", "--PRINT_CONFIG", "False", "--SEQ_INFO", m.cfg.Config["SeqInfo"], "--SKIP_SPLIT_FOL", "True", "--TRACKERS_TO_EVAL", "", "--TRACKER_SUB_FOLDER", "")
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	f, err := os.Open(fmt.Sprintf("%s/%s", m.path, "pedestrian_summary.txt"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = ' '

	data, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	met := make(map[string]float64)
	for i := 0; i < len(data[0]); i++ {
		tmp, err := strconv.ParseFloat(data[1][i], 64)
		if err != nil {
			return nil, err
		}
		met[data[0][i]] = tmp
	}

	for name := range m.cfg.Metrics {
		out[name] = met[name]
	}

	return out, nil
}
