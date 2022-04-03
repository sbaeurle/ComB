package modules

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sbaeurle/comb/metrics/config"
	mock_config "github.com/sbaeurle/comb/metrics/config/mocks"
	"github.com/sbaeurle/comb/metrics/outputs"
	mock_outputs "github.com/sbaeurle/comb/metrics/outputs/mocks"
)

func TestMOTAddMeasurements(t *testing.T) {
	type testCase struct {
		body   []byte
		output map[string]float64
	}
	tests := map[string]testCase{
		"simple-script": {
			body: []byte(`
				{
					"count": 1,
					"detections": [
						{
						"id": 1,
						"conf": 1,
						"bb_left": 100,
						"bb_height": 100,
						"bb_top": 100,
						"bb_width": 100
						}
					]					
				}			
			`),
			output: map[string]float64{"frame-number": 1.0, "id": 1.0, "conf": 1.0, "bb_left": 100.0, "bb_top": 100.0, "bb_height": 100.0, "bb_width": 100.0, "x": -1.0, "y": -1.0, "z": -1.0},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := config.EndpointConfig{}
			input := make(chan []byte, 10)

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockLogger := mock_config.NewMockLogger(mockCtrl)
			mockLogger.EXPECT().Error(gomock.Any()).MaxTimes(0)

			mockOutput := mock_outputs.NewMockOutput(mockCtrl)
			mockOutput.EXPECT().WriteResult(tc.output).Return(nil).Times(1)

			scr := MOT{log: mockLogger, cfg: cfg, input: input, outputz: []outputs.Output{mockOutput}}

			go scr.AddMeasurements()

			input <- tc.body
			time.Sleep(time.Millisecond * 100)
		})
	}
}

func TestMOTCollectMetrics(t *testing.T) {
	type testCase struct {
		cfg  map[string]string
		out  map[string]float64
		err  error
		path string
	}
	tests := map[string]testCase{
		"valid": {
			cfg: map[string]string{
				"MotScript":   "TrackEval/scripts/run_mot_challenge.py",
				"Benchmark":   "MOT20",
				"SplitToEval": "train",
				"SeqInfo":     "MOT20-01",
				"GTFolder":    "../data",
			},
			out: map[string]float64{
				"MOTA": -24.701,
				"MOTP": 0,
			},
			path: "testdata",
		},
		// "invalid": {
		// 	cfg: map[string]string{
		// 		"MotScript":   "TrackEval/scripts/run_mot_challenge.py",
		// 		"Benchmark":   "MOT20",
		// 		"SplitToEval": "train",
		// 		"SeqInfo":     "MOT20-011",
		// 		"GTFolder":    "../../data",
		// 	},
		// 	path: "testdata2",
		// },
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := config.EndpointConfig{
				Config: tc.cfg,
				Metrics: map[string][]string{
					"MOTA": {},
					"MOTP": {},
				},
			}
			scr := &MOT{path: tc.path, cfg: cfg}
			out, err := scr.CollectMetrics()

			if !errors.Is(tc.err, err) {
				t.Fatalf("expected: %v, got: %v", tc.err, err)
			}

			if !reflect.DeepEqual(tc.out, out) {
				t.Fatalf("expected: %v, got: %v", tc.out, out)
			}
		})
	}
}
