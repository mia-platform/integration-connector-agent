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

package basic

import (
	"testing"

	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook"
	"github.com/stretchr/testify/assert"
)

func TestValidation(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		authentication webhook.Authentication
		expectedErr    error
	}{
		"valid authentication": {
			authentication: Authentication{
				Username: "testuser",
				Secret:   "secret",
			},
		},
		"empty config is valid": {
			authentication: Authentication{},
		},
		"missing username is valid": {
			authentication: Authentication{
				Username: "",
				Secret:   "secret",
			},
		},
		"missing secret": {
			authentication: Authentication{
				Username: "testuser",
				Secret:   "",
			},
			expectedErr: webhook.ErrInvalidWebhookAuthenticationConfig,
		},
	}

	for testName, test := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			err := test.authentication.Validate()
			assert.ErrorIs(t, err, test.expectedErr)
		})
	}
}

func TestChekBasicAuth(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request        fakeValidatingRequest
		authentication webhook.Authentication
		expectedErr    error
	}{
		"valid basic auth return nil": {
			authentication: &Authentication{
				Username: "testuser",
				Secret:   "secret",
			},
			request: fakeValidatingRequest{
				headers: map[string][]string{
					"Authorization": {"Basic dGVzdHVzZXI6c2VjcmV0"},
				},
			},
		},
		"invalid auth return error": {
			authentication: &Authentication{
				Username: "",
				Secret:   "secret",
			},
			request: fakeValidatingRequest{
				headers: map[string][]string{
					"Authorization": {"Basic dGVzdHVzZXI6c2VjcmV0"},
				},
			},
			expectedErr: ErrUnauthorized,
		},
		"multiple auth headers return error": {
			authentication: &Authentication{
				Username: "testuser",
				Secret:   "secret",
			},
			request: fakeValidatingRequest{
				headers: map[string][]string{
					"Authorization": {"Basic dGVzdHVzZXI6c2VjcmV0", "Basic dGVzdHVzZXI6c2VjcmV0"},
				},
			},
			expectedErr: ErrMultipleAuthenticationHeadersFound,
		},
		"different authorization type return error": {
			authentication: &Authentication{
				Username: "testuser",
				Secret:   "secret",
			},
			request: fakeValidatingRequest{
				headers: map[string][]string{
					"Authorization": {"Bearer dGVzdHVzZXI6c2VjcmV0"},
				},
			},
			expectedErr: ErrInvalidAuthenticationType,
		},
		"no authorization header return error": {
			authentication: &Authentication{
				Username: "testuser",
				Secret:   "secret",
			},
			request: fakeValidatingRequest{
				headers: map[string][]string{},
			},
			expectedErr: ErrNoAuthenticationHeaderFound,
		},
	}

	for testName, test := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			err := test.authentication.CheckSignature(test.request)
			assert.ErrorIs(t, err, test.expectedErr)
		})
	}
}

type fakeValidatingRequest struct {
	headers map[string][]string
	body    []byte
}

func (r fakeValidatingRequest) GetReqHeaders() map[string][]string { return r.headers }
func (r fakeValidatingRequest) Body() []byte                       { return r.body }
