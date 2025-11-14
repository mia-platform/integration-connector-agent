// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package token

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
				Token:      "token",
			},
		},
		"missing header name": {
			authentication: Authentication{
				Token: "token",
			},
			expectedErr: webhook.ErrInvalidWebhookAuthenticationConfig,
		},
		"missing token is valid": {
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

func TestCheckToken(t *testing.T) {
	t.Parallel()

	webhookTokenHeader := "X-Webhook-Token"
	tests := map[string]struct {
		request        fakeValidatingRequest
		authentication webhook.Authentication
		expectedErr    error
	}{
		"no header and no token return no error": {
			authentication: &Authentication{},
		},
		"missing token return error": {
			authentication: &Authentication{
				HeaderName: webhookTokenHeader,
				Token:      "token",
			},
			request: fakeValidatingRequest{
				body: []byte("Hello World!"),
			},
			expectedErr: ErrTokenButNoHeader,
		},
		"multiple header return error": {
			authentication: &Authentication{
				HeaderName: webhookTokenHeader,
				Token:      "token",
			},
			request: fakeValidatingRequest{
				headers: map[string][]string{
					webhookTokenHeader: {"signature", "other"},
				},
			},
			expectedErr: ErrMultipleTokenHeaders,
		},
		"valid token return nil": {
			authentication: &Authentication{
				HeaderName: webhookTokenHeader,
				Token:      "token",
			},
			request: fakeValidatingRequest{
				body: []byte("Hello World!"),
				headers: map[string][]string{
					webhookTokenHeader: {"token"},
				},
			},
		},
		"invalid signature return error": {
			authentication: &Authentication{
				HeaderName: webhookTokenHeader,
				Token:      "Token",
			},
			request: fakeValidatingRequest{
				body: []byte("Hello World!"),
				headers: map[string][]string{
					webhookTokenHeader: {"anotherToken"},
				},
			},
			expectedErr: ErrInvalidToken,
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
