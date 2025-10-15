// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package consoleclient

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type testResource map[string]any

func TestCatalogApply(t *testing.T) {
	const marketplaceBaseURL = "127.0.0.1:45874"
	const tenantID = "tenant123"

	applyPath := fmt.Sprintf("/api/tenants/%s/marketplace/items", tenantID)

	client := New[testResource](fmt.Sprintf("http://%s/", marketplaceBaseURL), &mockedTokenManager{})
	item := MarketplaceResource[testResource]{
		ItemID:   "myItem",
		TenantID: tenantID,
		Name:     "myItemName",
		ItemTypeDefinitionRef: ItemTypeDefinitionRef{
			Name:      "resType",
			Namespace: "default",
		},
		Resources: testResource{
			"reskey": "resval",
		},
	}

	t.Run("returns error if the execution of request fails", func(t *testing.T) {
		itemID, err := client.Apply(t.Context(), &item)
		require.Empty(t, itemID)
		require.ErrorIs(t, err, ErrMarketplaceRequestExecution)
	})

	t.Run("returns error if the response is not 200", func(t *testing.T) {
		m := runMocha(t, marketplaceBaseURL)
		m = registerAPI(t, m,
			MockExpectation{
				path:     applyPath,
				verb:     http.MethodPost,
				tenantID: tenantID,
				headers: map[string]string{
					"content-type": "application/json",
				},
			},
			MockResponse{
				statusCode: http.StatusInternalServerError,
			},
		)

		itemID, err := client.Apply(t.Context(), &item)
		require.Empty(t, itemID)
		require.Equal(t, "failed to apply resource, status code: 500", err.Error())
		m.AssertCalled(t)
	})

	t.Run("returns error if the response body is unknown", func(t *testing.T) {
		m := runMocha(t, marketplaceBaseURL)
		m = registerAPI(t, m,
			MockExpectation{
				path:     applyPath,
				verb:     http.MethodPost,
				tenantID: tenantID,
			},
			MockResponse{
				statusCode: http.StatusOK,
				body:       []byte("unknown body"),
			},
		)

		itemID, err := client.Apply(t.Context(), &item)
		require.Empty(t, itemID)
		require.ErrorIs(t, err, ErrMarketplaceResponseParse)
		m.AssertCalled(t)
	})

	t.Run("returns error if there are validation errors", func(t *testing.T) {
		mockedValidationError1 := ValidationError{Message: "valMsg1"}
		mockedValidationError2 := ValidationError{Message: "valMsg2"}
		mockedResponse := &marketplacePostExtensionResponse{
			Done: false,
			Items: []responseItem{
				{
					ItemID: item.ItemID,
					Errors: []ValidationError{
						mockedValidationError1,
						mockedValidationError2,
					},
				},
			},
		}

		m := runMocha(t, marketplaceBaseURL)
		m = registerAPI(t, m,
			MockExpectation{
				path:     applyPath,
				verb:     http.MethodPost,
				tenantID: tenantID,
			},
			MockResponse{
				statusCode: http.StatusOK,
				body:       mockedResponse,
			},
		)

		itemID, err := client.Apply(t.Context(), &item)
		require.Empty(t, itemID)
		var parsedErr *MarketplaceValidationError
		require.ErrorAs(t, err, &parsedErr)
		require.Len(t, parsedErr.Errors, 2)
		require.Equal(t, mockedValidationError1.Message, parsedErr.Errors[0])
		require.Equal(t, mockedValidationError2.Message, parsedErr.Errors[1])
		m.AssertCalled(t)
	})

	t.Run("returns the created item id if succeed", func(t *testing.T) {
		mockedItemID := "createdItemId"
		mockedResponse := &marketplacePostExtensionResponse{
			Done:  true,
			Items: []responseItem{{ItemID: mockedItemID}},
		}

		m := runMocha(t, marketplaceBaseURL)
		m = registerAPI(t, m,
			MockExpectation{
				path:     applyPath,
				verb:     http.MethodPost,
				tenantID: tenantID,
			}, MockResponse{
				statusCode: http.StatusOK,
				body:       mockedResponse,
			},
		)

		itemID, err := client.Apply(t.Context(), &item)
		require.NoError(t, err)
		require.Equal(t, mockedItemID, itemID)
		m.AssertCalled(t)
	})

	t.Run("request body is stringified correctly", func(t *testing.T) {
		mockedItemID := "createdItemId"
		mockedResponse := &marketplacePostExtensionResponse{
			Done:  true,
			Items: []responseItem{{ItemID: mockedItemID}},
		}

		item := MarketplaceResource[testResource]{
			ItemID:   "myItem",
			Name:     "myItemName",
			TenantID: tenantID,
			ItemTypeDefinitionRef: ItemTypeDefinitionRef{
				Name:      "resType",
				Namespace: "default",
			},
			LifecycleStatus: Draft,
			Resources: testResource{
				"k1": "v1",
			},
		}

		expectedMarketplaceRequestBodyString := fmt.Sprintf("{\"resources\":[{\"description\":\"\",\"itemId\":\"myItem\",\"itemTypeDefinitionRef\":{\"name\":\"resType\",\"namespace\":\"default\"},\"lifecycleStatus\":\"draft\",\"name\":\"myItemName\",\"resources\":{\"k1\":\"v1\"},\"tenantId\":\"%s\"}]}", tenantID)

		m := runMocha(t, marketplaceBaseURL)
		m = registerAPI(t, m,
			MockExpectation{
				path:       applyPath,
				verb:       http.MethodPost,
				tenantID:   tenantID,
				bodyString: expectedMarketplaceRequestBodyString,
			},
			MockResponse{
				statusCode: http.StatusOK,
				body:       mockedResponse,
			},
		)

		itemID, err := client.Apply(t.Context(), &item)
		require.NoError(t, err)
		require.Equal(t, mockedItemID, itemID)
		m.AssertCalled(t)
	})

	t.Run("correctly sets authorization header", func(t *testing.T) {
		url := fmt.Sprintf("http://%s/", marketplaceBaseURL)
		tknmngr, err := NewClientCredentialsTokenManager(url, "myClientId", "myClientSecret")
		require.NoError(t, err)
		client := New[testResource](url, tknmngr)

		mockedItemID := "createdItemId"
		mockedResponse := &marketplacePostExtensionResponse{
			Done:  true,
			Items: []responseItem{{ItemID: mockedItemID}},
		}

		item := MarketplaceResource[testResource]{
			ItemID:   "myItem",
			Name:     "myItemName",
			TenantID: tenantID,
			ItemTypeDefinitionRef: ItemTypeDefinitionRef{
				Name:      "resType",
				Namespace: "default",
			},
			Resources: testResource{
				"k1": "v1",
			},
		}

		m := runMocha(t, marketplaceBaseURL)
		m = registerAPI(t, m,
			MockExpectation{
				path:     applyPath,
				verb:     http.MethodPost,
				tenantID: tenantID,
				headers: map[string]string{
					"Authorization": "Bearer the-new-token",
				},
			},
			MockResponse{
				statusCode: http.StatusOK,
				body:       mockedResponse,
			},
		)
		m = registerAPI(t, m,
			MockExpectation{
				path: "/api/m2m/oauth/token",
				verb: http.MethodPost,
				headers: map[string]string{
					"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte("myClientId:myClientSecret")),
					"Content-Type":  "application/x-www-form-urlencoded",
				},
			},
			MockResponse{
				statusCode: http.StatusOK,
				body: map[string]any{
					"access_token": "the-new-token",
					"token_type":   "Bearer",
					"expires_in":   3600,
				},
			},
		)

		itemID, err := client.Apply(t.Context(), &item)
		require.NoError(t, err)
		require.Equal(t, mockedItemID, itemID)
		m.AssertCalled(t)
	})

	t.Run("correctly reuses non expired authorization header", func(t *testing.T) {
		url := fmt.Sprintf("http://%s/", marketplaceBaseURL)
		tknmngr, err := NewClientCredentialsTokenManager(url, "myClientId", "myClientSecret")
		require.NoError(t, err)

		client := New[testResource](url, tknmngr)

		mockedItemID := "createdItemId"
		mockedResponse := &marketplacePostExtensionResponse{
			Done:  true,
			Items: []responseItem{{ItemID: mockedItemID}},
		}

		item1 := MarketplaceResource[testResource]{
			ItemID:   "myItem",
			Name:     "myItemName",
			TenantID: tenantID,
			ItemTypeDefinitionRef: ItemTypeDefinitionRef{
				Name:      "resType",
				Namespace: "default",
			},
			LifecycleStatus: Published,
			Resources: testResource{
				"k1": "v1",
			},
		}
		expectedMarketplaceRequestBodyString1 := fmt.Sprintf("{\"resources\":[{\"description\":\"\",\"itemId\":\"myItem\",\"itemTypeDefinitionRef\":{\"name\":\"resType\",\"namespace\":\"default\"},\"lifecycleStatus\":\"published\",\"name\":\"myItemName\",\"resources\":{\"k1\":\"v1\"},\"tenantId\":\"%s\"}]}", tenantID)

		m := runMocha(t, marketplaceBaseURL)
		m = registerAPI(t, m,
			MockExpectation{
				path: "/api/m2m/oauth/token",
				verb: http.MethodPost,
				headers: map[string]string{
					"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte("myClientId:myClientSecret")),
					"Content-Type":  "application/x-www-form-urlencoded",
				},
			},
			MockResponse{
				statusCode: http.StatusOK,
				body: map[string]any{
					"access_token": "the-new-token",
					"token_type":   "Bearer",
					"expires_in":   3600,
				},
				times: 1,
			},
		)

		m = registerAPI(t, m,
			MockExpectation{
				path:     applyPath,
				verb:     http.MethodPost,
				tenantID: tenantID,
				headers: map[string]string{
					"Authorization": "Bearer the-new-token",
				},
				bodyString: expectedMarketplaceRequestBodyString1,
			},
			MockResponse{
				statusCode: http.StatusOK,
				body:       mockedResponse,
			},
			MockResponse{
				statusCode: http.StatusOK,
				body:       mockedResponse,
			},
		)

		itemID, err := client.Apply(t.Context(), &item1)
		require.NoError(t, err)
		require.Equal(t, mockedItemID, itemID)
		m.AssertCalled(t)

		itemID, err = client.Apply(t.Context(), &item1)
		require.NoError(t, err)
		require.Equal(t, mockedItemID, itemID)
		m.AssertCalled(t)
	})
}

