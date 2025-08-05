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

package hmac

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook"
)

var (
	ErrSignatureHeaderButNoSecret = errors.New("secret not configured for validating webhook signature")
	ErrMultipleSignatureHeaders   = errors.New("multiple signature headers found")
	ErrInvalidSignature           = errors.New("invalid signature in request")
)

var _ webhook.Authentication = Authentication{}

type Authentication struct {
	Secret     config.SecretSource `json:"secret"`
	HeaderName string              `json:"headerName"`
}

// CheckSignature will read the webhook signature header and the given secret for validating the webhook
// payload. It will fail if there is a mismatch in the signatures
func (a Authentication) CheckSignature(req webhook.ValidatingRequest) error {
	secret := a.Secret.String()
	signatureValue, err := hmacSignatureHeaderValue(req, a.HeaderName, secret)
	if err != nil || signatureValue == "" {
		return err
	}

	signature, _ := strings.CutPrefix(signatureValue, "sha256=")
	if !validateBody(req.Body(), secret, signature) {
		return ErrInvalidSignature
	}

	return nil
}

// Validate checks if the HMAC configuration is valid. It requires that both secret and headerName are set
func (a Authentication) Validate() error {
	switch {
	case a.HeaderName == "":
		if a.Secret.String() != "" {
			return fmt.Errorf("%w: headerName not present but secret is set", webhook.ErrInvalidWebhookAuthenticationConfig)
		}
	case a.Secret.String() == "":
		if a.HeaderName != "" {
			return fmt.Errorf("%w: secret not present but headerName is set", webhook.ErrInvalidWebhookAuthenticationConfig)
		}
	}

	return nil
}

// validateBody will generate an hmac encoding of bodyData using secret, and than compare it with the expectedSignature
func validateBody(bodyData []byte, secret, expectedSignature string) bool {
	hasher := hmac.New(sha256.New, []byte(secret))
	hasher.Write(bodyData)
	generatedMAC := hasher.Sum(nil)

	expectedMac, err := hex.DecodeString(expectedSignature)
	if err != nil {
		return false
	}

	return hmac.Equal(generatedMAC, expectedMac)
}

func hmacSignatureHeaderValue(req webhook.ValidatingRequest, headerName, secret string) (string, error) {
	headerValues := req.GetReqHeaders()[headerName]
	switch {
	case len(headerValues) == 0 && len(secret) == 0:
		return "", nil
	case len(headerValues) != 0 && len(secret) == 0:
		return "", ErrSignatureHeaderButNoSecret
	case len(headerValues) > 1:
		return "", ErrMultipleSignatureHeaders
	}

	return headerValues[0], nil
}
