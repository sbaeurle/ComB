package executor

type Executor interface {
	VerifyEnvironment() []error
	RunMatching(workloads map[string]string) error
}
