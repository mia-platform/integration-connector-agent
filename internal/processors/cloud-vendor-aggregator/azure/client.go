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

package azure

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mia-platform/integration-connector-agent/internal/azure"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/commons"
)

type AzureClient struct {
	client      azure.ClientInterface
	eventSource string
}

func NewClient(getter azure.ClientInterface, source string) *AzureClient {
	return &AzureClient{
		client:      getter,
		eventSource: source,
	}
}

func (a *AzureClient) GetData(ctx context.Context, event *azure.ActivityLogEventRecord) ([]byte, error) {
	// it cannot fail because the event is already validated from the main processor
	data, _ := json.Marshal(event)
	entity, found := event.Properties["entity"]
	if !found {
		return nil, fmt.Errorf("entity not found in event properties")
	}

	resource, err := a.client.GetByID(ctx, entity.(string), apiVersionForSource(a.eventSource))
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

func apiVersionForSource(source string) string {
	// how to find the API version for a given source:
	// https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/resource-providers-and-types#azure-portal
	apiVersionsMap := map[string]string{
		azure.WebSitesEventSource:              "2024-11-01",
		azure.ComputeVirtualMachineEventSource: "2024-11-01",
		azure.ComputeDiskEventSource:           "2025-01-02",
	}
	if version, ok := apiVersionsMap[source]; ok {
		return version
	}

	return "2025-01-01" // Default API version if not found
}
