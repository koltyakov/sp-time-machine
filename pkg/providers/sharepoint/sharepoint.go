package sharepoint

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/koltyakov/gosip/api"
	"github.com/koltyakov/sp-time-machine/pkg/providers"
	"github.com/koltyakov/spsync"
)

// Client struct
type Client struct {
	sp *api.SP
}

// NewClient constructor
func NewClient(sp *api.SP) providers.Provider {
	return &Client{
		sp: sp,
	}
}

// SyncItems runs entity items batch sync
func (c *Client) SyncItems(ctx context.Context, entity string, items []spsync.Item) error {
	ent := c.sp.Web().GetList(entity).Items()

	for _, item := range items {
		data, err := ent.Filter(fmt.Sprintf("SourceID eq '%d'", item.ID)).Select("ID").Get()
		if err != nil {
			return err
		}

		d := data.Data()

		metadata, folder := c.mapPayload(entity, item)
		// fmt.Printf("%+v %s\n", metadata, folder)

		if len(d) == 0 {
			_, err := ent.AddValidate(metadata, &api.ValidateAddOptions{
				NewDocumentUpdate: true,
				DecodedPath:       folder,
			})
			if err != nil {
				fmt.Printf("Item: %+v\n", metadata)
				return err
			}
			continue
		}

		if len(d) == 1 {
			_, err := ent.GetByID(d[0].Data().ID).UpdateValidate(metadata, &api.ValidateUpdateOptions{
				NewDocumentUpdate: true,
			})
			if err != nil {
				fmt.Printf("Item: %+v\n", metadata)
				return err
			}
			continue
		}

		return fmt.Errorf("multiple items with ID %d", item.ID)
	}

	return nil
}

// DropByIDs drops items by IDs
func (c *Client) DropByIDs(ctx context.Context, entity string, ids []int) error {
	ent := c.sp.Web().GetList(entity).Items()
	limit := 10

	for i := 0; i < len(ids); i += limit {
		batch := ids[i:min(i+limit, len(ids))]

		filterConds := []string{}
		for _, id := range batch {
			filterConds = append(filterConds, fmt.Sprintf("SourceID eq %d", id))
		}

		itemsToDelete, err := ent.Select("ID").Filter(strings.Join(filterConds, " or ")).Get()
		if err != nil {
			return err
		}

		for _, item := range itemsToDelete.Data() {
			if err := ent.GetByID(item.Data().ID).Delete(); err != nil {
				return err
			}
		}
	}

	return nil
}

// EnsureEntity ensures entity sync table
func (c *Client) EnsureEntity(ctx context.Context, entity string) error {
	_, err := c.sp.Web().GetList(entity).Get()
	if err != nil {
		return err
	}

	return nil
}

var userCache = map[string]string{}

// var rootPath = ""
var jobsMap = map[string]string{}
var typesMap = map[string]string{}

func (c *Client) getUserLogin(name string) (string, error) {
	login, ok := userCache[name]
	if !ok {
		u, err := c.sp.Web().EnsureUser(name)
		if err != nil {
			return "", err
		}

		userCache[name] = u.LoginName
		login = u.LoginName
	}

	return login, nil
}

func (c *Client) mapPayload(entity string, item spsync.Item) (map[string]string, string) {
	payload := map[string]string{}

	folder := ""
	// f, ok := item.Data["FileDirRef"]
	// if ok {
	// 	if len(rootPath) == 0 {
	// 		r, _ := c.sp.Web().RootFolder().Get()
	// 		rootPath = r.Data().ServerRelativeURL
	// 	}
	// 	ff := strings.Split(f.(string), entity+"/")
	// 	if len(ff) > 1 {
	// 		folder = rootPath + entity + "/" + ff[1]
	// 	}
	// 	delete(item.Data, "FileDirRef")
	// }

	payload["SourceID"] = fmt.Sprintf("%d", item.ID)

	if item.Data["Author"] != nil {
		if author, err := c.getUserLogin(item.Data["Author"].(string)); err == nil {
			payload["Author"] = fmt.Sprintf(`[{ Key: "%s", "IsResolved": true }]`, author)
		}
		delete(item.Data, "Author")
	}

	if item.Data["Editor"] != nil {
		if editor, err := c.getUserLogin(item.Data["Editor"].(string)); err == nil {
			payload["Editor"] = fmt.Sprintf(`[{ Key: "%s", "IsResolved": true }]`, editor)
		}
		delete(item.Data, "Editor")
	}

	// https://gosamples.dev/date-time-format-cheatsheet/
	payload["Created"] = item.Created.Format("02.01.2006")
	payload["Modified"] = item.Modified.Format("02.01.2006")

	delete(item.Data, "FileDirRef")

	for k, v := range item.Data {
		if strings.Contains(k, "Id") {
			continue
		}
		switch v.(type) {
		case string:
			payload[k] = fmt.Sprintf("%s", v)
		case int:
			payload[k] = fmt.Sprintf("%d", v)
		}
	}

	if entity == "Lists/SPFTSheetsTimeEntries" {
		type Vector struct {
			ID       int
			SourceID string
		}

		if len(jobsMap) == 0 {
			var jj []Vector
			jobs, _ := c.sp.Web().GetList("Lists/SPFTSheetsJobs").Items().Select("ID,SourceID").Top(5000).Get()
			_ = json.Unmarshal(jobs.Normalized(), &jj)
			for _, j := range jj {
				jobsMap[j.SourceID] = strconv.Itoa(j.ID)
			}
		}

		if len(typesMap) == 0 {
			var tt []Vector
			types, _ := c.sp.Web().GetList("Lists/SPFTSheetsJobs1").Items().Select("ID,SourceID").Top(5000).Get()
			_ = json.Unmarshal(types.Normalized(), &tt)
			for _, j := range tt {
				typesMap[j.SourceID] = strconv.Itoa(j.ID)
			}
		}

		if item.Data["SPFTSheetsJobId"] != nil {
			payload["SPFTSheetsJob"] = jobsMap[fmt.Sprintf("%d", int(item.Data["SPFTSheetsJobId"].(float64)))]
		}
		if item.Data["SPFTSheetsJob1Id"] != nil {
			payload["SPFTSheetsJob1"] = typesMap[fmt.Sprintf("%d", int(item.Data["SPFTSheetsJob1Id"].(float64)))]
		}

		t, _ := time.Parse(time.RFC3339, item.Data["SPFTSheetsDate"].(string))
		payload["SPFTSheetsDate"] = t.Add(time.Hour * 12).Format("02.01.2006")
		payload["SPFTSheetsDuration"] = fmt.Sprintf("%d", int(item.Data["SPFTSheetsDuration"].(float64)))
	}

	return payload, folder
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}
