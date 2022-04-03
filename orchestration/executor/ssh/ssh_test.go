package ssh

import (
	"reflect"
	"testing"

	"github.com/sbaeurle/comb/orchestration/config"
)

func TestGenerateOptions(t *testing.T) {
	type testCase struct {
		matching      map[string]string
		nodeCache     map[string]*config.NodeGroup
		workloadCache map[string]*config.WorkloadConfig
		possibleTags  map[string][]string
		numberOfRuns  int
	}
	tests := map[string]testCase{
		"trivial": {
			matching: map[string]string{
				"WL1": "NG1",
				"WL2": "NG2",
			},
			nodeCache: map[string]*config.NodeGroup{
				"NG1": {Capabilities: []string{"cap1", "cap2"}},
				"NG2": {Capabilities: []string{"cap1"}},
			},
			workloadCache: map[string]*config.WorkloadConfig{
				"WL1": {Tags: []string{"cap2"}},
				"WL2": {Tags: []string{"cap1"}},
			},
			possibleTags: map[string][]string{
				"WL1": {"cap2"},
				"WL2": {"cap1"},
			},
			numberOfRuns: 1,
		},
		"complex": {
			matching: map[string]string{
				"WL1": "NG1",
				"WL2": "NG4",
				"WL3": "NG2",
				"WL4": "NG3",
			},
			nodeCache: map[string]*config.NodeGroup{
				"NG1": {Capabilities: []string{"cpu"}},
				"NG2": {Capabilities: []string{"cpu", "cuda"}},
				"NG3": {Capabilities: []string{"cpu", "opencl"}},
				"NG4": {Capabilities: []string{"cpu", "l4t"}},
			},
			workloadCache: map[string]*config.WorkloadConfig{
				"WL1": {Tags: []string{"cpu"}},
				"WL2": {Tags: []string{"cpu", "l4t"}},
				"WL3": {Tags: []string{"cpu"}},
				"WL4": {Tags: []string{"cpu", "opencl"}},
			},
			possibleTags: map[string][]string{
				"WL1": {"cpu"},
				"WL2": {"cpu", "l4t"},
				"WL3": {"cpu"},
				"WL4": {"cpu", "opencl"},
			},
			numberOfRuns: 2,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			results, runs := generateRuns(tc.matching, tc.nodeCache, tc.workloadCache)
			if tc.numberOfRuns != runs {
				t.Fatalf("expected: %v, got: %v", tc.numberOfRuns, runs)
			}
			if !reflect.DeepEqual(tc.possibleTags, results) {
				t.Fatalf("expected: %v, got: %v", tc.possibleTags, results)
			}
		})

	}
}
