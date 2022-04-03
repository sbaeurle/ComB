package matching

import (
	"reflect"
	"testing"

	"github.com/sbaeurle/comb/orchestration/config"
)

func TestGenerateOptions(t *testing.T) {
	type testCase struct {
		graph     [][]int
		matchings [][]int
	}
	tests := map[string]testCase{
		"trivial-3x3": {
			graph: [][]int{
				{0, 1, 0},
				{1, 1, 0},
				{0, 0, 1},
			},
			matchings: [][]int{{1, 0, 2}, {1, 1, 2}},
		},
		"complex-4x4": {
			graph: [][]int{
				{1, 1, 0, 0},
				{1, 0, 1, 0},
				{0, 0, 1, 1},
				{1, 1, 1, 1},
			},
			matchings: [][]int{
				{0, 0, 2, 0}, {0, 0, 2, 1}, {0, 0, 2, 2}, {0, 0, 2, 3},
				{0, 0, 3, 0}, {0, 0, 3, 1}, {0, 0, 3, 2}, {0, 0, 3, 3},
				{0, 2, 2, 0}, {0, 2, 2, 1}, {0, 2, 2, 2}, {0, 2, 2, 3},
				{0, 2, 3, 0}, {0, 2, 3, 1}, {0, 2, 3, 2}, {0, 2, 3, 3},
				{1, 0, 2, 0}, {1, 0, 2, 1}, {1, 0, 2, 2}, {1, 0, 2, 3},
				{1, 0, 3, 0}, {1, 0, 3, 1}, {1, 0, 3, 2}, {1, 0, 3, 3},
				{1, 2, 2, 0}, {1, 2, 2, 1}, {1, 2, 2, 2}, {1, 2, 2, 3},
				{1, 2, 3, 0}, {1, 2, 3, 1}, {1, 2, 3, 2}, {1, 2, 3, 3},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := generateOptions(tc.graph, nil)
			if !reflect.DeepEqual(tc.matchings, result) {
				t.Fatalf("expected: %v, got: %v", tc.matchings, result)
			}
		})

	}
}

func TestGenerateGraph(t *testing.T) {
	type testCase struct {
		cfg   config.Config
		graph [][]int
	}
	tests := map[string]testCase{
		"empty": {
			cfg: config.Config{
				NodeGroups: []config.NodeGroup{
					{Name: "NG1", Arch: "x86", Capabilities: []string{"cpu", "cuda"}},
					{Name: "NG2", Arch: "arm64", Capabilities: []string{"cpu"}},
				},
			},
			graph: [][]int{},
		},
		"trivial": {
			cfg: config.Config{
				Workload: []config.WorkloadConfig{
					{Name: "WL1", Tags: []string{"cuda"}, Arch: []string{"x86"}},
					{Name: "WL2", Tags: []string{"cpu", "cuda"}, Arch: []string{"x86", "arm64"}},
					{Name: "WL3", Tags: []string{"cpu"}, Arch: []string{"arm64"}},
				},
				NodeGroups: []config.NodeGroup{
					{Name: "NG1", Arch: "x86", Capabilities: []string{"cpu", "cuda"}},
					{Name: "NG2", Arch: "arm64", Capabilities: []string{"cpu"}},
				},
			},
			graph: [][]int{
				{1, 0},
				{1, 1},
				{0, 1},
			},
		},
		"complex": {
			cfg: config.Config{
				Workload: []config.WorkloadConfig{
					{Name: "WL1", Tags: []string{"cpu", "opencl", "cuda"}, Arch: []string{"x86"}}, //1,2
					{Name: "WL2", Tags: []string{"cuda"}, Arch: []string{"x86", "arm64"}},         //1,3
					{Name: "WL3", Tags: []string{"cpu"}, Arch: []string{"arm64"}},                 //3,4
					{Name: "WL4", Tags: []string{"cpu"}, Arch: []string{"x86", "arm64"}},          //1,2,3,4
				},
				NodeGroups: []config.NodeGroup{
					{Name: "NG1", Arch: "x86", Capabilities: []string{"cpu", "opencl", "cuda"}},
					{Name: "NG2", Arch: "x86", Capabilities: []string{"cpu", "opencl"}},
					{Name: "NG3", Arch: "arm64", Capabilities: []string{"cpu", "opencl", "cuda"}},
					{Name: "NG4", Arch: "arm64", Capabilities: []string{"cpu"}},
				},
			},
			graph: [][]int{
				{1, 1, 0, 0},
				{1, 0, 1, 0},
				{0, 0, 1, 1},
				{1, 1, 1, 1},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := generateGraph(tc.cfg)
			if !reflect.DeepEqual(tc.graph, result) {
				t.Fatalf("expected: %v, got: %v", tc.graph, result)
			}
		})

	}
}

