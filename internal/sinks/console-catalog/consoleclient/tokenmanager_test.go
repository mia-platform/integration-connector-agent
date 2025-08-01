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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestTokenManager(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		handler       http.Handler
		cachedToken   *oauth2.Token
		expectedToken string
		expectedError string
	}{
		"request a new token": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, m2mTokenPath, r.RequestURI)
				require.Equal(t, http.MethodPost, r.Method)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(`{"access_token": "new-token", "token_type": "Bearer", "expires_in": 3600}`))
				require.NoError(t, err)
			}),
			expectedToken: "Bearer new-token",
		},
		"use cached token": {
			handler: failingHandler(t),
			cachedToken: &oauth2.Token{
				AccessToken: "cached-token",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
				Expiry:      time.Now().Add(time.Hour),
			},
			expectedToken: "Bearer cached-token",
		},
		"cached token expired, request a new one": {
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, m2mTokenPath, r.RequestURI)
				require.Equal(t, http.MethodPost, r.Method)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(`{"access_token": "new-token", "token_type": "Bearer", "expires_in": 3600}`))
				require.NoError(t, err)
			}),
			cachedToken: &oauth2.Token{
				AccessToken: "expired-token",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
				Expiry:      time.Now(), // Simulate an expired token
			},
			expectedToken: "Bearer new-token",
		},
		"server error while requesting token": {
			handler:       failingHandler(t),
			expectedError: "oauth2: cannot fetch token",
		},
	}

	for testName, test := range testCases {
		t.Run(testName, func(t *testing.T) {
			server := testServer(t, test.handler)
			defer server.Close()

			tm, err := NewClientCredentialsTokenManager(server.URL, "test-client-id", "test-client-secret")
			require.NoError(t, err)
			tm.cachedTkn = test.cachedToken

			request := httptest.NewRequest(http.MethodGet, server.URL, nil)
			err = tm.SetAuthHeader(request)
			if len(test.expectedError) > 0 {
				assert.ErrorContains(t, err, test.expectedError)
				return
			}

			assert.NoError(t, err)
			require.NotEmpty(t, request.Header.Get("Authorization"))
			assert.Equal(t, test.expectedToken, request.Header.Get("Authorization"))
		})
	}
}

func failingHandler(t *testing.T) http.Handler {
	t.Helper()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	})
}

func testServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	return httptest.NewServer(handler)
}
