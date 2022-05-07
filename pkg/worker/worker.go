package worker

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/koltyakov/gosip"
	"github.com/koltyakov/gosip/api"
	"github.com/koltyakov/sp-time-machine/pkg/config"
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
			Delete: func(ctx context.Context, ids []int) error {
				fmt.Printf("Deletes %s: %+v\n", listName, ids)
				return nil
			},
		}

		// newState
		_, err := spsync.Run(ctx, options)
		if err != nil {
			return fmt.Errorf("error syncing \"%s\": %s", listName, err)
		}

		log.Infof("List \"%s\" sync completed in %f s", listName, time.Since(start).Seconds())
	}

	log.Infof("Sync session completed in %f s", time.Since(startDate).Seconds())

	return nil
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
