// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

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
