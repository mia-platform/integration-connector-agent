// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package hmac

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
				HeaderName: "X-Header-Name",
				Secret:     "It's a Secret to Everybody",
			},
		},
		"missing hader name": {
			authentication: Authentication{
				Secret: "It's a Secret to Everybody",
			},
			expectedErr: webhook.ErrInvalidWebhookAuthenticationConfig,
		},
		"missing secret is valid": {
			authentication: Authentication{
				HeaderName: "X-Header-Name",
			},
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

func TestCheckHMACSignature(t *testing.T) {
	t.Parallel()

	webhookSignatureHeader := "X-Hub-Signature"
	tests := map[string]struct {
		request        fakeValidatingRequest
		authentication webhook.Authentication
		expectedErr    error
	}{
		"no header and no secret return no error": {
			authentication: &Authentication{},
		},
		"missing secret return error": {
			authentication: &Authentication{
				HeaderName: webhookSignatureHeader,
			},
			request: fakeValidatingRequest{
				headers: map[string][]string{
					webhookSignatureHeader: {"signature"},
				},
			},
			expectedErr: ErrSignatureHeaderButNoSecret,
		},
		"multiple header return error": {
			authentication: &Authentication{
				HeaderName: webhookSignatureHeader,
				Secret:     "secret",
			},
			request: fakeValidatingRequest{
				headers: map[string][]string{
					webhookSignatureHeader: {"signature", "other"},
				},
			},
			expectedErr: ErrMultipleSignatureHeaders,
		},
		"valid signature return nil": {
			authentication: &Authentication{
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
			authentication: &Authentication{
				HeaderName: webhookSignatureHeader,
				Secret:     "It's a Secret to Everybody",
			},
			request: fakeValidatingRequest{
				body: []byte("tampered body"),
				headers: map[string][]string{
					webhookSignatureHeader: {"sha256=a4771c39fbe90f317c7824e83ddef3caae9cb3d976c214ace1f2937e133263c9"},
				},
			},
			expectedErr: ErrInvalidSignature,
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			err := test.authentication.CheckSignature(test.request)
			assert.Equal(t, test.expectedErr, err)
		})
	}
}

type fakeValidatingRequest struct {
	headers map[string][]string
	body    []byte
}

func (r fakeValidatingRequest) GetReqHeaders() map[string][]string { return r.headers }
func (r fakeValidatingRequest) Body() []byte                       { return r.body }
