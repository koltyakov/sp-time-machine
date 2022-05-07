package state

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/koltyakov/sp-time-machine/pkg/config"
	"github.com/koltyakov/spsync"

	"github.com/araddon/dateparse"
)

var defaultLastRun = time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)

// Grid struct
type Grid struct {
	Lists map[string]*List `json:"lists"`
}

// List struct
type List struct {
	EntID       string      `json:"entId"`
	SyncMode    spsync.Mode `json:"syncMode"`
	SyncDate    time.Time   `json:"syncDate"`
	SyncStage   string      `json:"syncStage"`
	ChangeToken string      `json:"changeToken"`
	PageToken   string      `json:"pageToken"`
	Fails       int         `json:"fails"`

	Lock *time.Time `json:"sessionLock,omitempty"`
	MD5  string     `json:"md5"`
}

// State interface
type State interface {
	Get() *Grid
	GetList(listUri string) *List
	Save(state *Grid) error
	SaveList(listUri string, listState *List) error
}

// DefaultStartDate ...
func DefaultStartDate() time.Time {
	if d, err := dateparse.ParseAny(os.Getenv("SYNC_START_DATE")); err == nil {
		return d
	}
	return defaultLastRun
}

// CheckSum calculates settings checksum
func CheckSum(listName string) string {
	s := config.GetSettings()
	e, ok := s.Lists[listName]
	if !ok {
		return ""
	}
	m := map[string]interface{}{}
	m["select"] = e.Select
	m["expand"] = e.Expand
	bytes, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	checkSum := md5.Sum(bytes)
	return fmt.Sprintf("%x", checkSum)
}

// ListStateToMap ...
func ListStateToMap(s *List) map[string]interface{} {
	m := map[string]interface{}{}
	b, _ := json.Marshal(s)
	_ = json.Unmarshal(b, &m)
	return m
}

// ListStateFromMap ...
func ListStateFromMap(m map[string]interface{}) *List {
	s := &List{}
	b, _ := json.Marshal(m)
	_ = json.Unmarshal(b, &s)
	return s
}

// Lists gets sync entities slice
func Lists(s *Grid) []string {
	lists := []string{}
	for key := range s.Lists {
		lists = append(lists, key)
	}
	sort.Strings(lists)
	return lists
}
