// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"
	"github.com/mia-platform/integration-connector-agent/entities"
)

type GraphLiveData struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Location string            `json:"location"`
	Tags     map[string]string `json:"tags"`
}

type GraphClientInterface interface {
	Resources(ctx context.Context, typesToFilter []string) ([]*entities.Event, error)
}

type GraphClient struct {
	client *armresourcegraph.Client
}

func NewGraphClient(config AuthConfig) (GraphClientInterface, error) {
	var credentials azcore.TokenCredential
	var client *armresourcegraph.Client
	var err error

	if credentials, err = config.AzureTokenProvider(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrClientInitialization, err)
	}

	if client, err = armresourcegraph.NewClient(credentials, nil); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrClientInitialization, err)
	}

	return &GraphClient{
		client: client,
	}, nil
}

func typeQueryFilter(typesToFilter []string) string {
	if len(typesToFilter) == 0 {
		return ""
	}

	return fmt.Sprintf("| where ((type in~ ('%s')) or (isempty(type)))", strings.Join(typesToFilter, "','"))
}

func (c *GraphClient) Resources(ctx context.Context, typesToFilter []string) ([]*entities.Event, error) {
	listResourceQuery := fmt.Sprintf("Resources %s | project id, name, type, location, tags", typeQueryFilter(typesToFilter))
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
