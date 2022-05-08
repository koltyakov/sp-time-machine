package csv

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/koltyakov/sp-time-machine/pkg/providers"
	"github.com/koltyakov/sp-time-machine/pkg/utils"
	"github.com/koltyakov/spsync"
)

// Client struct
type Client struct {
	folderPath string
}

// NewClient constructor
func NewClient(folderPath string) providers.Provider {
	return &Client{
		folderPath: folderPath,
	}
}

// SyncItems runs entity items batch sync
func (c *Client) SyncItems(ctx context.Context, entity string, items []spsync.Item) error {
	csvFilePath := c.getFilePath(entity)

	data := [][]string{}

	for _, item := range items {
		// ToDo: Delete a line of append
		jsonBytes, _ := json.Marshal(item.Data)
		data = append(data, []string{
			fmt.Sprintf("%d", item.ID),
			item.Modified.UTC().Format("2006-01-02T15:04:05.000Z"),
			string(jsonBytes),
		})
	}

	file, err := os.OpenFile(csvFilePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, value := range data {
		if err := writer.Write(value); err != nil {
			return err
		}
	}

	return nil
}

// DropByIDs drops items by IDs
func (c *Client) DropByIDs(ctx context.Context, entity string, ids []int) error {
	if len(ids) == 0 {
		return nil
	}

	csvFilePath := c.getFilePath(entity)
	if _, err := os.Stat(csvFilePath); os.IsNotExist(err) {
		return nil
	}

	file, err := os.Open(csvFilePath)
	if err != nil {
		return err
	}
	data, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return err
	}
	file.Close()

	_ = os.Remove(csvFilePath)
	file, err = os.Create(csvFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for i, value := range data {
		id, _ := strconv.Atoi(value[0])
		if i > 0 && utils.IndexOfIntArr(ids, id) != -1 {
			continue
		}
		if err := writer.Write(value); err != nil {
			return err
		}
	}

	return nil
}

// EnsureEntity ensures entity sync table
func (c *Client) EnsureEntity(ctx context.Context, entity string) error {
	csvFilePath := c.getFilePath(entity)

	headers := []string{"id", "modified", "data"}

	if _, err := os.Stat(c.folderPath); os.IsNotExist(err) {
		os.MkdirAll(c.folderPath, 0700)
	}

	if _, err := os.Stat(csvFilePath); os.IsNotExist(err) {
		file, err := os.Create(csvFilePath)
		if err != nil {
			return err
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		if err := writer.Write(headers); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) getFilePath(entity string) string {
	fileName := strings.Replace(entity, "/", "_", -1)
	csvFilePath := path.Join(c.folderPath, fileName+".csv")
	return csvFilePath
}
