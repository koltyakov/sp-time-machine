package handlers

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/koltyakov/sp-time-machine/pkg/utils"
	"github.com/koltyakov/sp-time-machine/pkg/worker"

	log "github.com/sirupsen/logrus"
)

var functionTimeout = 600

func (h *Handlers) Sync(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorln(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.WithTimeout(time.Duration(functionTimeout-10)*time.Second, func(done context.CancelFunc) {
		if err := worker.Run(); err != nil {
			log.Errorf("Error: %s\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		done()
	})

	response := &InvokeResponse{
		Outputs:     map[string]interface{}{"result": "ok"},
		ReturnValue: string(data),
		Logs:        []string{},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
