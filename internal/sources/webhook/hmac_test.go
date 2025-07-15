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

	"github.com/stretchr/testify/assert"
)

func TestCheckHMACSignature(t *testing.T) {
	t.Parallel()

	webhookSignatureHeader := "X-Hub-Signature"

	tests := map[string]struct {
		request        fakeValidatingRequest
		authentication Authentication
		expectedErr    error
	}{
		"no header and no secret return no error": {
			authentication: &HMAC{},
		},
		"missing secret return error": {
			authentication: &HMAC{
				HeaderName: webhookSignatureHeader,
			},
			request: fakeValidatingRequest{
				headers: map[string][]string{
					webhookSignatureHeader: {"signature"},
				},
			},
			expectedErr: errors.New(SignatureHeaderButNoSecretError),
		},
		"missing header return error": {
			authentication: &HMAC{
				HeaderName: webhookSignatureHeader,
				Secret:     "secret",
			},
			expectedErr: errors.New(NoSignatureHeaderButSecretError),
		},
		"multiple header return error": {
			authentication: &HMAC{
				HeaderName: webhookSignatureHeader,
				Secret:     "secret",
			},
			request: fakeValidatingRequest{
				headers: map[string][]string{
					webhookSignatureHeader: {"signature", "other"},
				},
			},
			expectedErr: errors.New(MultipleSignatureHeadersError),
		},
		"valid signature return nil": {
			authentication: &HMAC{
				HeaderName: webhookSignatureHeader,
				Secret:     "It's a Secret to Everybody",
			},
			request: fakeValidatingRequest{
				body: []byte("Hello World!"),
				headers: map[string][]string{
					webhookSignatureHeader: {"sha256=a4771c39fbe90f317c7824e83ddef3caae9cb3d976c214ace1f2937e133263c9"},
				},
			},
		},
		"invalid signature return error": {
			authentication: &HMAC{
				HeaderName: webhookSignatureHeader,
				Secret:     "It's a Secret to Everybody",
			},
			request: fakeValidatingRequest{
				body: []byte("tampered body"),
				headers: map[string][]string{
					webhookSignatureHeader: {"sha256=a4771c39fbe90f317c7824e83ddef3caae9cb3d976c214ace1f2937e133263c9"},
				},
			},
			expectedErr: errors.New(InvalidSignatureError),
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			err := test.authentication.CheckSignature(test.request)
			assert.Equal(t, test.expectedErr, err)
		})
	}
}

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
