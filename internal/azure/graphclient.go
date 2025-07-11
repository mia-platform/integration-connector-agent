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

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"
	"github.com/mia-platform/integration-connector-agent/entities"
)

type GraphClientInterface interface {
	Resources(ctx context.Context) ([]*entities.Event, error)
}

type GraphClient struct {
	client *armresourcegraph.Client
}

func NewGraphClient(config AuthConfig) (GraphClientInterface, error) {
	var credentials azcore.TokenCredential
	var client *armresourcegraph.Client
	var err error

	if credentials, err = config.AzureTokenProvider(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrClientInitialization, err)
	}

	if client, err = armresourcegraph.NewClient(credentials, nil); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrClientInitialization, err)
	}

	return &GraphClient{
		client: client,
	}, nil
}

func (c *GraphClient) Resources(ctx context.Context) ([]*entities.Event, error) {
	listResourceQuery := "Resources | project id, name, type, location, tags"
	queryRequest := armresourcegraph.QueryRequest{
		Query: &listResourceQuery,
	}

	returnedResults := make([]*entities.Event, 0)
	for {
		results, err := c.client.Resources(ctx, queryRequest, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to query Azure resources: %w", err)
		}

		castedResults, ok := results.Data.([]any)
		if !ok {
			return nil, fmt.Errorf("unexpected data type in query results: %T", results.Data)
		}

		for _, result := range castedResults {
			resource, ok := result.(map[string]any)
			if !ok {
				continue
			}

			data, err := json.Marshal(resource)
			if err != nil {
				continue
			}

			returnedResults = append(returnedResults, &entities.Event{
				PrimaryKeys:   primaryKeys(resource["id"].(string)),
				Type:          EventTypeFromLiveLoad.String(),
				OperationType: entities.Write,
				OriginalRaw:   data,
			})
		}

		if results.ResultTruncated != nil && *results.ResultTruncated == armresourcegraph.ResultTruncatedFalse {
			break
		}
		queryRequest.Options = &armresourcegraph.QueryRequestOptions{
			SkipToken: results.SkipToken,
		}
	}

	return returnedResults, nil
}
