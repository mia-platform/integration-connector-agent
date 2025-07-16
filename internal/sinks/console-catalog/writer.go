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

package consolecatalog

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/sinks"
)

type Writer[T entities.PipelineEvent] struct {
	config *Config
	client IClient[any]
}

func NewWriter[T entities.PipelineEvent](config *Config) (sinks.Sink[T], error) {
	tokenManager := newClientCredentialsTokenManager(config.URL, config.ClientID, config.ClientSecret.String())
	client := newConsoleClient[any](config.URL, tokenManager)
	return &Writer[T]{
		client: client,
		config: config,
	}, nil
}

func (w *Writer[T]) WriteData(ctx context.Context, event T) error {
	if event.Operation() == entities.Delete {
		return fmt.Errorf("console-catalog sink does not support delete operations")
	}

	item, err := w.createCatalogItem(event)
	if err != nil {
		return fmt.Errorf("error creating catalog item: %w", err)
	}

	if _, err := w.client.Apply(ctx, item); err != nil {
		return err
	}
	return nil
}

func (w *Writer[T]) createCatalogItem(event T) (*MarketplaceResource[any], error) {
	res := make(map[string]any)
	if err := json.Unmarshal(event.Data(), &res); err != nil {
		return nil, err
	}

	return &MarketplaceResource[any]{
		TenantID:  w.config.TenantID,
		Name:      "TODO",
		ItemID:    "TODO IID",
		Type:      w.config.ItemType,
		Resources: res,
	}, nil
}