func TestCatalogDelete(t *testing.T) {
	const marketplaceBaseURL = "127.0.0.1:45874"
	const tenantID = "tenant123"
	const itemID = "item123"

	deletePath := fmt.Sprintf("/api/tenants/%s/marketplace/items/%s/versions/NA", tenantID, itemID)

	client := New[testResource](fmt.Sprintf("http://%s/", marketplaceBaseURL), &mockedTokenManager{})

	t.Run("returns error if the execution of request fails", func(t *testing.T) {
		err := client.Delete(t.Context(), tenantID, itemID)
		require.ErrorIs(t, err, ErrMarketplaceRequestExecution)
	})

	t.Run("returns error if the response is not 200", func(t *testing.T) {
		m := runMocha(t, marketplaceBaseURL)
		m = registerAPI(t, m,
			MockExpectation{
				path:     deletePath,
				verb:     http.MethodDelete,
				tenantID: tenantID,
			},
			MockResponse{
				statusCode: http.StatusInternalServerError,
			},
		)

		err := client.Delete(t.Context(), tenantID, itemID)
		require.Equal(t, "failed to delete resource, status code: 500", err.Error())
		m.AssertCalled(t)
	})

	t.Run("does not return error if the response is 200", func(t *testing.T) {
		m := runMocha(t, marketplaceBaseURL)
		m = registerAPI(t, m,
			MockExpectation{
				path:     deletePath,
				verb:     http.MethodDelete,
				tenantID: tenantID,
			},
			MockResponse{
				statusCode: http.StatusOK,
			},
		)

		err := client.Delete(t.Context(), tenantID, itemID)
		require.NoError(t, err)
		m.AssertCalled(t)
	})

	t.Run("does not return error if the response is 204", func(t *testing.T) {
		m := runMocha(t, marketplaceBaseURL)
		m = registerAPI(t, m,
			MockExpectation{
				path:     deletePath,
				verb:     http.MethodDelete,
				tenantID: tenantID,
			},
			MockResponse{
				statusCode: http.StatusNoContent,
			},
		)

		err := client.Delete(t.Context(), tenantID, itemID)
		require.NoError(t, err)
		m.AssertCalled(t)
	})
}
