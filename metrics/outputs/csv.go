package outputs

import (
	"encoding/csv"
	"fmt"
	"os"
	"sync"

	"github.com/sbaeurle/comb/metrics/config"
)

func init() {
	Outputz[".csv"] = NewCSVOutput
	Outputz[".txt"] = NewCSVOutput
}

type syncedWriter struct {
	mutex sync.Mutex
	w     *csv.Writer
}

func newSyncedWriter(filepath string) (*syncedWriter, error) {
	tmp := &syncedWriter{}
	file, err := os.Create(filepath)
	if err != nil {
		return nil, err
	}
	tmp.w = csv.NewWriter(file)

	return tmp, nil
}

func (c *syncedWriter) WriteLine(line []string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	err := c.w.Write(line)
	if err != nil {
		return err
	}
	c.w.Flush()

	return nil
}

type CSVOutput struct {
	log      config.Logger
	w        *syncedWriter
	filename string
	fields   []string
}

func NewCSVOutput(log config.Logger, filename string, filepath string, fields []string, header bool) (Output, error) {
	w, err := newSyncedWriter(fmt.Sprintf("%s/%s", filepath, filename))
	if err != nil {
		return nil, err
	}
	if header {
		w.WriteLine(fields)
	}

	return &CSVOutput{log: log, w: w, filename: filename, fields: fields}, nil
}

func (c *CSVOutput) WriteResult(out map[string]float64) error {
	// TODO: Implement proper CSV writing. Maybe even use parsed structure from configuration for performance reasons
	var tmp []string
	for _, v := range c.fields {
		if val, ok := out[v]; ok {
			tmp = append(tmp, fmt.Sprintf("%v", val))
		} else {
			tmp = append(tmp, fmt.Sprintf("%v", 0))
		}
	}
	return c.w.WriteLine(tmp)
}
