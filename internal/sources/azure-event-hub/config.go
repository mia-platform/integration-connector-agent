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

package azureeventhub

import (
	"fmt"
	"strings"

	"github.com/mia-platform/integration-connector-agent/internal/config"
)

type Config struct {
	SubscriptionID                 string              `json:"subscriptionId"`
	EventHubNamespace              string              `json:"eventHubNamespace"`
	EventHubName                   string              `json:"eventHubName"`
	CheckpointStorageAccountName   string              `json:"checkpointStorageAccountName"`
	CheckpointStorageContainerName string              `json:"checkpointStorageContainerName"`
	TenantID                       string              `json:"tenantId"`
	ClientID                       config.SecretSource `json:"clientId"`
	ClientSecret                   config.SecretSource `json:"clientSecret"`
	EventConsumer                  EventConsumer
}

func (c *Config) Validate() error { //nolint: gocyclo
	if c.SubscriptionID == "" {
		return fmt.Errorf("subscriptionId is required")
	}

	if c.EventHubNamespace == "" {
		return fmt.Errorf("eventHubNamespace is required")
	}

	if c.EventHubName == "" {
		return fmt.Errorf("eventHubName is required")
	}

	if c.CheckpointStorageAccountName == "" {
		return fmt.Errorf("checkpointStorageAccountName is required")
	}

	if c.CheckpointStorageContainerName == "" {
		return fmt.Errorf("checkpointStorageContainerName is required")
	}

	switch {
	case len(c.TenantID) > 0:
		if len(c.ClientID) == 0 || len(c.ClientSecret) == 0 {
			return fmt.Errorf("clientId and clientSecret are required when tenantId is provided")
		}
	case len(c.ClientID) > 0:
		if len(c.TenantID) == 0 || len(c.ClientSecret) == 0 {
			return fmt.Errorf("tenantId and clientSecret are required when clientId is provided")
		}
	case len(c.ClientSecret) > 0:
		if len(c.TenantID) == 0 || len(c.ClientID) == 0 {
			return fmt.Errorf("tenantId and clientId are required when clientSecret is provided")
		}
	}

	if !strings.HasSuffix(c.EventHubNamespace, ".servicebus.windows.net") {
		c.EventHubNamespace = fmt.Sprintf("%s.servicebus.windows.net", c.EventHubNamespace)
	}

	if !strings.HasSuffix(c.CheckpointStorageAccountName, "https://") {
		c.CheckpointStorageAccountName = fmt.Sprintf("https://%s.blob.core.windows.net", c.CheckpointStorageAccountName)
	}

	return nil
}
