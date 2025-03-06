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

package webhook

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateConfiguration(t *testing.T) {
	testCases := map[string]struct {
		config Configuration

		expectedErr string
	}{
		"empty configuration": {
			expectedErr: "webhook path is required",
		},
		"empty events": {
			config: Configuration{
				WebhookPath: "/path",
			},
			expectedErr: "events are empty",
		},
		"ok": {
			config: Configuration{
				WebhookPath: "/path",
				Events:      &Events{},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := tc.config.Validate()
			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

type fakeRequest struct {
	body    []byte
	headers map[string][]string
}

func (f fakeRequest) GetReqHeaders() map[string][]string {
	return f.headers
}
func (f fakeRequest) Body() []byte {
	return f.body
}

func TestCheckSignature(t *testing.T) {
	testCases := map[string]struct {
		config Configuration
		req    ValidatingRequest

		expectedErr string
	}{
		"no authentication": {},
		"request is nil": {
			config: Configuration{
				Authentication: &HMAC{},
			},
			expectedErr: "invalid webhook authentication configuration: request is nil",
		},
		"ok": {
			config: Configuration{},
			req: &fakeRequest{
				body:    []byte("body"),
				headers: map[string][]string{"header": {"sha256=signature"}},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := tc.config.CheckSignature(tc.req)
			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
