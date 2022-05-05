package state

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/koltyakov/sp-time-machine/pkg/config"

	"github.com/araddon/dateparse"
)

var defaultLastRun = time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)

// SyncState struct
type SyncState struct {
	Lists map[string]*ListState `json:"lists"`
}

// ListState struct
type ListState struct {
	LastRun         time.Time  `json:"lastRun"`
	FullSync        time.Time  `json:"fullSync"`
	FullSyncSession *time.Time `json:"fullSyncSession,omitempty"`
	MD5             string     `json:"md5"`
}

// State interface
type State interface {
	Get() *SyncState
	GetList(listUri string) *ListState
	Save(state *SyncState) error
	SaveList(listUri string, listState *ListState) error
}

// IsFullSync ...
func IsFullSync(since time.Time) bool {
	return since == DefaultStartDate()
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
	m["fields"] = e.Fields
	bytes, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	checkSum := md5.Sum(bytes)
	return fmt.Sprintf("%x", checkSum)
}

// ListStateToMap ...
func ListStateToMap(s *ListState) map[string]interface{} {
	m := map[string]interface{}{}
	b, _ := json.Marshal(s)
	_ = json.Unmarshal(b, &m)
	return m
}

// ListStateFromMap ...
func ListStateFromMap(m map[string]interface{}) *ListState {
	s := &ListState{}
	b, _ := json.Marshal(m)
	_ = json.Unmarshal(b, &s)
	return s
}

// Lists gets sync entities slice
func Lists(s *SyncState) []string {
	lists := []string{}
	for key := range s.Lists {
		lists = append(lists, key)
	}
	sort.Strings(lists)
	return lists
}
