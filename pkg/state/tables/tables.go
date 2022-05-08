package tables

import (
	"fmt"
	"time"

	"github.com/koltyakov/sp-time-machine/pkg/config"
	"github.com/koltyakov/sp-time-machine/pkg/state"

	"github.com/Azure/azure-sdk-for-go/storage"
)

var (
	tableName    = "state"
	partitionKey = "DEFAULT" // ToDo: To support multi tenant state in the same table provide partition key as init variable
	clients      = map[string]storage.Client{}
)

// TablesState struct
type TablesState struct {
	*state.Grid
	connectionString string
}

// NewState Azure Starage Account Tables state constructor
func NewState(connectionString string) (state.State, error) {
	ls := &TablesState{
		connectionString: connectionString,
	}
	s, err := ls.read()
	if err != nil {
		return nil, err
	}
	ls.Grid = s
	return ls, nil
}

// Get gets state
func (t *TablesState) Get() *state.Grid {
	return t.Grid
}

// GetEnt gets entity state
func (t *TablesState) GetList(listUri string) *state.List {
	return t.Lists[listUri]
}

// Save saves state
func (t *TablesState) Save(s *state.Grid) error {
	client, err := t.getClient()
	if err != nil {
		return err
	}
	tableService := client.GetTableService()
	table := tableService.GetTableReference(tableName)

	for _, chunk := range chunkStringSlice(state.Lists(s), 100) {
		batch := table.NewBatch()
		for _, listUri := range chunk {
			entityState := s.Lists[listUri]
			entity := table.GetEntityReference(partitionKey, listUri)
			entity.Properties = state.ListStateToMap(entityState)
			entity.TimeStamp = time.Now()
			batch.InsertOrReplaceEntityByForce(entity)
		}
		if err := batch.ExecuteBatch(); err != nil {
			return err
		}
	}

	return nil
}

// SaveEnt saves entity state
func (t *TablesState) SaveList(listUri string, entityState *state.List) error {
	t.Lists[listUri] = entityState

	client, err := t.getClient()
	if err != nil {
		return err
	}
	tableService := client.GetTableService()
	table := tableService.GetTableReference(tableName)

	entity := table.GetEntityReference(partitionKey, listUri)
	entity.Properties = state.ListStateToMap(entityState)
	entity.TimeStamp = time.Now()
	return entity.InsertOrReplace(nil)
}

// Lock locks entity sync for other clients
func (t *TablesState) Lock(listUri string) error {
	// ToDo: Implement entity sync locking
	return nil
}

// Unlock unlocks entity sync for other clients
func (t *TablesState) Unlock(listUri string) error {
	// ToDo: Implement entity sync unlocking
	return nil
}

// reads state from storage
func (t *TablesState) read() (*state.Grid, error) {
	s := &state.Grid{
		Lists: map[string]*state.List{},
	}

	client, err := t.getClient()
	if err != nil {
		return nil, err
	}
	tableService := client.GetTableService()
	table := tableService.GetTableReference(tableName)

	_ = table.Create(30, storage.EmptyPayload, nil) // ignore error if a table already exist

	res, err := table.QueryEntities(30, storage.MinimalMetadata, &storage.QueryOptions{
		Top:    1000,
		Filter: fmt.Sprintf("PartitionKey eq '%s'", partitionKey),
	})
	if err != nil {
		return nil, err
	}

	for _, row := range res.Entities {
		s.Lists[row.RowKey] = state.ListStateFromMap(row.Properties)
	}

	for key, ent := range config.GetSettings().Lists {
		if ent.Disable {
			continue
		}
		entity, ok := s.Lists[key]
		if !ok {
			entity = &state.List{}
		}
		if entity.SyncDate.IsZero() {
			entity.SyncDate = state.DefaultStartDate()
		}
		s.Lists[key] = entity
	}

	return s, nil
}

// getClient gets cached client
func (t *TablesState) getClient() (storage.Client, error) {
	client, ok := clients[t.connectionString]
	if ok {
		return client, nil
	}

	client, err := storage.NewClientFromConnectionString(t.connectionString)
	if err != nil {
		return client, err
	}

	clients[t.connectionString] = client

	return client, nil
}
