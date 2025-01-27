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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/mia-platform/integration-connector-agent/internal/config"
)

const (
	InvalidWebhookAuthenticationConfig = "invalid webhook authentication configuration"
	SignatureHeaderButNoSecretError    = "secret not configured for validating webhook signature"
	NoSignatureHeaderButSecretError    = "missing webhook signature"
	MultipleSignatureHeadersError      = "multiple signature headers found"
	InvalidSignatureError              = "invalid signature in request"
)

type HMAC struct {
	Secret     config.SecretSource `json:"secret"`
	HeaderName string              `json:"headerName"`
}

// CheckSignature will read the webhook signature header and the given secret for validating the webhook
// payload. It will fail if there is a mismatch in the signatures and if a signature or a secret is provided and the
// other is not present.
func (h HMAC) CheckSignature(req ValidatingRequest) error {
	if req == nil {
		return fmt.Errorf("%s: request is nil", InvalidWebhookAuthenticationConfig)
	}
	secret := h.Secret.String()
	if secret != "" && h.HeaderName == "" {
		return fmt.Errorf("%s: secret is set but headerName not present", InvalidWebhookAuthenticationConfig)
	}

	headerValues, err := GetHeaderValues(req, h.HeaderName, secret)
	if err != nil {
		return err
	}
	if headerValues == nil {
		return nil
	}

	signature, _ := strings.CutPrefix(headerValues[0], "sha256=")
	if !validateBody(req.Body(), secret, signature) {
		return errors.New(InvalidSignatureError)
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

func GetHeaderValues(req ValidatingRequest, headerName, secret string) ([]string, error) {
	headerValues := req.GetReqHeaders()[headerName]
	switch {
	case len(headerValues) == 0 && len(secret) == 0:
		return nil, nil
	case len(headerValues) == 0 && len(secret) != 0:
		return nil, errors.New(NoSignatureHeaderButSecretError)
	case len(headerValues) != 0 && len(secret) == 0:
		return nil, errors.New(SignatureHeaderButNoSecretError)
	case len(headerValues) > 1:
		return nil, errors.New(MultipleSignatureHeadersError)
	}

	return headerValues, nil
}
