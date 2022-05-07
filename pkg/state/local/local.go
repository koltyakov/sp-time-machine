package local

import (
	"encoding/json"
	"io/ioutil"

	"github.com/koltyakov/sp-time-machine/pkg/config"
	"github.com/koltyakov/sp-time-machine/pkg/state"
	"github.com/koltyakov/spsync"
)

var stateFile = "state.json"

// LocalState struct
type LocalState struct {
	*state.Grid
}

// NewState local state constructor
func NewState() (state.State, error) {
	ls := &LocalState{}
	s, err := ls.read()
	if err != nil {
		return nil, err
	}
	ls.Grid = s
	return ls, nil
}

// Get gets state
func (ls *LocalState) Get() *state.Grid {
	return ls.Grid
}

// GetEnt gets entity state
func (ls *LocalState) GetList(listUri string) *state.List {
	return ls.Lists[listUri]
}

// Save saves state
func (ls *LocalState) Save(s *state.Grid) error {
	file, _ := json.MarshalIndent(s, "", "  ")
	return ioutil.WriteFile(stateFile, file, 0644)
}

// SaveEnt saves entity state
func (ls *LocalState) SaveList(listUri string, entityState *state.List) error {
	ls.Lists[listUri] = entityState
	return ls.Save(ls.Grid)
}

// reads state from storage
func (ls *LocalState) read() (*state.Grid, error) {
	s := &state.Grid{}

	bdat, _ := ioutil.ReadFile(stateFile)
	_ = json.Unmarshal(bdat, s)

	if s.Lists == nil {
		s.Lists = map[string]*state.List{}
	}

	for key, ent := range config.GetSettings().Lists {
		if ent.Disable {
			continue
		}
		entity, ok := s.Lists[key]
		if !ok {
			entity = &state.List{
				EntID:    key,
				SyncMode: spsync.Full,
				SyncDate: state.DefaultStartDate(),
			}
		}
		if entity.SyncDate.IsZero() {
			entity.SyncDate = state.DefaultStartDate()
		}
		s.Lists[key] = entity
	}

	return s, nil
}
