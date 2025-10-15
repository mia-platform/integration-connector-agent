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

	"github.com/sirupsen/logrus"
)

type Writer[T entities.PipelineEvent] struct {
	config *Config
	client consoleclient.CatalogClient[any]
	log    *logrus.Logger
}

func NewWriter[T entities.PipelineEvent](config *Config, log *logrus.Logger) (sinks.Sink[T], error) {
	tokenManager, err := consoleclient.NewClientCredentialsTokenManager(config.URL, config.ClientID, config.ClientSecret.String())
	if err != nil {
		return nil, fmt.Errorf("error creating catalog client: %w", err)
	}

	client := consoleclient.New[any](config.URL, tokenManager)
	return &Writer[T]{
		client: client,
		config: config,
		log:    log,
	}, nil
}

func (w *Writer[T]) WriteData(ctx context.Context, event T) error {
	w.log.WithFields(logrus.Fields{
		"sinkType":    "console-catalog",
		"eventType":   event.GetType(),
		"primaryKeys": event.GetPrimaryKeys().Map(),
		"operation":   event.Operation(),
	}).Debug("starting to write event to console catalog sink")

	if event.Operation() == entities.Delete {
		itemID, err := w.getItemID(event)
		if err != nil {
			w.log.WithFields(logrus.Fields{
				"sinkType":    "console-catalog",
				"eventType":   event.GetType(),
				"primaryKeys": event.GetPrimaryKeys().Map(),
				"operation":   "delete",
			}).WithError(err).Error("failed to get item ID for console catalog delete operation")
			return fmt.Errorf("error processing item ID template: %w", err)
		}

		if err := w.client.Delete(ctx, w.config.TenantID, itemID); err != nil {
			w.log.WithFields(logrus.Fields{
				"sinkType":    "console-catalog",
				"eventType":   event.GetType(),
				"primaryKeys": event.GetPrimaryKeys().Map(),
				"itemId":      itemID,
				"operation":   "delete",
			}).WithError(err).Error("failed to delete item from console catalog")
			return err
		}

		w.log.WithFields(logrus.Fields{
			"sinkType":    "console-catalog",
			"eventType":   event.GetType(),
			"primaryKeys": event.GetPrimaryKeys().Map(),
			"itemId":      itemID,
			"operation":   "delete",
		}).Debug("successfully deleted item from console catalog")
		return nil
	}

	item, err := w.createCatalogItem(event)
	if err != nil {
		w.log.WithFields(logrus.Fields{
			"sinkType":    "console-catalog",
			"eventType":   event.GetType(),
			"primaryKeys": event.GetPrimaryKeys().Map(),
			"operation":   "upsert",
		}).WithError(err).Error("failed to create catalog item")
		return fmt.Errorf("error creating catalog item: %w", err)
	}

	if _, err := w.client.Apply(ctx, item); err != nil {
		w.log.WithFields(logrus.Fields{
			"sinkType":    "console-catalog",
			"eventType":   event.GetType(),
			"primaryKeys": event.GetPrimaryKeys().Map(),
			"itemId":      item.ItemID,
			"operation":   "upsert",
		}).WithError(err).Error("failed to apply item to console catalog")
		return err
	}

	w.log.WithFields(logrus.Fields{
		"sinkType":    "console-catalog",
		"eventType":   event.GetType(),
		"primaryKeys": event.GetPrimaryKeys().Map(),
		"itemId":      item.ItemID,
		"operation":   "upsert",
	}).Debug("successfully applied item to console catalog")

	return nil
}

func (w *Writer[T]) Close(_ context.Context) error {
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

	catalogMetadata, err := w.createCatalogMetadata(event)
	if err != nil {
		return nil, fmt.Errorf("error creating catalog metadata: %w", err)
	}

	marketplaceResource := &consoleclient.MarketplaceResource[any]{
		TenantID:              w.config.TenantID,
		Name:                  itemName,
		ItemID:                itemID,
		ItemTypeDefinitionRef: w.config.ItemTypeDefinitionRef,
		LifecycleStatus:       w.config.ItemLifecycleStatus,
		Resources:             res,
	}

	// Embed catalog metadata fields directly into the marketplace resource
	if catalogMetadata != nil {
		// Set catalog metadata fields directly on the marketplace resource
		if catalogMetadata.Documentation != nil {
			marketplaceResource.Documentation = catalogMetadata.Documentation
		}
		if len(catalogMetadata.Labels) > 0 {
			marketplaceResource.Labels = catalogMetadata.Labels
		}
		if len(catalogMetadata.Links) > 0 {
			marketplaceResource.Links = catalogMetadata.Links
		}
		if len(catalogMetadata.Maintainers) > 0 {
			marketplaceResource.Maintainers = catalogMetadata.Maintainers
		}
		if len(catalogMetadata.Tags) > 0 {
			marketplaceResource.Tags = catalogMetadata.Tags
		}
		if len(catalogMetadata.Annotations) > 0 {
			marketplaceResource.Annotations = catalogMetadata.Annotations
		}
	}

	return marketplaceResource, nil
}

