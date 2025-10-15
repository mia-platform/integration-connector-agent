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

package azure

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs/v2"
	"github.com/mia-platform/integration-connector-agent/internal/config"
)

var (
	ErrAzureClientSecretCredential = errors.New("failed to create an Azure client secret credential")
	ErrAzureDefaultCredential      = errors.New("failed to create a default Azure credential")

	ErrMissingSubscriptionID               = errors.New("subscriptionId is required")
	ErrIncompleteAuthConfigForTenantID     = errors.New("both clientId and clientSecret are required when tenantId is provided")
	ErrIncompleteAuthConfigForClientID     = errors.New("both tenantId and clientSecret are required when clientId is provided")
	ErrIncompleteAuthConfigForClientSecret = errors.New("both tenantId and clientId are required when clientSecret is provided")

	ErrEventHubNamespaceRequired              = errors.New("eventHubNamespace is required")
	ErrEventHubNameRequired                   = errors.New("eventHubName is required")
	ErrCheckpointStorageAccountNameRequired   = errors.New("checkpointStorageAccountName is required")
	ErrCheckpointStorageContainerNameRequired = errors.New("checkpointStorageContainerName is required")
)

type AuthConfig struct {
	SubscriptionID string              `json:"subscriptionId,omitempty"`
	TenantID       string              `json:"tenantId,omitempty"`
	ClientID       config.SecretSource `json:"clientId,omitempty"`
	ClientSecret   config.SecretSource `json:"clientSecret,omitempty"`
}

func (c *AuthConfig) Validate() error {
	switch {
	case c.SubscriptionID == "":
		return ErrMissingSubscriptionID
	case len(c.TenantID) > 0:
		if len(c.ClientID.String()) == 0 || len(c.ClientSecret.String()) == 0 {
			return ErrIncompleteAuthConfigForTenantID
		}
	case len(c.ClientID) > 0:
		if len(c.TenantID) == 0 || len(c.ClientSecret.String()) == 0 {
			return ErrIncompleteAuthConfigForClientID
		}
	case len(c.ClientSecret) > 0:
		if len(c.TenantID) == 0 || len(c.ClientID.String()) == 0 {
			return ErrIncompleteAuthConfigForClientSecret
		}
	}

	return nil
}

func (c *AuthConfig) AzureTokenProvider() (azcore.TokenCredential, error) {
	credentials := make([]azcore.TokenCredential, 0)

	if len(c.TenantID) > 0 && len(c.ClientID.String()) > 0 && len(c.ClientSecret.String()) > 0 {
		secretCredential, err := azidentity.NewClientSecretCredential(
			c.TenantID,
			c.ClientID.String(),
			c.ClientSecret.String(),
			nil, // Options
		)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrAzureClientSecretCredential, err)
		}
		credentials = append(credentials, secretCredential)
	}

	defaultCredential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrAzureDefaultCredential, err)
	}
	credentials = append(credentials, defaultCredential)

	return azidentity.NewChainedTokenCredential(credentials, nil)
}

type EventConsumer func(event *azeventhubs.ReceivedEventData) error

type EventHubConfig struct {
	AuthConfig

	EventHubNamespace              string        `json:"eventHubNamespace,omitempty"`
	EventHubName                   string        `json:"eventHubName,omitempty"`
	CheckpointStorageAccountName   string        `json:"checkpointStorageAccountName,omitempty"`
	CheckpointStorageContainerName string        `json:"checkpointStorageContainerName,omitempty"`
	EventConsumer                  EventConsumer `json:"-"`
}

func (c *EventHubConfig) Validate() error {
	if err := c.AuthConfig.Validate(); err != nil {
		return err
	}

	switch {
	case c.EventHubNamespace == "":
		return ErrEventHubNamespaceRequired
	case c.EventHubName == "":
		return ErrEventHubNameRequired
	case c.CheckpointStorageAccountName == "":
		return ErrCheckpointStorageAccountNameRequired
	case c.CheckpointStorageContainerName == "":
		return ErrCheckpointStorageContainerNameRequired
	}

	if !strings.HasSuffix(c.EventHubNamespace, ".servicebus.windows.net") {
		c.EventHubNamespace += ".servicebus.windows.net"
	}

	if !strings.HasSuffix(c.CheckpointStorageAccountName, ".blob.core.windows.net") {
		c.CheckpointStorageAccountName = fmt.Sprintf("https://%s.blob.core.windows.net", c.CheckpointStorageAccountName)
	}

	return nil
}
