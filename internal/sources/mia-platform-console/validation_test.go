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

package console

import (
	"fmt"
	"testing"

	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook"
	"github.com/stretchr/testify/require"
)

func TestCheckConsoleSignature(t *testing.T) {
	t.Parallel()

	signatureHeader := "X-Hub-Signature"

	tests := map[string]struct {
		request        *fakeValidatingRequest
		authentication ValidationConfig
		expectedErr    string
	}{
		"no request return error": {
			authentication: ValidationConfig{},
			expectedErr:    "request is nil",
		},
		"no header and no secret return no error": {
			authentication: ValidationConfig{},
			request:        &fakeValidatingRequest{},
		},
		"missing secret return error": {
			authentication: ValidationConfig{
				HeaderName: signatureHeader,
			},
			request: &fakeValidatingRequest{
				headers: map[string][]string{
					signatureHeader: {"signature"},
				},
			},
			expectedErr: webhook.SignatureHeaderButNoSecretError,
		},
		"missing configured header but secret present returns error": {
			authentication: ValidationConfig{
				Secret: config.SecretSource("secret"),
			},
			request:     &fakeValidatingRequest{},
			expectedErr: fmt.Sprintf("%s: secret is set but headerName not present", webhook.InvalidWebhookAuthenticationConfig),
		},
		"missing header return error": {
			authentication: ValidationConfig{
				HeaderName: signatureHeader,
				Secret:     "secret",
			},
			request:     &fakeValidatingRequest{},
			expectedErr: webhook.NoSignatureHeaderButSecretError,
		},
		"multiple header return error": {
			authentication: ValidationConfig{
				HeaderName: signatureHeader,
				Secret:     "secret",
			},
			request: &fakeValidatingRequest{
				headers: map[string][]string{
					signatureHeader: {"signature", "other"},
				},
			},
			expectedErr: webhook.MultipleSignatureHeadersError,
		},
		"valid signature return nil": {
			authentication: ValidationConfig{
				HeaderName: signatureHeader,
				Secret:     "It's a Secret to Everybody",
			},
			request: &fakeValidatingRequest{
				body: []byte("Hello World!"),
				headers: map[string][]string{
					signatureHeader: {"sha256=b738052486cd876b13b5404a45479bd8caca05e5267c10d8436f1570547e3056"},
				},
			},
		},
		"invalid signature return error": {
			authentication: ValidationConfig{
				HeaderName: signatureHeader,
				Secret:     "It's a Secret to Everybody",
			},
			request: &fakeValidatingRequest{
				body: []byte("tampered body"),
				headers: map[string][]string{
					signatureHeader: {"sha256=b738052486cd876b13b5404a45479bd8caca05e5267c10d8436f1570547e3056"},
				},
			},
			expectedErr: webhook.InvalidSignatureError,
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			err := test.authentication.CheckSignature(test.request)
			if test.expectedErr != "" {
				require.EqualError(t, err, test.expectedErr)
				return
			}
			require.NoError(t, err)
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
