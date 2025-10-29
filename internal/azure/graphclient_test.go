// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package azure

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	fakeazcore "github.com/Azure/azure-sdk-for-go/sdk/azcore/fake"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"
	fakearmresourcegraph "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph/fake"
	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGraphClient(t *testing.T) {
	t.Parallel()

	response := map[string]any{
		"id":       "resource1",
		"name":     "Resource 1",
		"type":     "Microsoft.Resources/resourceGroups",
		"location": "eastus",
		"tags": map[string]string{
			"tag1": "value1",
		},
	}

	rawResponse, err := json.Marshal(response)
	require.NoError(t, err)

	testCases := map[string]struct {
		responder      func(token *string) fakeazcore.Responder[armresourcegraph.ClientResourcesResponse]
		errorResponder fakeazcore.ErrorResponder
		expectedEvents []*entities.Event
		expectedError  error
	}{
		"only one request": {
			responder: func(token *string) fakeazcore.Responder[armresourcegraph.ClientResourcesResponse] {
				responder := fakeazcore.Responder[armresourcegraph.ClientResourcesResponse]{}
				responder.SetResponse(http.StatusOK, armresourcegraph.ClientResourcesResponse{
					QueryResponse: armresourcegraph.QueryResponse{
						Count: to.Ptr[int64](0),
						Data: []map[string]any{
							response,
						},
						ResultTruncated: to.Ptr(armresourcegraph.ResultTruncatedFalse),
					},
				}, nil)
				return responder
			},
			expectedEvents: []*entities.Event{
				{
					PrimaryKeys: []entities.PkField{
						{
							Key:   "resourceId",
							Value: "resource1",
						},
					},
					Type:          "microsoft.resources/resourcegroups",
					OperationType: entities.Write,
					OriginalRaw:   rawResponse,
				},
			},
		},
		"skip token used": {
			responder: func(token *string) fakeazcore.Responder[armresourcegraph.ClientResourcesResponse] {
				resultTruncated := to.Ptr(armresourcegraph.ResultTruncatedTrue)
				skipToken := to.Ptr("skipToken")
				if token != nil && *token == "skipToken" {
					resultTruncated = to.Ptr(armresourcegraph.ResultTruncatedFalse)
					skipToken = nil
				}

				responder := fakeazcore.Responder[armresourcegraph.ClientResourcesResponse]{}
				responder.SetResponse(http.StatusOK, armresourcegraph.ClientResourcesResponse{
					QueryResponse: armresourcegraph.QueryResponse{
						Count: to.Ptr[int64](0),
						Data: []map[string]any{
							response,
						},
						ResultTruncated: resultTruncated,
						SkipToken:       skipToken,
					},
				}, nil)
				return responder
			},
			expectedEvents: []*entities.Event{
				{
					PrimaryKeys: []entities.PkField{
						{
							Key:   "resourceId",
							Value: "resource1",
						},
					},
					Type:          "microsoft.resources/resourcegroups",
					OperationType: entities.Write,
					OriginalRaw:   rawResponse,
				},
				{
					PrimaryKeys: []entities.PkField{
						{
							Key:   "resourceId",
							Value: "resource1",
						},
					},
					Type:          "microsoft.resources/resourcegroups",
					OperationType: entities.Write,
					OriginalRaw:   rawResponse,
				},
			},
		},
		"error on request": {
			errorResponder: func() fakeazcore.ErrorResponder {
				errResponder := fakeazcore.ErrorResponder{}
				errResponder.SetError(errors.New("test error"))
				return errResponder
			}(),
			expectedError: errors.New("failed to query Azure resources: test error"),
		},
	}

	for testName, test := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			client := &GraphClient{
				client: testGraphClient(t, test.responder, test.errorResponder),
			}

			response, err := client.Resources(t.Context(), nil)
			if test.expectedError != nil {
				assert.Equal(t, test.expectedError.Error(), err.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.expectedEvents, response)
		})
	}
}

func testGraphClient(t *testing.T, responderFunc func(*string) fakeazcore.Responder[armresourcegraph.ClientResourcesResponse], errResponder fakeazcore.ErrorResponder) *armresourcegraph.Client {
	t.Helper()

	testGraphClient, err := armresourcegraph.NewClient(nil, &arm.ClientOptions{
		ClientOptions: policy.ClientOptions{
			Transport: fakearmresourcegraph.NewServerFactoryTransport(&fakearmresourcegraph.ServerFactory{
				Server: fakearmresourcegraph.Server{
					Resources: func(_ context.Context, query armresourcegraph.QueryRequest, _ *armresourcegraph.ClientResourcesOptions) (fakeazcore.Responder[armresourcegraph.ClientResourcesResponse], fakeazcore.ErrorResponder) {
						responder := fakeazcore.Responder[armresourcegraph.ClientResourcesResponse]{}
						var skipToken *string
						if query.Options != nil && query.Options.SkipToken != nil {
							skipToken = query.Options.SkipToken
						}

						if responderFunc != nil {
							responder = responderFunc(skipToken)
						}

						return responder, errResponder
					},
				},
			}),
		},
	})

	require.NoError(t, err)
	return testGraphClient
}

func TestTypeQueryFilter(t *testing.T) {
	t.Parallel()
	testCases := map[string]struct {
		types          []string
		expectedString string
	}{
		"empty types": {
			types:          []string{},
			expectedString: "",
		},
		"nil types": {
			types:          nil,
			expectedString: "",
		},
		"single type": {
			types:          []string{"Microsoft.Compute/virtualMachines"},
			expectedString: "| where ((type in~ ('Microsoft.Compute/virtualMachines')) or (isempty(type)))",
		},
		"multiple types": {
			types:          []string{"Microsoft.Compute/virtualMachines", "Microsoft.Storage/storageAccounts"},
			expectedString: "| where ((type in~ ('Microsoft.Compute/virtualMachines','Microsoft.Storage/storageAccounts')) or (isempty(type)))",
		},
	}

	for testName, test := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			result := typeQueryFilter(test.types)
			assert.Equal(t, test.expectedString, result)
		})
	}
}
