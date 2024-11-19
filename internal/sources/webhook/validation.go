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
)

const (
	invalidWebhookAuthenticationConfig = "invalid webhook authentication configuration"
	signatureHeaderButNoSecretError    = "secret not configured for validating webhook signature"
	noSignatureHeaderButSecretError    = "missing webhook signature"
	multipleSignatureHeadersError      = "multiple signature headers found"
	invalidSignatureError              = "invalid signature in request"
)

type ValidatingRequest interface {
	GetReqHeaders() map[string][]string
	Body() []byte
}

// ValidateWebhookRequest will read the webhook signature header and the given secret for validating the webhook
// payload. It will fail if there is a mismatch in the signatures and if a signature or a secret is provided and the
// other is not present.
func ValidateWebhookRequest(req ValidatingRequest, authentication Authentication) error {
	secret := authentication.Secret.String()
	if secret != "" && authentication.HeaderName == "" {
		return fmt.Errorf("%s: secret is set but headerName not present", invalidWebhookAuthenticationConfig)
	}

	headerValues := req.GetReqHeaders()[authentication.HeaderName]
	switch {
	case len(headerValues) == 0 && len(secret) == 0:
		return nil
	case len(headerValues) == 0 && len(secret) != 0:
		return errors.New(noSignatureHeaderButSecretError)
	case len(headerValues) != 0 && len(secret) == 0:
		return errors.New(signatureHeaderButNoSecretError)
	case len(headerValues) > 1:
		return errors.New(multipleSignatureHeadersError)
	}

	signature, _ := strings.CutPrefix(headerValues[0], "sha256=")
	if !validateBody(req.Body(), secret, signature) {
		return errors.New(invalidSignatureError)
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
