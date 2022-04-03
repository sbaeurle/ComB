//go:generate mockgen --destination mocks/mock_outputs.go github.com/sbaeurle/comb/metrics/outputs Output
package outputs

import "github.com/sbaeurle/comb/metrics/config"

type Output interface {
	WriteResult(out map[string]float64) error
}

var Outputz map[string]func(log config.Logger, filename string, filepath string, fields []string, header bool) (Output, error) = make(map[string]func(config.Logger, string, string, []string, bool) (Output, error))
