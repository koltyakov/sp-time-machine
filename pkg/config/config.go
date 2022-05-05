package config

import (
	"encoding/json"
	"io/ioutil"
	"sort"

	"github.com/tidwall/jsonc"
)

var configFile = "config.jsonc"

type Settings struct {
	State string           `json:"state"`
	Lists map[string]*List `json:"lists"`
}

type List struct {
	Description string   `json:"description"`
	Fields      []string `json:"fields"`
	Disable     bool     `json:"disable"`
}

// GetSettings ...
func GetSettings() *Settings {
	settings := &Settings{}

	bdat, _ := ioutil.ReadFile(configFile)
	_ = json.Unmarshal(jsonc.ToJSON(bdat), settings)

	return settings
}

// ActiveLists gets active sync entities
func (s *Settings) ActiveLists() []string {
	entities := []string{}
	for key, ent := range s.Lists {
		if !ent.Disable {
			entities = append(entities, key)
		}
	}
	sort.Strings(entities)
	return entities
}
