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

package basic

import (
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook"
)

var (
	ErrNoAuthenticationHeaderFound        = errors.New("no authentication header found in request")
	ErrMultipleAuthenticationHeadersFound = errors.New("multiple authentication headers found in request")
	ErrInvalidAuthenticationType          = errors.New("invalid authentication type in request")
	ErrUnauthorized                       = errors.New("unauthorized request")
)

var _ webhook.Authentication = Authentication{}

type Authentication struct {
	Username string              `json:"username"`
	Secret   config.SecretSource `json:"secret"`
}

// CheckSignature will read the webhook authentication header and will check if the
// provided username and secret match the ones in the request.
func (a Authentication) CheckSignature(req webhook.ValidatingRequest) error {
	authHeader, found := req.GetReqHeaders()["Authorization"]
	if !found {
		return fmt.Errorf("%w: %w", ErrNoAuthenticationHeaderFound, ErrUnauthorized)
	}
	if len(authHeader) > 1 {
		return fmt.Errorf("%w: %w", ErrMultipleAuthenticationHeadersFound, ErrUnauthorized)
	}
	parts := strings.Fields(authHeader[0])
	if len(parts) != 2 || strings.ToLower(parts[0]) != "basic" {
		return fmt.Errorf("%w: %w", ErrInvalidAuthenticationType, ErrUnauthorized)
	}

	buffer := new(bytes.Buffer)
	fmt.Fprintf(buffer, "%s:%s", a.Username, a.Secret.String())
	expectedAuthentication := base64.StdEncoding.AppendEncode([]byte{}, buffer.Bytes())
	expectedAuthenticationHash := sha256.Sum256(expectedAuthentication)
	authenticationHash := sha256.Sum256([]byte(parts[1]))
	if subtle.ConstantTimeCompare(expectedAuthenticationHash[:], authenticationHash[:]) == 1 {
		return nil
	}

	return ErrUnauthorized
}

// Validate checks if the Basic configuration is valid. It requires that if username is set also secret must be set.
func (a Authentication) Validate() error {
	if a.Username != "" && a.Secret.String() == "" {
		return fmt.Errorf("%w: secret not present but username is set", webhook.ErrInvalidWebhookAuthenticationConfig)
	}

	return nil
}