func TestGenerateSchedules(t *testing.T) {
	type testCase struct {
		cfg       config.Config
		schedules []map[string]string
	}
	tests := map[string]testCase{
		"empty": {
			cfg: config.Config{
				NodeGroups: []config.NodeGroup{
					{Name: "NG1", Arch: "x86", Capabilities: []string{"cpu", "cuda"}},
					{Name: "NG2", Arch: "arm64", Capabilities: []string{"cpu"}},
				},
			},
			schedules: nil,
		},
		"trivial": {
			cfg: config.Config{
				Workload: []config.WorkloadConfig{
					{Name: "WL1", Tags: []string{"cuda"}, Arch: []string{"x86"}},
					{Name: "WL2", Tags: []string{"cpu", "cuda"}, Arch: []string{"x86", "arm64"}},
					{Name: "WL3", Tags: []string{"cpu"}, Arch: []string{"arm64"}},
				},
				NodeGroups: []config.NodeGroup{
					{Name: "NG1", Arch: "x86", Capabilities: []string{"cpu", "cuda"}, Nodes: []string{"test1", "test2", "test3"}, NodeCapacity: 1},
					{Name: "NG2", Arch: "arm64", Capabilities: []string{"cpu"}, Nodes: []string{"test1"}, NodeCapacity: 1},
				},
			},
			schedules: []map[string]string{{"WL1": "NG1", "WL2": "NG1", "WL3": "NG2"}},
		},
		"complex": {
			cfg: config.Config{
				Workload: []config.WorkloadConfig{
					{Name: "WL1", Tags: []string{"cpu", "opencl", "cuda"}, Arch: []string{"x86"}}, //1,2
					{Name: "WL2", Tags: []string{"cuda"}, Arch: []string{"x86", "arm64"}},         //1,3
					{Name: "WL3", Tags: []string{"cpu"}, Arch: []string{"arm64"}},                 //3,4
					{Name: "WL4", Tags: []string{"cpu"}, Arch: []string{"x86", "arm64"}},          //1,2,3,4
				},
				NodeGroups: []config.NodeGroup{
					{Name: "NG1", Arch: "x86", Capabilities: []string{"cpu", "opencl", "cuda"}, Nodes: []string{"test1"}, NodeCapacity: 1},
					{Name: "NG2", Arch: "x86", Capabilities: []string{"cpu"}, Nodes: []string{"test1"}, NodeCapacity: 1},
					{Name: "NG3", Arch: "arm64", Capabilities: []string{"cpu", "opencl", "cuda"}, Nodes: []string{"test1"}, NodeCapacity: 1},
					{Name: "NG4", Arch: "arm64", Capabilities: []string{"cpu"}, Nodes: []string{"test1"}, NodeCapacity: 1},
				},
			},
			schedules: []map[string]string{
				{"WL1": "NG1", "WL2": "NG3", "WL3": "NG4", "WL4": "NG2"},
				{"WL1": "NG2", "WL2": "NG1", "WL3": "NG3", "WL4": "NG4"},
				{"WL1": "NG2", "WL2": "NG1", "WL3": "NG4", "WL4": "NG3"},
				{"WL1": "NG2", "WL2": "NG3", "WL3": "NG4", "WL4": "NG1"},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := GenerateSchedules(tc.cfg)
			if !reflect.DeepEqual(tc.schedules, result) {
				t.Fatalf("expected: %v, got: %v", tc.schedules, result)
			}
		})

	}
}
