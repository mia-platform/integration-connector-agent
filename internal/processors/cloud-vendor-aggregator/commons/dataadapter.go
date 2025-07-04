package commons

import (
	"context"

	gcppubsubevents "github.com/mia-platform/integration-connector-agent/internal/sources/gcp-pubsub/events"
)

type DataAdapter interface {
	GetData(ctx context.Context, event *gcppubsubevents.InventoryEvent) ([]byte, error)
}

type Closable interface {
	Close() error
}