//nolint:gocyclo // Complex metadata creation logic - refactoring would reduce readability
func (w *Writer[T]) createCatalogMetadata(event T) (*consoleclient.CatalogMetadata, error) {
	if w.config.CatalogMetadata == nil {
		return nil, nil
	}

	metadata := &consoleclient.CatalogMetadata{}

	// Process display name
	if w.config.CatalogMetadata.DisplayName != "" {
		displayName, err := templetize(w.config.CatalogMetadata.DisplayName, event.Data())
		if err != nil {
			return nil, fmt.Errorf("error processing displayName template: %w", err)
		}
		metadata.DisplayName = displayName
	}

	// Process description
	if w.config.CatalogMetadata.Description != "" {
		description, err := templetize(w.config.CatalogMetadata.Description, event.Data())
		if err != nil {
			return nil, fmt.Errorf("error processing description template: %w", err)
		}
		metadata.Description = description
	}

	// Process icon
	if w.config.CatalogMetadata.Icon != nil {
		icon := &consoleclient.Icon{}
		if w.config.CatalogMetadata.Icon.MediaType != "" {
			mediaType, err := templetize(w.config.CatalogMetadata.Icon.MediaType, event.Data())
			if err != nil {
				return nil, fmt.Errorf("error processing icon mediaType template: %w", err)
			}
			icon.MediaType = mediaType
		}
		if w.config.CatalogMetadata.Icon.Base64Data != "" {
			base64Data, err := templetize(w.config.CatalogMetadata.Icon.Base64Data, event.Data())
			if err != nil {
				return nil, fmt.Errorf("error processing icon base64Data template: %w", err)
			}
			icon.Base64Data = base64Data
		}
		metadata.Icon = icon
	}

	// Process documentation
	if w.config.CatalogMetadata.Documentation != nil {
		doc := &consoleclient.Documentation{}
		if w.config.CatalogMetadata.Documentation.Type != "" {
			docType, err := templetize(w.config.CatalogMetadata.Documentation.Type, event.Data())
			if err != nil {
				return nil, fmt.Errorf("error processing documentation type template: %w", err)
			}
			doc.Type = docType
		}
		if w.config.CatalogMetadata.Documentation.URL != "" {
			url, err := templetize(w.config.CatalogMetadata.Documentation.URL, event.Data())
			if err != nil {
				return nil, fmt.Errorf("error processing documentation URL template: %w", err)
			}
			doc.URL = url
		}
		metadata.Documentation = doc
	}

	// Process labels
	if len(w.config.CatalogMetadata.Labels) > 0 {
		labels := make(map[string]string)
		for key, valueTemplate := range w.config.CatalogMetadata.Labels {
			value, err := templetize(valueTemplate, event.Data())
			if err != nil {
				return nil, fmt.Errorf("error processing label %s template: %w", key, err)
			}
			labels[key] = value
		}
		metadata.Labels = labels
	}

	// Process annotations
	if len(w.config.CatalogMetadata.Annotations) > 0 {
		annotations := make(map[string]string)
		for key, valueTemplate := range w.config.CatalogMetadata.Annotations {
			value, err := templetize(valueTemplate, event.Data())
			if err != nil {
				return nil, fmt.Errorf("error processing annotation %s template: %w", key, err)
			}
			annotations[key] = value
		}
		metadata.Annotations = annotations
	}

	// Process tags
	if len(w.config.CatalogMetadata.Tags) > 0 {
		tags := make([]string, len(w.config.CatalogMetadata.Tags))
		for i, tagTemplate := range w.config.CatalogMetadata.Tags {
			tag, err := templetize(tagTemplate, event.Data())
			if err != nil {
				return nil, fmt.Errorf("error processing tag template: %w", err)
			}
			tags[i] = tag
		}
		metadata.Tags = tags
	}

	// Process links
	if len(w.config.CatalogMetadata.Links) > 0 {
		links := make([]consoleclient.Link, len(w.config.CatalogMetadata.Links))
		for i, linkMapping := range w.config.CatalogMetadata.Links {
			link := consoleclient.Link{}
			if linkMapping.DisplayName != "" {
				displayName, err := templetize(linkMapping.DisplayName, event.Data())
				if err != nil {
					return nil, fmt.Errorf("error processing link displayName template: %w", err)
				}
				link.DisplayName = displayName
			}
			if linkMapping.URL != "" {
				url, err := templetize(linkMapping.URL, event.Data())
				if err != nil {
					return nil, fmt.Errorf("error processing link URL template: %w", err)
				}
				link.URL = url
			}
			links[i] = link
		}
		metadata.Links = links
	}

	// Process maintainers
	if len(w.config.CatalogMetadata.Maintainers) > 0 {
		maintainers := make([]consoleclient.Maintainer, len(w.config.CatalogMetadata.Maintainers))
		for i, maintainerMapping := range w.config.CatalogMetadata.Maintainers {
			maintainer := consoleclient.Maintainer{}
			if maintainerMapping.Name != "" {
				name, err := templetize(maintainerMapping.Name, event.Data())
				if err != nil {
					return nil, fmt.Errorf("error processing maintainer name template: %w", err)
				}
				maintainer.Name = name
			}
			if maintainerMapping.Email != "" {
				email, err := templetize(maintainerMapping.Email, event.Data())
				if err != nil {
					return nil, fmt.Errorf("error processing maintainer email template: %w", err)
				}
				maintainer.Email = email
			}
			maintainers[i] = maintainer
		}
		metadata.Maintainers = maintainers
	}

	// Process publisher
	if w.config.CatalogMetadata.Publisher != nil {
		publisher := &consoleclient.Publisher{}
		if w.config.CatalogMetadata.Publisher.Name != "" {
			name, err := templetize(w.config.CatalogMetadata.Publisher.Name, event.Data())
			if err != nil {
				return nil, fmt.Errorf("error processing publisher name template: %w", err)
			}
			publisher.Name = name
		}
		if w.config.CatalogMetadata.Publisher.URL != "" {
			url, err := templetize(w.config.CatalogMetadata.Publisher.URL, event.Data())
			if err != nil {
				return nil, fmt.Errorf("error processing publisher URL template: %w", err)
			}
			publisher.URL = url
		}
		if w.config.CatalogMetadata.Publisher.Image != nil {
			image := &consoleclient.Icon{}
			if w.config.CatalogMetadata.Publisher.Image.MediaType != "" {
				mediaType, err := templetize(w.config.CatalogMetadata.Publisher.Image.MediaType, event.Data())
				if err != nil {
					return nil, fmt.Errorf("error processing publisher image mediaType template: %w", err)
				}
				image.MediaType = mediaType
			}
			if w.config.CatalogMetadata.Publisher.Image.Base64Data != "" {
				base64Data, err := templetize(w.config.CatalogMetadata.Publisher.Image.Base64Data, event.Data())
				if err != nil {
					return nil, fmt.Errorf("error processing publisher image base64Data template: %w", err)
				}
				image.Base64Data = base64Data
			}
			publisher.Image = image
		}
		metadata.Publisher = publisher
	}

	return metadata, nil
}

func (w *Writer[T]) getItemIDFromEvent(event T) (string, error) {
	if w.config.ItemIDTemplate == "" {
		var itemIDBuilder strings.Builder
		pks := event.GetPrimaryKeys()
		for i, pk := range pks {
			fmt.Fprintf(&itemIDBuilder, "%s-%s", slugify(pk.Key), slugify(pk.Value))
			if i != len(pks)-1 {
				itemIDBuilder.WriteString("-")
			}
		}
		return itemIDBuilder.String(), nil
	}

	itemID, err := templetize(w.config.ItemIDTemplate, event.Data())
	if err != nil {
		return "", fmt.Errorf("error processing item ID template: %w", err)
	}
	return slugify(itemID), nil
}

func (w *Writer[T]) getItemID(event T) (string, error) {
	itemID, err := w.getItemIDFromEvent(event)
	if err != nil {
		return "", fmt.Errorf("error getting item ID from event: %w", err)
	}
	digest := digestForCatalog63Bytes([]byte(itemID))

	w.log.WithFields(logrus.Fields{
		"itemId":      itemID,
		"digest":      digest,
		"primaryKeys": event.GetPrimaryKeys().Map(),
	}).Trace("itemId for Console Catalog")

	return digest, nil
}
