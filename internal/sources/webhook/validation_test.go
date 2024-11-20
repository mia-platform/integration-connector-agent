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
	"errors"
	"testing"

	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestValidateWebhookRequest(t *testing.T) {
	t.Parallel()

	webhookSignatureHeader := "X-Hub-Signature"

	tests := map[string]struct {
		request        fakeValidatingRequest
		authentication Authentication
		expectedErr    error
	}{
		"no header and no secret return no error": {},
		"missing secret return error": {
			authentication: Authentication{
				HeaderName: webhookSignatureHeader,
			},
			request: fakeValidatingRequest{
				headers: map[string][]string{
					webhookSignatureHeader: {"signature"},
				},
			},
			expectedErr: errors.New(signatureHeaderButNoSecretError),
		},
		"missing header return error": {
			authentication: Authentication{
				HeaderName: webhookSignatureHeader,
				Secret:     config.SecretSource("secret"),
			},
			expectedErr: errors.New(noSignatureHeaderButSecretError),
		},
		"multiple header return error": {
			authentication: Authentication{
				HeaderName: webhookSignatureHeader,
				Secret:     config.SecretSource("secret"),
			},
			request: fakeValidatingRequest{
				headers: map[string][]string{
					webhookSignatureHeader: {"signature", "other"},
				},
			},
			expectedErr: errors.New(multipleSignatureHeadersError),
		},
		"valid signature return nil": {
			authentication: Authentication{
				HeaderName: webhookSignatureHeader,
				Secret:     config.SecretSource("It's a Secret to Everybody"),
			},
			request: fakeValidatingRequest{
				body: []byte("Hello World!"),
				headers: map[string][]string{
					webhookSignatureHeader: {"sha256=a4771c39fbe90f317c7824e83ddef3caae9cb3d976c214ace1f2937e133263c9"},
				},
			},
		},
		"invalid signature return error": {
			authentication: Authentication{
				HeaderName: webhookSignatureHeader,
				Secret:     config.SecretSource("It's a Secret to Everybody"),
			},
			request: fakeValidatingRequest{
				body: []byte("tampered body"),
				headers: map[string][]string{
					webhookSignatureHeader: {"sha256=a4771c39fbe90f317c7824e83ddef3caae9cb3d976c214ace1f2937e133263c9"},
				},
			},
			expectedErr: errors.New(invalidSignatureError),
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			err := ValidateWebhookRequest(test.request, test.authentication)
			assert.Equal(t, test.expectedErr, err)
		})
	}
}

var _ ValidatingRequest = fakeValidatingRequest{}

type fakeValidatingRequest struct {
	headers map[string][]string
	body    []byte
}

func (r fakeValidatingRequest) GetReqHeaders() map[string][]string {
	return r.headers
}

func (r fakeValidatingRequest) Body() []byte {
	return r.body
}
