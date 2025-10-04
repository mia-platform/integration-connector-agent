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
