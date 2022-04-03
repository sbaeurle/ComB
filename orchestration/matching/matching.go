package matching

import (
	"github.com/sbaeurle/comb/orchestration/config"
)

func GenerateSchedules(cfg config.Config) []map[string]string {
	if len(cfg.Workload) == 0 || len(cfg.NodeGroups) == 0 {
		return nil
	}

	var results []map[string]string

	// generate adjacency matrix for graph connecting workloads and nodes
	graph := generateGraph(cfg)

	// generate all possible scheduling options
	options := generateOptions(graph, nil)

	// filter possible scheduling options to not exceed calculated job limits
	for _, option := range options {
		result := make(map[string]string)
		jobs := make(map[string]int)
		for wl, n := range option {
			workload := cfg.Workload[wl]
			node := cfg.NodeGroups[n]
			if jobs[node.Name] < node.GetCapacity() {
				result[workload.Name] = node.Name
				jobs[node.Name] += 1
			}
		}

		if len(result) == len(cfg.Workload) {
			results = append(results, result)
		}
	}

	return results
}

func generateGraph(cfg config.Config) [][]int {
	var graph [][]int = make([][]int, len(cfg.Workload))

	for i, workload := range cfg.Workload {
		graph[i] = make([]int, len(cfg.NodeGroups))
		for n, node := range cfg.NodeGroups {
			if checkWorkloadNodeMatching(workload, node) {
				graph[i][n] = 1
			}

		}
	}

	return graph
}

func checkWorkloadNodeMatching(workload config.WorkloadConfig, node config.NodeGroup) bool {
	architecture := false
	capabilities := false
	for _, arch := range workload.Arch {
		if arch == node.Arch {
			architecture = true
		}
	}

	for _, needs := range workload.Tags {
		for _, has := range node.Capabilities {
			if needs == has {
				capabilities = true
				break
			}
		}
	}

	return architecture && capabilities
}

func generateOptions(graph [][]int, path []int) [][]int {
	var schedules [][]int

	if len(graph) == 0 {
		schedules = append(schedules, path)
		return schedules
	}
	tmp := make([]int, len(path))
	copy(tmp, path)

	for y, val := range graph[0] {
		if val == 1 {
			schedules = append(schedules, generateOptions(graph[1:], append(tmp, y))...)
		}
	}

	return schedules
}
