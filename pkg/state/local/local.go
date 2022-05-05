package local

import (
	"encoding/json"
	"io/ioutil"

	"github.com/koltyakov/sp-time-machine/pkg/config"
	"github.com/koltyakov/sp-time-machine/pkg/state"
)

var stateFile = "state.json"

// LocalState struct
type LocalState struct {
	*state.SyncState
}

// NewState local state constructor
func NewState() (state.State, error) {
	ls := &LocalState{}
	s, err := ls.read()
	if err != nil {
		return nil, err
	}
	ls.SyncState = s
	return ls, nil
}

// Get gets state
func (ls *LocalState) Get() *state.SyncState {
	return ls.SyncState
}

// GetEnt gets entity state
func (ls *LocalState) GetList(listUri string) *state.ListState {
	return ls.Lists[listUri]
}

// Save saves state
func (ls *LocalState) Save(s *state.SyncState) error {
	file, _ := json.MarshalIndent(s, "", "  ")
	return ioutil.WriteFile(stateFile, file, 0644)
}

// SaveEnt saves entity state
func (ls *LocalState) SaveList(listUri string, entityState *state.ListState) error {
	ls.Lists[listUri] = entityState
	return ls.Save(ls.SyncState)
}

// reads state from storage
func (ls *LocalState) read() (*state.SyncState, error) {
	s := &state.SyncState{}

	bdat, _ := ioutil.ReadFile(stateFile)
	_ = json.Unmarshal(bdat, s)

	if s.Lists == nil {
		s.Lists = map[string]*state.ListState{}
	}

	for key, ent := range config.GetSettings().Lists {
		if ent.Disable {
			continue
		}
		entity, ok := s.Lists[key]
		if !ok {
			entity = &state.ListState{}
		}
		if entity.LastRun.IsZero() {
			entity.LastRun = state.DefaultStartDate()
		}
		if entity.FullSync.IsZero() {
			entity.FullSync = state.DefaultStartDate()
		}
		s.Lists[key] = entity
	}

	return s, nil
}
