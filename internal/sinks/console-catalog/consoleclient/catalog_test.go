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

package consoleclient

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vitorsalgado/mocha/v3"
	"github.com/vitorsalgado/mocha/v3/expect"
	"github.com/vitorsalgado/mocha/v3/reply"
)

type testResource map[string]any

func TestMarketplaceApply(t *testing.T) {
	const marketplaceBaseURL = "127.0.0.1:45874"
	const tenantID = "tenant123"

	client := New[testResource](fmt.Sprintf("http://%s/", marketplaceBaseURL), &mockedTokenManager{})
	item := MarketplaceResource[testResource]{
		ItemID:   "myItem",
		TenantID: tenantID,
		Name:     "myItemName",
		Type:     "resType",
		Resources: testResource{
			"reskey": "resval",
		},
	}

	t.Run("returns error if the execution of request fails", func(t *testing.T) {
		itemID, err := client.Apply(t.Context(), &item)
		require.Equal(t, "", itemID)
		require.True(t, errors.Is(err, ErrMarketplaceRequestExecution))
	})

	t.Run("returns error if the response is not 200", func(t *testing.T) {
		expectedRequest := MockExpectation{
			tenantID: tenantID,
			headers: map[string]string{
				"content-type": "application/json",
			},
		}

		mockedResponse := MockResponse{
			statusCode: http.StatusInternalServerError,
		}

		m := runMocha(t, marketplaceBaseURL)
		m = registerPostItemMock(t, m, expectedRequest, mockedResponse)

		itemID, err := client.Apply(t.Context(), &item)
		require.Equal(t, "", itemID)
		require.Equal(t, "failed to apply resource, status code: 500", err.Error())
		m.AssertCalled(t)
	})

	t.Run("returns error if the response body is unknown", func(t *testing.T) {
		expectedRequest := MockExpectation{tenantID: tenantID}

		mockedResponse := MockResponse{
			statusCode: http.StatusOK,
			body:       []byte("unknown body"),
		}

		m := runMocha(t, marketplaceBaseURL)
		m = registerPostItemMock(t, m, expectedRequest, mockedResponse)

		itemID, err := client.Apply(t.Context(), &item)
		require.Equal(t, "", itemID)
		require.True(t, errors.Is(err, ErrMarketplaceResponseParse))
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
					ValidationErrors: []ValidationError{
						mockedValidationError1,
						mockedValidationError2,
					},
				},
			},
		}

		expectedRequest := MockExpectation{tenantID: tenantID}

		mochaMockedResponse := MockResponse{
			statusCode: http.StatusOK,
			body:       mockedResponse,
		}

		m := runMocha(t, marketplaceBaseURL)
		m = registerPostItemMock(t, m, expectedRequest, mochaMockedResponse)

		itemID, err := client.Apply(t.Context(), &item)
		require.Equal(t, "", itemID)
		var parsedErr *MarketplaceValidationError
		require.True(t, errors.As(err, &parsedErr))
		require.Equal(t, 2, len(parsedErr.Errors))
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

		expectedRequest := MockExpectation{tenantID: tenantID}

		mochaMockedResponse := MockResponse{
			statusCode: http.StatusOK,
			body:       mockedResponse,
		}

		m := runMocha(t, marketplaceBaseURL)
		m = registerPostItemMock(t, m, expectedRequest, mochaMockedResponse)

		itemID, err := client.Apply(t.Context(), &item)
		require.Nil(t, err)
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
			Type:     "resType",
			Resources: testResource{
				"k1": "v1",
			},
		}

		expectedMarketplaceRequestBodyString := fmt.Sprintf("{\"resources\":[{\"description\":\"\",\"itemId\":\"myItem\",\"name\":\"myItemName\",\"resources\":{\"k1\":\"v1\"},\"tenantId\":\"%s\",\"type\":\"resType\"}]}", tenantID)
		expectedRequest := MockExpectation{
			tenantID:   tenantID,
			bodyString: expectedMarketplaceRequestBodyString,
		}

		mochaMockedResponse := MockResponse{
			statusCode: http.StatusOK,
			body:       mockedResponse,
		}

		m := runMocha(t, marketplaceBaseURL)
		m = registerPostItemMock(t, m, expectedRequest, mochaMockedResponse)

		itemID, err := client.Apply(t.Context(), &item)
		require.Nil(t, err)
		require.Equal(t, mockedItemID, itemID)
		m.AssertCalled(t)
	})

	t.Run("correctly sets authorization header", func(t *testing.T) {
		url := fmt.Sprintf("http://%s/", marketplaceBaseURL)
		tknmngr := NewClientCredentialsTokenManager(url, "myClientId", "myClientSecret")
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
			Type:     "resType",
			Resources: testResource{
				"k1": "v1",
			},
		}
		expectedRequest := MockExpectation{
			tenantID: tenantID,
			headers: map[string]string{
				"Authorization": "Bearer the-new-token",
			},
		}

		mochaMockedResponse := MockResponse{
			statusCode: http.StatusOK,
			body:       mockedResponse,
		}

		m := runMocha(t, marketplaceBaseURL)
		m = registerPostItemMock(t, m, expectedRequest, mochaMockedResponse)
		m = registerOauthTokenMock(t, m, MockExpectation{
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte("myClientId:myClientSecret"))),
				"Content-Type":  "application/x-www-form-urlencoded",
			},
		}, MockResponse{
			statusCode: http.StatusOK,
			body: map[string]any{
				"access_token": "the-new-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			},
		})

		itemID, err := client.Apply(t.Context(), &item)
		require.Nil(t, err)
		require.Equal(t, mockedItemID, itemID)
		m.AssertCalled(t)
	})

	t.Run("correctly reuses non expired authorization header", func(t *testing.T) {
		url := fmt.Sprintf("http://%s/", marketplaceBaseURL)
		tknmngr := NewClientCredentialsTokenManager(url, "myClientId", "myClientSecret")

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
			Type:     "resType",
			Resources: testResource{
				"k1": "v1",
			},
		}
		expectedMarketplaceRequestBodyString1 := fmt.Sprintf("{\"resources\":[{\"description\":\"\",\"itemId\":\"myItem\",\"name\":\"myItemName\",\"resources\":{\"k1\":\"v1\"},\"tenantId\":\"%s\",\"type\":\"resType\"}]}", tenantID)

		m := runMocha(t, marketplaceBaseURL)
		m = registerOauthTokenMock(t, m,
			MockExpectation{
				headers: map[string]string{
					"Authorization": fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte("myClientId:myClientSecret"))),
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

		m = registerPostItemMock(t, m,
			MockExpectation{
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
		require.Nil(t, err)
		require.Equal(t, mockedItemID, itemID)
		m.AssertCalled(t)

		itemID, err = client.Apply(t.Context(), &item1)
		require.Nil(t, err)
		require.Equal(t, mockedItemID, itemID)
		m.AssertCalled(t)
	})
}

type MockExpectation struct {
	tenantID   string
	headers    map[string]string
	bodyString string
}

type MockResponse struct {
	statusCode int
	body       any
	times      int
}

func runMocha(t *testing.T, mockAddr string) *mocha.Mocha {
	t.Helper()

	options := mocha.Configure().Addr(mockAddr)
	if testing.Verbose() {
		options = options.LogVerbosity(mocha.LogVerbose)
	}

	m := mocha.New(t, options.Build())
	m.CloseOnCleanup(t)
	m.Start()

	return m
}

func registerOauthTokenMock(t *testing.T, m *mocha.Mocha, request MockExpectation, response MockResponse) *mocha.Mocha {
	t.Helper()

	responseStatus := response.statusCode
	if responseStatus == 0 {
		responseStatus = http.StatusOK
	}

	mock := mocha.Post(expect.URLPath("/api/m2m/oauth/token")).Repeat(response.times)

	if request.headers != nil {
		for key, value := range request.headers {
			mock = mock.Header(key, expect.ToEqual(value))
		}
	}

	if request.bodyString != "" {
		mock = mock.Body(expect.ToEqual(request.bodyString))
	}

	mock = mock.Reply(reply.Status(responseStatus).Header("content-type", "application/json").BodyJSON(response.body))

	m.AddMocks(mock)

	return m
}

func registerPostItemMock(t *testing.T, m *mocha.Mocha, request MockExpectation, responses ...MockResponse) *mocha.Mocha {
	t.Helper()

	mock := mocha.Post(expect.URLPath(fmt.Sprintf("/api/marketplace/tenants/%s/resources", request.tenantID)))

	if request.headers != nil {
		for key, value := range request.headers {
			mock = mock.Header(key, expect.ToEqual(value))
		}
	}

	replySequence := reply.Seq()
	for _, response := range responses {
		responseStatus := response.statusCode
		if responseStatus == 0 {
			responseStatus = http.StatusOK
		}

		if request.bodyString != "" {
			mock = mock.Body(expect.Func(func(v any, _ expect.Args) (bool, error) {
				bodyRaw, err := json.Marshal(v)
				if err != nil {
					return false, fmt.Errorf("unexpected error to read request body on mocha")
				}
				require.Equal(t, request.bodyString, string(bodyRaw))
				return true, nil
			}))
		}

		replySequence.Add(reply.Status(responseStatus).Header("content-type", "application/json").BodyJSON(response.body))
	}
	mock = mock.Reply(replySequence)

	m.AddMocks(mock)

	return m
}

type mockedTokenManager struct{}

func (t *mockedTokenManager) SetAuthHeader(req *http.Request) error {
	return nil
}
