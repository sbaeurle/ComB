package config

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Panic(args ...interface{})
	Fatal(args ...interface{})
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Panicf(template string, args ...interface{})
	Fatalf(template string, args ...interface{})
}

type Config struct {
	SSH        *SSHConfig
	Workload   []WorkloadConfig
	NodeGroups []NodeGroup
	Evaluation string
}

type SSHConfig struct {
	User     string
	KeyFile  string
	Commands map[string]string
	RequiredPackages string
  	Services string
}

type NodeGroup struct {
	Name         string
	Arch         string
	Capabilities []string
	Nodes        []string
	NodeCapacity int
}

func (n *NodeGroup) GetCapacity() int {
	return len(n.Nodes) * n.NodeCapacity
}

type WorkloadConfig struct {
	Name      string
	Image     string
	Ports     []string
	Mounts    []string
	LocalData []string
	Tags      []string
	Arch      []string
	Command   string
}
