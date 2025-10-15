// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
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

package consolecatalog

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/mia-platform/integration-connector-agent/config"
	"github.com/mia-platform/integration-connector-agent/internal/sinks/console-catalog/consoleclient"
)

var (
	ErrURLNotSet              = errors.New("URL not set in Console Catalog sink configuration")
	ErrInvalidURL             = errors.New("invalid URL in Console Catalog sink configuration")
	ErrInvalidLifecycleStatus = errors.New("invalid itemLifecycleStatus in Console Catalog sink configuration")
	ErrMissingField           = errors.New("missing required field in Console Catalog sink configuration")
)

type Config struct {
	URL                   string                              `json:"url"`
	TenantID              string                              `json:"tenantId"`
	ClientID              string                              `json:"clientId"`
	ClientSecret          config.SecretSource                 `json:"clientSecret"`
	ItemTypeDefinitionRef consoleclient.ItemTypeDefinitionRef `json:"itemTypeDefinitionRef"`
	ItemLifecycleStatus   consoleclient.LifecycleStatus       `json:"itemLifecycleStatus"`
	ItemIDTemplate        string                              `json:"itemIdTemplate"`
	ItemNameTemplate      string                              `json:"itemNameTemplate"`

	// New sync model fields - moved from catalogMetadata to direct fields
	CategoryIDTemplate          string                `json:"categoryIdTemplate,omitempty"`
	DescriptionTemplate         string                `json:"descriptionTemplate,omitempty"`
	ImageURLTemplate            string                `json:"imageUrlTemplate,omitempty"`
	Documentation               *DocumentationMapping `json:"documentation,omitempty"`
	Labels                      map[string]string     `json:"labels,omitempty"`
	Annotations                 map[string]string     `json:"annotations,omitempty"`
	Tags                        []string              `json:"tags,omitempty"`
	Links                       []LinkMapping         `json:"links,omitempty"`
	Maintainers                 []MaintainerMapping   `json:"maintainers,omitempty"`
	Relationships               []RelationshipMapping `json:"relationships,omitempty"`
	ReleaseDateTemplate         string                `json:"releaseDateTemplate,omitempty"`
	RepositoryURLTemplate       string                `json:"repositoryUrlTemplate,omitempty"`
	SupportedByTemplate         string                `json:"supportedByTemplate,omitempty"`
	SupportedByImageURLTemplate string                `json:"supportedByImageUrlTemplate,omitempty"`
	Version                     *VersionMapping       `json:"version,omitempty"`
	Visibility                  *VisibilityMapping    `json:"visibility,omitempty"`

	// Legacy catalogMetadata support for backward compatibility
	CatalogMetadata *CatalogMetadataMapping `json:"catalogMetadata,omitempty"`
}

type CatalogMetadataMapping struct {
	DisplayName   string                `json:"displayName,omitempty"`
	Description   string                `json:"description,omitempty"`
	Icon          *IconMapping          `json:"icon,omitempty"`
	Documentation *DocumentationMapping `json:"documentation,omitempty"`
	Labels        map[string]string     `json:"labels,omitempty"`
	Annotations   map[string]string     `json:"annotations,omitempty"`
	Tags          []string              `json:"tags,omitempty"`
	Links         []LinkMapping         `json:"links,omitempty"`
	Maintainers   []MaintainerMapping   `json:"maintainers,omitempty"`
	Publisher     *PublisherMapping     `json:"publisher,omitempty"`
}

type IconMapping struct {
	MediaType  string `json:"mediaType,omitempty"`
	Base64Data string `json:"base64Data,omitempty"`
}

type DocumentationMapping struct {
	Type string `json:"type,omitempty"`
	URL  string `json:"url,omitempty"`
}

type LinkMapping struct {
	DisplayName string `json:"displayName,omitempty"`
	URL         string `json:"url,omitempty"`
}

type MaintainerMapping struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

type PublisherMapping struct {
	Name  string       `json:"name,omitempty"`
	URL   string       `json:"url,omitempty"`
	Image *IconMapping `json:"image,omitempty"`
}

type RelationshipMapping struct {
	Type   string `json:"type,omitempty"`
	Target string `json:"target,omitempty"`
}

type VersionMapping struct {
	Name        string `json:"name,omitempty"`
	ReleaseNote string `json:"releaseNote,omitempty"`
}

type VisibilityMapping struct {
	Public     bool `json:"public"`
	AllTenants bool `json:"allTenants"`
}

func (c *Config) Validate() error {
	if c.URL == "" {
		return ErrURLNotSet
	}

	if _, err := url.Parse(c.URL); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidURL, err)
	}

	if c.TenantID == "" {
		return fmt.Errorf("%w: tenantId", ErrMissingField)
	}

	if c.ItemTypeDefinitionRef.Name == "" {
		return fmt.Errorf("%w: itemTypeDefinitionRef.name", ErrMissingField)
	}

	if c.ItemTypeDefinitionRef.Namespace == "" {
		return fmt.Errorf("%w: itemTypeDefinitionRef.namespace", ErrMissingField)
	}

	if c.ClientID == "" {
		return fmt.Errorf("%w: clientId", ErrMissingField)
	}

	if c.ClientSecret.String() == "" {
		return fmt.Errorf("%w: clientSecret", ErrMissingField)
	}

	if c.ItemNameTemplate == "" {
		return fmt.Errorf("%w: itemNameTemplate", ErrMissingField)
	}

	if c.ItemLifecycleStatus == "" {
		c.ItemLifecycleStatus = consoleclient.Published
	}

	if !consoleclient.IsValidLifecycleStatus(string(c.ItemLifecycleStatus)) {
		return fmt.Errorf("%w: %s", ErrInvalidLifecycleStatus, c.ItemLifecycleStatus)
	}

	return nil
}
