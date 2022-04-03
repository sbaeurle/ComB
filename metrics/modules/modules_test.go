package modules

import (
	"reflect"
	"testing"
)

func TestCalculateAggregations(t *testing.T) {
	type testCase struct {
		values       []float64
		metric       string
		aggregations []string
		out          map[string]float64
	}
	tests := map[string]testCase{
		"simple": {
			values:       []float64{1.0},
			metric:       "test",
			aggregations: []string{"MIN", "MAX", "AVG"},
			out: map[string]float64{
				"test-MIN": 1.0,
				"test-MAX": 1.0,
				"test-AVG": 1.0,
			},
		},
		"complex": {
			values:       []float64{1.0, 2.0, 3.0, 4.0, 8.0, 9.0},
			metric:       "test",
			aggregations: []string{"MIN", "MAX", "AVG", "P50"},
			out: map[string]float64{
				"test-MIN": 1.0,
				"test-MAX": 9.0,
				"test-AVG": 4.5,
				"test-P50": 4.0,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			out := calculateAggregations(tc.values, tc.metric, tc.aggregations)

			if !reflect.DeepEqual(tc.out, out) {
				t.Fatalf("expected: %v, got: %v", tc.out, out)
			}
		})
	}
}
