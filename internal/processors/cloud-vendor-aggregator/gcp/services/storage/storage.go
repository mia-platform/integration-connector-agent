// Copyright Mia srl
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
