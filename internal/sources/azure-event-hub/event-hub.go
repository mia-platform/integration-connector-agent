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

package azureeventhub

import (
	"context"

	"github.com/mia-platform/integration-connector-agent/internal/azure"
	"github.com/mia-platform/integration-connector-agent/internal/azure/eventhub"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs/v2"
	"github.com/sirupsen/logrus"
)

func SetupEventHub(ctx context.Context, config azure.EventHubConfig, logger *logrus.Logger) {
	consumerClient, err := eventhub.NewConsumerClient(config, azeventhubs.DefaultConsumerGroup)
	if err != nil {
		logger.WithError(err).Error("error initializing azure event hub consumer client")
		logger.Error("Azure Event Hub configuration help:\n" +
			"1. Verify Event Hub namespace and name are correct\n" +
			"2. Ensure service principal has 'Azure Event Hubs Data Receiver' role\n" +
			"3. Check Azure credentials (AZURE_TENANT_ID, AZURE_CLIENT_ID, AZURE_CLIENT_SECRET)\n" +
			"4. Confirm Event Hub exists and is accessible")
		return
	}
	defer consumerClient.Close(ctx)

	if err := eventhub.RunEventHubProcessor(ctx, config, consumerClient, logger); err != nil {
		logger.WithError(err).Error("error running azure event hub processor")
		logger.Error("Azure Event Hub processing help:\n" +
			"1. Verify checkpoint storage account exists and is accessible\n" +
			"2. Ensure service principal has 'Storage Blob Data Contributor' role on storage account\n" +
			"3. Check that the checkpoint container exists in the storage account\n" +
			"4. Verify network connectivity to Azure services")
	}
}
