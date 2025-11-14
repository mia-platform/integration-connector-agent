// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package token

import (
	"errors"
	"fmt"

	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook"
)

var (
	ErrTokenHeaderButNoToken = errors.New("token not configured for validating webhook")
	ErrTokenButNoHeader      = errors.New("token configured but no header found in request")
	ErrMultipleTokenHeaders  = errors.New("multiple token headers found")
	ErrInvalidToken          = errors.New("invalid token in request")
)

var _ webhook.Authentication = Authentication{}

type Authentication struct {
	HeaderName string              `json:"headerName"`
	Token      config.SecretSource `json:"token"`
}

// CheckSignature will read the webhook token header and the given secret.
// It will fail if there is a mismatch between the two values.
func (a Authentication) CheckSignature(req webhook.ValidatingRequest) error {
	tokenValue, err := tokenSignatureHeaderValue(req, a.HeaderName, a.Token.String())
	if err != nil {
		return err
	}

	if a.Token.String() != tokenValue {
		return ErrInvalidToken
	}

	return nil
}

// Validate checks if the Token configuration is valid. It requires that both token and headerName are set
func (a Authentication) Validate() error {
	if a.HeaderName == "" && a.Token.String() != "" {
		return fmt.Errorf("%w: headerName not present but token is set", webhook.ErrInvalidWebhookAuthenticationConfig)
	}

	return nil
}

func tokenSignatureHeaderValue(req webhook.ValidatingRequest, headerName, token string) (string, error) {
	headerValues := req.GetReqHeaders()[headerName]
	switch {
	case len(headerValues) == 0 && len(token) == 0:
		return "", nil
	case len(headerValues) != 0 && len(token) == 0:
		return "", ErrTokenHeaderButNoToken
	case len(headerValues) == 0 && len(token) != 0:
		return "", ErrTokenButNoHeader
	case len(headerValues) > 1:
		return "", ErrMultipleTokenHeaders
	}

	return headerValues[0], nil
}
