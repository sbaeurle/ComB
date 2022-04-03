package modules

import (
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sbaeurle/comb/metrics/config"
	mock_config "github.com/sbaeurle/comb/metrics/config/mocks"
	"github.com/sbaeurle/comb/metrics/outputs"
	mock_outputs "github.com/sbaeurle/comb/metrics/outputs/mocks"
)

func TestScriptModule(t *testing.T) {
	type testCase struct {
		body   []byte
		errors int
		script []byte
		output map[string]float64
	}
	tests := map[string]testCase{
		"simple-script": {
			body: []byte(`
				{
					"test1": 1.0,
					"test2": 2.9
				}			
			`),
			script: []byte(`
			tmp := 0
			for x in input { tmp += x }
			output := {sum: tmp}
			`),
			output: map[string]float64{"sum": 3.9},
			errors: 0,
		},
		"syntax-error": {
			body: []byte(`
				{
					"test1": 1.0,
					"test2": 2.9
				}			
			`),
			script: []byte(`
			tmp := 0
			for in x input { tmp += x }
			output := {sum: tmp}
			`),
			errors: 1,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			path, err := os.Getwd()
			if err != nil {
				t.Error(err)
			}
			cfg := config.EndpointConfig{
				Config: map[string]string{
					"ScriptPath": path + "/test.tengo",
				},
			}
			input := make(chan []byte, 10)
			os.WriteFile(cfg.Config["ScriptPath"], tc.script, 0755)
			defer os.Remove(cfg.Config["ScriptPath"])

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockLogger := mock_config.NewMockLogger(mockCtrl)
			mockLogger.EXPECT().Error(gomock.Any()).Times(tc.errors)

			mockOutput := mock_outputs.NewMockOutput(mockCtrl)
			if tc.output != nil {
				mockOutput.EXPECT().WriteResult(tc.output).Return(nil).Times(1)
			}

			scr := Script{log: mockLogger, cfg: cfg, input: input, outputz: []outputs.Output{mockOutput}}

			go scr.AddMeasurements()

			input <- tc.body
			time.Sleep(time.Millisecond * 100)
		})
	}
}
