package providers

import (
	"context"

	"github.com/koltyakov/spsync"
)

// Provider sync interface
type Provider interface {
	SyncItems(ctx context.Context, entity string, items []spsync.Item) error
	DropByIDs(ctx context.Context, entity string, ids []int) error
	EnsureEntity(ctx context.Context, entity string) error
}
