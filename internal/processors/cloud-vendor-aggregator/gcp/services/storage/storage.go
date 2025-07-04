package storage

import (
	"context"
	"encoding/json"

	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/commons"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/gcp/clients/storage"
	gcppubsubevents "github.com/mia-platform/integration-connector-agent/internal/sources/gcp-pubsub/events"
)

const (
	StorageAssetType = "storage.googleapis.com/Bucket"
)

type GCPStorageDataAdapter struct {
	client storage.Client
}

func NewGCPRunServiceDataAdapter(ctx context.Context, client storage.Client) commons.DataAdapter {
	return &GCPStorageDataAdapter{
		client: client,
	}
}

func (g *GCPStorageDataAdapter) GetData(ctx context.Context, event *gcppubsubevents.InventoryEvent) ([]byte, error) {
	// it cannot fail because the event is already validated from the main processor
	data, _ := json.Marshal(event)

	bucket, err := g.client.GetBucket(ctx, event.Asset.Name)
	if err != nil {
		return nil, err
	}

	asset := &commons.Asset{
		Name:          bucket.Name,
		Type:          event.Asset.AssetType,
		Provider:      commons.GCPAssetProvider,
		Location:      bucket.Location,
		Tags:          bucket.Labels,
		Relationships: event.Asset.Ancestors,
		RawData:       data,
	}

	return json.Marshal(asset)
}
