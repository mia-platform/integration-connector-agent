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

package vm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mia-platform/integration-connector-agent/internal/azure"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/commons"
)

type AzureVM struct {
	client azure.ClientInterface
}

func New(getter azure.ClientInterface) *AzureVM {
	return &AzureVM{
		client: getter,
	}
}

func (a *AzureVM) GetData(ctx context.Context, event *azure.ActivityLogEventRecord) ([]byte, error) {
	// it cannot fail because the event is already validated from the main processor
	data, _ := json.Marshal(event)
	entity, found := event.Properties["entity"]
	if !found {
		return nil, fmt.Errorf("entity not found in event properties")
	}

	resource, err := a.client.GetByID(ctx, entity.(string), "2024-11-01")
	if err != nil {
		return nil, fmt.Errorf("failed to get resource by ID: %w", err)
	}

	return json.Marshal(
		commons.NewAsset(resource.Name, resource.Type, commons.AzureAssetProvider).
			WithLocation(resource.Location).
			WithTags(resource.Tags).
			WithRelationships(azure.RelationshipFromID(entity.(string))).
			WithRawData(data),
	)
}
