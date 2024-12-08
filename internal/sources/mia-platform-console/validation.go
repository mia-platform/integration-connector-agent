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
	"crypto/sha256"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook"
)

type ValidationConfig struct {
	Secret     config.SecretSource `json:"secret"`
	HeaderName string              `json:"-"`
}

// CheckSignature will read the webhook signature header and the given secret for validating the webhook
// payload.
func (h ValidationConfig) CheckSignature(req webhook.ValidatingRequest) error {
	if req == nil || reflect.ValueOf(req).IsNil() {
		return fmt.Errorf("request is nil")
	}
	secret := h.Secret.String()
	if secret != "" && h.HeaderName == "" {
		return fmt.Errorf("%s: secret is set but headerName not present", webhook.InvalidWebhookAuthenticationConfig)
	}

	headerValues, err := webhook.GetHeaderValues(req, h.HeaderName, secret)
	if err != nil {
		return err
	}
	if headerValues == nil {
		return nil
	}

	signature, _ := strings.CutPrefix(headerValues[0], "sha256=")
	if !validateBody(req.Body(), secret, signature) {
		return errors.New(webhook.InvalidSignatureError)
	}

	return nil
}

// validateBody will generate an hmac encoding of bodyData using secret, and than compare it with the expectedSignature
func validateBody(bodyData []byte, secret, expectedSignature string) bool {
	hasher := sha256.New()
	hasher.Write(bodyData)
	hasher.Write([]byte(secret))
	generatedHash := hasher.Sum(nil)
	return fmt.Sprintf("%x", generatedHash) == expectedSignature
}
