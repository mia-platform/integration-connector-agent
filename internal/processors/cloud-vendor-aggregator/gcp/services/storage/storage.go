// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

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

func NewGCPRunServiceDataAdapter(client storage.Client) commons.DataAdapter[gcppubsubevents.IInventoryEvent] {
	return &GCPStorageDataAdapter{
		client: client,
	}
}

func (g *GCPStorageDataAdapter) GetData(ctx context.Context, event gcppubsubevents.IInventoryEvent) ([]byte, error) {
	// it cannot fail because the event is already validated from the main processor
	data, _ := json.Marshal(event)

	bucket, err := g.client.GetBucket(ctx, event.Name())
	if err != nil {
		return nil, err
	}

	return json.Marshal(
		commons.NewAsset(bucket.Name, StorageAssetType, commons.GCPAssetProvider).
			WithLocation(bucket.Location).
			WithTags(bucket.Labels).
			WithRelationships(event.Ancestors()).
			WithRawData(data),
	)
}
