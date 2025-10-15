// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package consoleclient

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrMarketplaceResponseParse    = errors.New("failed to unmarshal marketplace response body")
	ErrMarketplaceRequestCreation  = errors.New("failed to create marketplace request")
	ErrMarketplaceRequestExecution = errors.New("failed to execute marketplace request")
	ErrMarketplaceRequestBodyParse = errors.New("failed to prepare marketplace request")
)

type LifecycleStatus string

const (
	ComingSoon  LifecycleStatus = "coming-soon"
	Draft       LifecycleStatus = "draft"
	Published   LifecycleStatus = "published"
	Maintenance LifecycleStatus = "maintenance"
	Deprecated  LifecycleStatus = "deprecated"
	Archived    LifecycleStatus = "archived"
)

func IsValidLifecycleStatus(status string) bool {
	switch LifecycleStatus(status) {
	case ComingSoon, Draft, Published, Maintenance, Deprecated, Archived:
		return true
	}
	return false
}

type Resource any

type CatalogClient[T Resource] interface {
	Apply(ctx context.Context, item *MarketplaceResource[T]) (string, error)
	Delete(ctx context.Context, tenantID string, itemID string) error
}

type ItemTypeDefinitionRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type MarketplaceResource[T Resource] struct {
	ID                    string                `json:"_id,omitempty"` //nolint:tagliatelle
	ItemID                string                `json:"itemId"`
	Name                  string                `json:"name"`
	ItemTypeDefinitionRef ItemTypeDefinitionRef `json:"itemTypeDefinitionRef"`
	Description           string                `json:"description"`
	TenantID              string                `json:"tenantId"`
	LifecycleStatus       LifecycleStatus       `json:"lifecycleStatus"`
	Resources             T                     `json:"resources"`

	// Catalog Metadata fields - moved from separate metadata object to be directly embedded
	CategoryID          string            `json:"categoryId,omitempty"`
	Documentation       *Documentation    `json:"documentation,omitempty"`
	ImageURL            string            `json:"imageUrl,omitempty"`
	Labels              map[string]string `json:"labels,omitempty"`
	Links               []Link            `json:"links,omitempty"`
	Maintainers         []Maintainer      `json:"maintainers,omitempty"`
	Relationships       []Relationship    `json:"relationships,omitempty"`
	ReleaseDate         string            `json:"releaseDate,omitempty"`
	RepositoryURL       string            `json:"repositoryUrl,omitempty"`
	SupportedBy         string            `json:"supportedBy,omitempty"`
	SupportedByImageURL string            `json:"supportedByImageUrl,omitempty"`
	Tags                []string          `json:"tags,omitempty"`
	Version             *Version          `json:"version,omitempty"`
	Visibility          *Visibility       `json:"visibility,omitempty"`
	Annotations         map[string]string `json:"annotations,omitempty"`
}

type CatalogMetadata struct {
	DisplayName   string            `json:"displayName,omitempty"`
	Description   string            `json:"description,omitempty"`
	Icon          *Icon             `json:"icon,omitempty"`
	Documentation *Documentation    `json:"documentation,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	Annotations   map[string]string `json:"annotations,omitempty"`
	Tags          []string          `json:"tags,omitempty"`
	Links         []Link            `json:"links,omitempty"`
	Maintainers   []Maintainer      `json:"maintainers,omitempty"`
	Publisher     *Publisher        `json:"publisher,omitempty"`
}

type Icon struct {
	MediaType  string `json:"mediaType,omitempty"`
	Base64Data string `json:"base64Data,omitempty"`
}

type Documentation struct {
	Type string `json:"type,omitempty"`
	URL  string `json:"url,omitempty"`
}

type Link struct {
	DisplayName string `json:"displayName,omitempty"`
	URL         string `json:"url,omitempty"`
}

type Maintainer struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

type Relationship struct {
	Type   string `json:"type,omitempty"`
	Target string `json:"target,omitempty"`
}

type Version struct {
	Name        string `json:"name,omitempty"`
	ReleaseNote string `json:"releaseNote,omitempty"`
}

type Visibility struct {
	Public     bool `json:"public"`
	AllTenants bool `json:"allTenants"`
}

type Publisher struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Image *Icon  `json:"image,omitempty"`
}

type ValidationError struct {
	Message string `json:"message"`
}

type MarketplaceValidationError struct {
	Errors []string
}

func (e *MarketplaceValidationError) Error() string {
	return "invalid catalog item, validation errors: " + strings.Join(e.Errors, ", ")
}
