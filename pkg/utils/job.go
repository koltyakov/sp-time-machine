package utils

import (
	"context"
	"time"
)

// WithTimeout run a job with timeout
func WithTimeout(limit time.Duration, job func(done context.CancelFunc)) {
	ctx, cancel := context.WithTimeout(context.Background(), limit)
	defer cancel()
	go job(cancel)
	<-ctx.Done()
	if err := ctx.Err(); err == nil || err.Error() == "context canceled" {
		return
	}
}
