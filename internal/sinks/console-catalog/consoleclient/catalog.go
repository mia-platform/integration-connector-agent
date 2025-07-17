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
	"encoding/json"
	"fmt"
	"net/http"
)

type marketplacePostExtensionBody[Resource any] struct {
	Resources []MarketplaceResource[Resource] `json:"resources"`
}
type responseItem struct {
	ItemID           string            `json:"itemId"`
	ValidationErrors []ValidationError `json:"validationErrors"`
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

	targetURL := fmt.Sprintf("%sapi/marketplace/tenants/%s/resources", c.url, item.TenantID)
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
		return "", fmt.Errorf("%w: %s", ErrMarketplaceResponseParse, err)
	}

	if !responseBody.Done {
		errors := make([]string, 0)
		for _, validationErr := range responseBody.Items[0].ValidationErrors {
			errors = append(errors, validationErr.Message)
		}

		return "", &MarketplaceValidationError{Errors: errors}
	}

	return responseBody.Items[0].ItemID, nil
}
