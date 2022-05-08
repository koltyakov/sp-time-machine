package worker

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/koltyakov/gosip"
	"github.com/koltyakov/gosip/api"
	"github.com/koltyakov/sp-time-machine/pkg/config"
	provider "github.com/koltyakov/sp-time-machine/pkg/providers/csv"
	"github.com/koltyakov/sp-time-machine/pkg/state"
	"github.com/koltyakov/spsync"

	strategy "github.com/koltyakov/gosip/auth/saml"
	"github.com/koltyakov/gosip/cpass"
	log "github.com/sirupsen/logrus"
)

func Run() error {
	startDate := time.Now()
	ctx := context.Background()

	settings := config.GetSettings()

	syncState, err := NewState(settings)
	if err != nil {
		return fmt.Errorf("can't initiate state provider: %s", err)
	}

	sp, err := getSourceSP()
	if err != nil {
		return fmt.Errorf("can't initiate SharePoint source: %s", err)
	}

	if _, err := sp.ContextInfo(); err != nil {
		return fmt.Errorf("can't connect to SharePoint: %s", err)
	}

	for _, listName := range settings.ActiveLists() {
		start := time.Now()

		// if err := syncState.Lock(listName); err != nil {
		// 	log.Warn(err)
		// 	continue
		// }

		client := provider.NewClient("./data")

		entState := mapListState(listName, syncState.GetList(listName))
		entMD5 := state.CheckSum(listName)

		e := settings.Lists[listName]
		entConf := &spsync.Ent{
			Select: e.Select,
			Expand: e.Expand,
			Top:    e.Top,
		}

		options := &spsync.Options{
			SP:    sp,
			State: entState,
			Ent:   entConf,
			Upsert: func(ctx context.Context, items []spsync.Item) error {
				fmt.Printf("Upserts %s: %d\n", listName, len(items))
				for i := range items {
					if _, ok := items[i].Data["Author"]; ok {
						items[i].Data["Author"] = items[i].Data["Author"].(map[string]interface{})["Title"]
						delete(items[i].Data, "Author@odata.navigationLinkUrl")
					}
					if _, ok := items[i].Data["Editor"]; ok {
						items[i].Data["Editor"] = items[i].Data["Editor"].(map[string]interface{})["Title"]
						delete(items[i].Data, "Editor@odata.navigationLinkUrl")
					}
				}
				_ = client.SyncItems(ctx, listName, items)
				return nil
			},
			Delete: func(ctx context.Context, ids []int) error {
				fmt.Printf("Deletes %s: %d\n", listName, len(ids))
				_ = client.DropByIDs(ctx, listName, ids)
				return nil
			},
			Persist: func(s *spsync.State) {
				if err := syncState.SaveList(listName, mapSyncState(listName, s, entMD5)); err != nil {
					log.Errorf("Can't persist state for List \"%s\": %s", listName, err)
				}
			},
			Events: &spsync.Events{
				FullSyncStarted: func(entity string, isBlank bool) {
					if isBlank {
						log.Infof("Full sync started for List \"%s\"", entity)
					} else {
						log.Infof("Full sync continued for List \"%s\"", entity)
					}
				},
				FullSyncFinished: func(entity string, isBlank bool) {
					log.Infof("Full sync finished for List \"%s\"", entity)
				},
				IncrSyncStarted: func(entity string) {
					log.Infof("Incr sync started for List \"%s\"", entity)
				},
				IncrSyncFinished: func(entity string) {
					log.Infof("Incr sync finished for List \"%s\"", entity)
				},
			},
		}

		_ = client.EnsureEntity(ctx, listName)

		n, err := spsync.Run(ctx, options)
		if err != nil {
			n.Fails += 1
			if err := syncState.SaveList(listName, mapSyncState(listName, n, entMD5)); err != nil {
				log.Errorf("Can't persist state for List \"%s\": %s", listName, err)
			}
			return fmt.Errorf("error syncing \"%s\": %s", listName, err)
		}

		if err := syncState.SaveList(listName, mapSyncState(listName, n, entMD5)); err != nil {
			log.Errorf("Can't persist state for List \"%s\": %s", listName, err)
		}

		log.Infof("List \"%s\" sync completed in %f s", listName, time.Since(start).Seconds())
	}

	log.Infof("Sync session completed in %f s", time.Since(startDate).Seconds())

	return nil
}

func mapListState(listName string, s *state.List) *spsync.State {
	syncState := &spsync.State{
		EntID:       listName,
		SyncMode:    s.SyncMode,
		SyncDate:    s.SyncDate,
		SyncStage:   s.SyncStage,
		ChangeToken: s.ChangeToken,
		PageToken:   s.PageToken,
		Fails:       s.Fails,
	}

	hash := state.CheckSum(listName)
	if hash != s.MD5 {
		// Sync profile changes detected which needs full sync downgrade
		syncState.SyncMode = spsync.Full
		syncState.SyncStage = ""
		syncState.ChangeToken = ""
		syncState.PageToken = ""
	}

	return syncState
}

func mapSyncState(listName string, s *spsync.State, hash string) *state.List {
	return &state.List{
		EntID:       listName,
		SyncMode:    s.SyncMode,
		SyncDate:    s.SyncDate,
		SyncStage:   s.SyncStage,
		ChangeToken: s.ChangeToken,
		PageToken:   s.PageToken,
		Fails:       s.Fails,
		MD5:         hash,
	}
}

func getSourceSP() (*api.SP, error) {
	// _ = godotenv.Load()

	c := cpass.Cpass(os.Getenv("SP_MASTER_KEY"))
	password, _ := c.Decode(os.Getenv("SP_SOURCE_PASSWORD"))

	auth := &strategy.AuthCnfg{
		SiteURL:  os.Getenv("SP_SOURCE_SITE_URL"),
		Username: os.Getenv("SP_SOURCE_USERNAME"),
		Password: password,
	}

	client := &gosip.SPClient{AuthCnfg: auth}
	return api.NewSP(client), nil
}
