package routes

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sbaeurle/comb/metrics/config"
	"github.com/sbaeurle/comb/metrics/modules"
)

func RegisterRoutes(r *mux.Router, log config.Logger, cfg config.Config) (map[string]modules.Module, error) {
	modz := make(map[string]modules.Module)
	for _, endpoint := range cfg.Endpoints {
		mod, ok := modules.Modules[endpoint.Module]
		if !ok {
			return nil, fmt.Errorf("module %s not found", endpoint.Module)
		}

		comm := make(chan []byte, cfg.BufferSize)
		modz[endpoint.Name] = mod(log, endpoint, comm)
		go modz[endpoint.Name].AddMeasurements()

		handler := func(w http.ResponseWriter, r *http.Request) {
			// Forward HTTP Body to separate GO routine
			tmp, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			comm <- tmp
			w.WriteHeader(http.StatusCreated)
		}

		r.HandleFunc(endpoint.Url, handler).Methods("POST")
	}

	return modz, nil
}
