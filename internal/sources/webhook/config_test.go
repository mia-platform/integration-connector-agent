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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateConfiguration(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		config      Configuration[*fakeAuthentication]
		expectedErr error
	}{
		"valid config without authorization": {
			config: Configuration[*fakeAuthentication]{
				WebhookPath: "/path",
				Events:      &Events{},
			},
		},
		"valid config with authorization": {
			config: Configuration[*fakeAuthentication]{
				WebhookPath:    "/path",
				Events:         &Events{},
				Authentication: &fakeAuthentication{},
			},
		},
		"empty configuration": {
			expectedErr: ErrWebhookPathRequired,
		},
		"empty events": {
			config: Configuration[*fakeAuthentication]{
				WebhookPath: "/path",
			},
			expectedErr: ErrSupportedEventsRequired,
		},
		"invalid authorization": {
			config: Configuration[*fakeAuthentication]{
				WebhookPath: "/path",
				Events:      &Events{},
				Authentication: &fakeAuthentication{
					validationErr: ErrInvalidWebhookAuthenticationConfig,
				},
			},
			expectedErr: ErrInvalidWebhookAuthenticationConfig,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := tc.config.Validate()
			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, tc.expectedErr)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestCheckSignature(t *testing.T) {
	t.Parallel()

	invalidAuhtErr := fmt.Errorf("invalid authentication")
	testCases := map[string]struct {
		config      Configuration[*fakeAuthentication]
		req         ValidatingRequest
		expectedErr error
	}{
		"no authentication": {},
		"request is nil": {
			config: Configuration[*fakeAuthentication]{
				Authentication: &fakeAuthentication{},
			},
			expectedErr: ErrMissingRequest,
		},
		"valid authentication": {
			config: Configuration[*fakeAuthentication]{
				Authentication: &fakeAuthentication{},
			},
			req: &fakeRequest{},
		},
		"invalid authentication": {
			config: Configuration[*fakeAuthentication]{
				Authentication: &fakeAuthentication{
					checkErr: invalidAuhtErr,
				},
			},
			req:         &fakeRequest{},
			expectedErr: invalidAuhtErr,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := tc.config.CheckSignature(tc.req)
			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, tc.expectedErr)
				return
			}

			assert.NoError(t, err)
		})
	}
}

type fakeRequest struct{}

func (f fakeRequest) GetReqHeaders() map[string][]string { return map[string][]string{} }
func (f fakeRequest) Body() []byte                       { return nil }

type fakeAuthentication struct {
	checkErr      error
	validationErr error
}

func (f *fakeAuthentication) CheckSignature(req ValidatingRequest) error { return f.checkErr }
func (f *fakeAuthentication) Validate() error                            { return f.validationErr }
