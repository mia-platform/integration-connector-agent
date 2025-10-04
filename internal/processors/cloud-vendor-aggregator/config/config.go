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

package config

import (
	"errors"

	"github.com/mia-platform/integration-connector-agent/internal/azure"
	"github.com/mia-platform/integration-connector-agent/internal/config"
)

var (
	ErrMissingCloudVendorName = errors.New("missing cloud vendor name")
	ErrInvalidCloudVendor     = errors.New("invalid cloud vendor name, must be one of: gcp, aws, azure")
)

type AuthOptions struct {
	// Azure
	azure.AuthConfig

	// GCP
	CredenialsJSON config.SecretSource `json:"credentialsJson,omitempty"` //nolint:tagliatelle

	// AWS
	AccessKeyID     string              `json:"accessKeyId"`
	SecretAccessKey config.SecretSource `json:"secretAccessKey"`
	SessionToken    config.SecretSource `json:"sessionToken"`
	Region          string              `json:"region"`
}

type Config struct {
	CloudVendorName string      `json:"cloudVendorName"`
	AuthOptions     AuthOptions `json:"authOptions"`
}

func (c Config) Validate() error {
	if c.CloudVendorName == "" {
		return ErrMissingCloudVendorName
	}

	if c.CloudVendorName != "gcp" &&
		c.CloudVendorName != "aws" &&
		c.CloudVendorName != "azure" {
		return ErrInvalidCloudVendor
	}

	return nil
}
