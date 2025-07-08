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

package functions

import (
	"context"
	"encoding/json"
	"fmt"

	azurecommons "github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/azure/commons"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/commons"
	azureactivitylogeventhubevents "github.com/mia-platform/integration-connector-agent/internal/sources/azure-activity-log-event-hub/events"
)

const (
	EventSource = "microsoft.web/sites/write"
)

type AzureFunction struct {
	client azurecommons.Client
}

func New(getter azurecommons.Client) *AzureFunction {
	return &AzureFunction{
		client: getter,
	}
}

func (a *AzureFunction) GetData(_ context.Context, event *azureactivitylogeventhubevents.ActivityLogEventRecord) ([]byte, error) {
	// it cannot fail because the event is already validated from the main processor
	data, _ := json.Marshal(event)
	entity, found := event.Properties["entity"]
	if !found {
		return nil, fmt.Errorf("entity not found in event properties")
	}

	resource, err := a.client.GetByID(entity.(string), "2024-11-01")
	if err != nil {
		return nil, fmt.Errorf("failed to get resource by ID: %w", err)
	}

	return json.Marshal(
		commons.NewAsset(resource.Name, resource.Type, commons.AzureAssetProvider).
			WithLocation(resource.Location).
			WithTags(resource.Tags).
			WithRelationships(azurecommons.RelationshipFromID(entity.(string))).
			WithRawData(data),
	)
}
