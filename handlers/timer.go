package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/koltyakov/gosip"
	"github.com/koltyakov/gosip/api"
	"github.com/koltyakov/sp-time-machine/pkg/config"
	"github.com/koltyakov/sp-time-machine/pkg/utils"
	"github.com/koltyakov/sp-time-machine/pkg/worker"
	"github.com/koltyakov/spsync"

	strategy "github.com/koltyakov/gosip/auth/saml"
	"github.com/koltyakov/gosip/cpass"
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

	// run incremental sync job
	utils.WithTimeout(time.Duration(functionTimeout-10)*time.Second, func(done context.CancelFunc) {
		if err := syncJob(); err != nil {
			log.Errorf("Error: %s\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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

func syncJob() error {
	startDate := time.Now()
	ctx := context.Background()

	settings := config.GetSettings()

	syncState, err := worker.NewState(settings)
	if err != nil {
		return fmt.Errorf("can't initiate state provider: %s", err)
	}

	sp, err := getSourceSP()
	if err != nil {
		return fmt.Errorf("can't initiate source SP: %s", err)
	}

	for _, listName := range settings.ActiveLists() {
		start := time.Now()

		s := syncState.GetList(listName)
		state := &spsync.State{
			EntID:       listName,
			SyncMode:    s.SyncMode,
			SyncDate:    s.SyncDate,
			SyncStage:   s.SyncStage,
			ChangeToken: s.ChangeToken,
			PageToken:   s.PageToken,
			Fails:       s.Fails,
		}

		e := settings.Lists[listName]
		entConf := &spsync.EntConf{
			Select: e.Select,
			Expand: e.Expand,
			Top:    e.Top,
		}

		options := &spsync.Options{
			SP:      sp,
			State:   state,
			EntConf: entConf,
			Upsert: func(ctx context.Context, items []spsync.ListItem) error {
				for _, item := range items {
					fmt.Printf("Upsert %s: %+v\n", listName, item.Data)
				}
				return nil
			},
			// Provide a handler to deal with deleted items
			Delete: func(ctx context.Context, ids []int) error {
				fmt.Printf("Deletes %s: %+v\n", listName, ids)
				return nil
			},
		}

		// newState
		_, err := worker.Run(ctx, options)
		if err != nil {
			return fmt.Errorf("error syncing \"%s\": %s", listName, err)
		}

		log.Infof("List \"%s\" sync completed in %f s", listName, time.Since(start).Seconds())
	}

	log.Infof("Sync session completed in %f s", time.Since(startDate).Seconds())

	return nil
}

func getSourceSP() (*api.SP, error) {
	_ = godotenv.Load()

	c := cpass.Cpass(os.Getenv("SP_MASTERKEY"))
	password, _ := c.Decode(os.Getenv("SP_SOURCE_PASSWORD"))

	auth := &strategy.AuthCnfg{
		SiteURL:  os.Getenv("SP_SOURCE_SITE_URL"),
		Username: os.Getenv("SP_SOURCE_USERNAME"),
		Password: password,
	}

	client := &gosip.SPClient{AuthCnfg: auth}
	return api.NewSP(client), nil
}

// // syncJob job definition runs all entities sync
// func syncJob(fullSync bool) error {
// 	startDate := time.Now()
// 	settings := config.GetSettings()

// 	// Init sync state provider
// 	syncState, err := worker.NewState(settings)
// 	if err != nil {
// 		return fmt.Errorf("can't initiate state provider: %s", err)
// 	}

// 	sp1, err := sp.NewSP(os.Getenv("SP_SOURCE_CREDS"), os.Getenv("SP_MASTER_KEY"))
// 	if _, err := sp1.ContextInfo(); err != nil {
// 		return fmt.Errorf("error connecting to source SharePoint site: %s", err)
// 	}
// 	sp1 = sp1.Conf(api.HeadersPresets.Nometadata)

// 	sp2, err := sp.NewSP(os.Getenv("SP_TARGET_CREDS"), os.Getenv("SP_MASTER_KEY"))
// 	if _, err := sp2.ContextInfo(); err != nil {
// 		return fmt.Errorf("error connecting to source SharePoint site: %s", err)
// 	}
// 	sp2 = sp2.Conf(api.HeadersPresets.Nometadata)

// 	// Process incremental sync sequentially to keep predictable load during the day
// 	if !fullSync {
// 		// Sync active entities defined in config file
// 		for _, listName := range settings.ActiveLists() {
// 			start := time.Now()
// 			if err := w.Sync(sp1, sp2, listName, startDate, syncState, fullSync); err != nil {
// 				return fmt.Errorf("error syncing \"%s\": %s", listName, err)
// 			}
// 			log.Infof("List \"%s\" sync completed in %f s", listName, time.Since(start).Seconds())
// 		}
// 		log.Infof("Sync session completed in %f s", time.Since(startDate).Seconds())
// 		return nil
// 	}

// 	// For full sync running all entities in parallel. This is required to write FullSyncSession for all entities,
// 	// so after 600 sec timeout the continues or incremental sync will catch up the progress.

// 	hasError := false
// 	lists := []string{}
// 	mu := &sync.Mutex{}
// 	wg := sync.WaitGroup{}

// 	// Sync active entities defined in config file
// 	for _, listName := range settings.ActiveLists() {
// 		wg.Add(1)
// 		go func(lstName string) {
// 			defer wg.Done()
// 			start := time.Now()
// 			if err := w.Sync(sp1, sp2, lstName, startDate, syncState, fullSync); err != nil {
// 				log.Errorf("Error syncing \"%s\": %s", lstName, err)
// 				mu.Lock()
// 				hasError = true
// 				lists = append(lists, lstName)
// 				mu.Unlock()
// 			} else {
// 				log.Infof("List \"%s\" sync completed in %f s", lstName, time.Since(start).Seconds())
// 			}
// 		}(listName)
// 	}

// 	wg.Wait()

// 	if hasError {
// 		return fmt.Errorf("sync failed for the lists: %s", strings.Join(lists, ", "))
// 	}

// 	log.Infof("Sync session completed in %f s", time.Since(startDate).Seconds())
// 	return nil
// }
