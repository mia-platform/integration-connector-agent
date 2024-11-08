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

package jira

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateWebhookRequest(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		request     fakeValidatingRequest
		secret      string
		expectedErr error
	}{
		"no header and no secret return no error": {},
		"missing secret return error": {
			request: fakeValidatingRequest{
				headers: map[string][]string{
					webhookSignatureHeader: {"signature"},
				},
			},
			expectedErr: errors.New(signatureHeaderButNoSecretError),
		},
		"missing header return error": {
			secret:      "secret",
			expectedErr: errors.New(noSignatureHeaderButSecretError),
		},
		"multiple header return error": {
			secret: "secret",
			request: fakeValidatingRequest{
				headers: map[string][]string{
					webhookSignatureHeader: {"signature", "other"},
				},
			},
			expectedErr: errors.New(multipleSignatureHeadersError),
		},
		"valid signature return nil": {
			secret: "It's a Secret to Everybody",
			request: fakeValidatingRequest{
				body: []byte("Hello World!"),
				headers: map[string][]string{
					webhookSignatureHeader: {"sha256=a4771c39fbe90f317c7824e83ddef3caae9cb3d976c214ace1f2937e133263c9"},
				},
			},
		},
		"invalid signature return error": {
			secret: "It's a Secret to Everybody",
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
			err := ValidateWebhookRequest(test.request, test.secret)
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
