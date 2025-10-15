// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package azure

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	fakeazcore "github.com/Azure/azure-sdk-for-go/sdk/azcore/fake"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources/v3"
	fakearmresources "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources/v3/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAPIVersion = "2025-01-01"
)

func TestClient(t *testing.T) {
	t.Parallel()

	resourceID := "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testGroup/providers/Microsoft.Test/testResources/testResource"
	testCases := map[string]struct {
		responder        fakeazcore.Responder[armresources.ClientGetByIDResponse]
		errorResponder   fakeazcore.ErrorResponder
		expectedResource *Resource
		expectedError    error
	}{
		"successful response": {
			responder: func() fakeazcore.Responder[armresources.ClientGetByIDResponse] {
				responder := fakeazcore.Responder[armresources.ClientGetByIDResponse]{}
				responder.SetResponse(http.StatusOK, armresources.ClientGetByIDResponse{
					GenericResource: armresources.GenericResource{
						Name:     to.Ptr("testResource"),
						Type:     to.Ptr("Microsoft.Test/testResources"),
						Location: to.Ptr("eastus"),
						Tags: map[string]*string{
							"tagName":  to.Ptr("tagValue"),
							"tagName2": nil,
						},
					},
				}, nil)
				return responder
			}(),
			expectedResource: &Resource{
				Name:     "testResource",
				Type:     "Microsoft.Test/testResources",
				Location: "eastus",
				Tags: map[string]string{
					"tagName":  "tagValue",
					"tagName2": "",
				},
			},
		},
		"resource not found": {
			errorResponder: func() fakeazcore.ErrorResponder {
				errResponder := fakeazcore.ErrorResponder{}
				errResponder.SetError(errors.New("resource not found"))
				return errResponder
			}(),
			expectedError: errors.New("resource not found"),
		},
	}

	for testName, test := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			client := &Client{
				armClient: testArmClient(t, test.responder, test.errorResponder),
			}

			response, err := client.GetByID(t.Context(), resourceID, testAPIVersion)
			if test.expectedError != nil {
				assert.Equal(t, test.expectedError.Error(), err.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.expectedResource, response)
		})
	}
}

func testArmClient(t *testing.T, responder fakeazcore.Responder[armresources.ClientGetByIDResponse], errResponder fakeazcore.ErrorResponder) *armresources.Client {
	t.Helper()

	testArmClient, err := armresources.NewClient("", nil, &arm.ClientOptions{
		ClientOptions: policy.ClientOptions{
			Transport: fakearmresources.NewServerFactoryTransport(&fakearmresources.ServerFactory{
				Server: fakearmresources.Server{
					GetByID: func(_ context.Context, _, apiVersion string, _ *armresources.ClientGetByIDOptions) (resp fakeazcore.Responder[armresources.ClientGetByIDResponse], errResp fakeazcore.ErrorResponder) {
						require.Equal(t, testAPIVersion, apiVersion)
						return responder, errResponder
					},
				},
			}),
		},
	})

	require.NoError(t, err)
	return testArmClient
}
