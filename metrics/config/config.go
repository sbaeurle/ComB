//go:generate mockgen --destination mocks/mock_config.go github.com/sbaeurle/comb/metrics/config Logger
package config

type EndpointConfig struct {
	Name    string
	Url     string
	Module  string
	Header  bool
	Config  map[string]string
	Fields  []string
	Outputs []string
	Metrics map[string][]string
}
type Config struct {
	Port           int
	BufferSize     int
	DateConfig     string
	GeneratePlots  bool
	Endpoints      []EndpointConfig
	RootFolder     string
	PlottingScript string
}
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
