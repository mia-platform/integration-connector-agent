package config

import (
	"errors"

	"github.com/mia-platform/integration-connector-agent/internal/config"
)

var (
	ErrMissingCloudVendorName = errors.New("missing cloud vendor name")
	ErrInvalidCloudVendor     = errors.New("invalid cloud vendor name, must be one of: gcp, aws, azure")
)

type AuthOptions struct {
	CredenialsJson config.SecretSource `json:"credentialsJson,omitempty"`
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
