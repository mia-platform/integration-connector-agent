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
	"strings"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/sinks"
	"github.com/mia-platform/integration-connector-agent/internal/sinks/console-catalog/consoleclient"
)

type Writer[T entities.PipelineEvent] struct {
	config *Config
	client consoleclient.CatalogClient[any]
}

func NewWriter[T entities.PipelineEvent](config *Config) (sinks.Sink[T], error) {
	tokenManager := consoleclient.NewClientCredentialsTokenManager(config.URL, config.ClientID, config.ClientSecret.String())
	client := consoleclient.New[any](config.URL, tokenManager)
	return &Writer[T]{
		client: client,
		config: config,
	}, nil
}

func (w *Writer[T]) WriteData(ctx context.Context, event T) error {
	if event.Operation() == entities.Delete {
		itemID, err := w.getItemID(event)
		if err != nil {
			return fmt.Errorf("error processing item ID template: %w", err)
		}
		return w.client.Delete(ctx, w.config.TenantID, itemID)
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

func (w *Writer[T]) createCatalogItem(event T) (*consoleclient.MarketplaceResource[any], error) {
	res := make(map[string]any)
	if err := json.Unmarshal(event.Data(), &res); err != nil {
		return nil, err
	}

	itemID, err := w.getItemID(event)
	if err != nil {
		return nil, fmt.Errorf("error processing item ID template: %w", err)
	}

	itemName, err := templetize(w.config.ItemNameTemplate, event.Data())
	if err != nil {
		return nil, fmt.Errorf("error processing item name template: %w", err)
	}

	return &consoleclient.MarketplaceResource[any]{
		TenantID:        w.config.TenantID,
		Name:            itemName,
		ItemID:          slugify(itemID),
		Type:            w.config.ItemType,
		LifecycleStatus: w.config.ItemLifecycleStatus,
		Resources:       res,
	}, nil
}

func (w *Writer[T]) getItemID(event T) (string, error) {
	if w.config.ItemIDTemplate == "" {
		var itemIdBuilder strings.Builder
		pks := event.GetPrimaryKeys()
		for i, pk := range pks {
			fmt.Fprintf(&itemIdBuilder, "%s-%s", slugify(pk.Key), slugify(pk.Value))
			if i != len(pks)-1 {
				itemIdBuilder.WriteString("-")
			}
		}
		return itemIdBuilder.String(), nil
	}

	itemID, err := templetize(w.config.ItemIDTemplate, event.Data())
	if err != nil {
		return "", fmt.Errorf("error processing item ID template: %w", err)
	}
	return slugify(itemID), nil
}
