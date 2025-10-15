// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package consoleclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type marketplacePostExtensionBody[Resource any] struct {
	Resources []MarketplaceResource[Resource] `json:"resources"`
}
type responseItem struct {
	ItemID string            `json:"itemId"`
	Errors []ValidationError `json:"errors"`
}

type marketplacePostExtensionResponse struct {
	Done  bool           `json:"done"`
	Items []responseItem `json:"items"`
}

func (c *consoleClient[T]) Apply(ctx context.Context, item *MarketplaceResource[T]) (string, error) {
	marketplacePostExtension := marketplacePostExtensionBody[T]{
		Resources: []MarketplaceResource[T]{
			*item,
		},
	}

	targetURL := fmt.Sprintf("%sapi/tenants/%s/marketplace/items", c.url, item.TenantID)
	resp, err := c.fireRequest(ctx, http.MethodPost, targetURL, marketplacePostExtension)
	if err != nil {
		return "", fmt.Errorf("error applying resource: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to apply resource, status code: %d", resp.StatusCode)
	}

	var responseBody marketplacePostExtensionResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return "", fmt.Errorf("%w: %w", ErrMarketplaceResponseParse, err)
	}

	if !responseBody.Done {
		errors := make([]string, 0)
		for _, validationErr := range responseBody.Items[0].Errors {
			errors = append(errors, validationErr.Message)
		}

		return "", &MarketplaceValidationError{Errors: errors}
	}

	return responseBody.Items[0].ItemID, nil
}

func (c *consoleClient[T]) Delete(ctx context.Context, tenantID string, itemID string) error {
	targetURL := fmt.Sprintf("%sapi/tenants/%s/marketplace/items/%s/versions/NA", c.url, tenantID, itemID)
	resp, err := c.fireRequest(ctx, http.MethodDelete, targetURL, nil)
	if err != nil {
		return fmt.Errorf("error deleting resource: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete resource, status code: %d", resp.StatusCode)
	}

	return nil
}
