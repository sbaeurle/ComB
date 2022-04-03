package modules

import (
	"fmt"
	"math"

	"github.com/sbaeurle/comb/metrics/config"
)

type results map[string]float64

type Module interface {
	StartMeasurement(path string) error
	AddMeasurements()
	CollectMetrics() (map[string]float64, error)
}

var Modules map[string]func(config.Logger, config.EndpointConfig, chan []byte) Module = make(map[string]func(config.Logger, config.EndpointConfig, chan []byte) Module)

func calculateAggregations(values []float64, metric string, aggregations []string) map[string]float64 {
	tmp := make(map[string]float64)
	for _, agg := range aggregations {
		out := 0.0
		switch agg {
		case "MIN":
			out = math.MaxFloat64
			for _, v := range values {
				out = math.Min(out, v)
			}
		case "MAX":
			out = -math.MaxFloat64
			for _, v := range values {
				out = math.Max(out, v)
			}
		case "AVG":
			for _, v := range values {
				out += v
			}
			out /= float64(len(values))
		case "P50":
			n := int(50.0 / 100.0 * float64(len(values)))
			out = values[n]
		}
		tmp[fmt.Sprintf("%s-%s", metric, agg)] = out
	}
	return tmp
}
