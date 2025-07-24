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
	"testing"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/sinks/console-catalog/consoleclient"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestWriteData(t *testing.T) {
	log, _ := test.NewNullLogger()

	t.Run("should return error on invalid data", func(t *testing.T) {
		writer, err := NewWriter[entities.PipelineEvent](&Config{
			URL:          "http://example.com",
			TenantID:     "tenant-id",
			ItemType:     "item-type",
			ClientID:     "client-id",
			ClientSecret: "secret",
		}, log)
		require.NoError(t, err)

		evt := &entities.Event{
			OriginalRaw: []byte(`invalidata`),
		}
		require.ErrorContains(t, writer.WriteData(context.Background(), evt), "error creating catalog item")
	})

	t.Run("should invoke apply with correct item", func(t *testing.T) {
		mockClient := &mockConsoleClient{
			ApplyResult: "item-id",
			ApplyAssert: func(ctx context.Context, item *consoleclient.MarketplaceResource[any]) {
				require.Equal(t, "tenant-id", item.TenantID)
				require.Equal(t, "item-type", item.Type)

				require.Equal(t, "bd2700071a46b945e610d1fad65eff454595a9ac", item.ItemID)
				require.Equal(t, "Test Name", item.Name)
			},
		}

		writer := &Writer[entities.PipelineEvent]{
			client: mockClient,
			log:    log,
			config: &Config{
				URL:              "http://example.com",
				TenantID:         "tenant-id",
				ItemType:         "item-type",
				ClientID:         "client-id",
				ClientSecret:     "secret",
				ItemIDTemplate:   "{{name}}-{{assetId}}",
				ItemNameTemplate: "{{name}}",
			},
		}

		evt := &entities.Event{
			OriginalRaw: []byte(`{"name": "Test Name","assetId": "12345"}`),
		}
		err := writer.WriteData(context.Background(), evt)
		require.NoError(t, err)
	})

	t.Run("should invoke delete with correct parameters", func(t *testing.T) {
		mockClient := &mockConsoleClient{
			DeleteAssert: func(ctx context.Context, tenantID string, itemID string) {
				require.Equal(t, "tenant-id", tenantID)
				require.Equal(t, "286be376dc09ca9196049a2ae222f36b6303b1f3", itemID)
			},
		}

		writer := &Writer[entities.PipelineEvent]{
			log:    log,
			client: mockClient,
			config: &Config{
				URL:            "http://example.com",
				TenantID:       "tenant-id",
				ItemType:       "item-type",
				ClientID:       "client-id",
				ClientSecret:   "secret",
				ItemIDTemplate: "{{name}}-{{assetId}}",
			},
		}

		err := writer.WriteData(context.Background(), &entities.Event{
			OperationType: entities.Delete,
			OriginalRaw:   []byte(`{"name": "The Name", "assetId": "the-asset-id"}`),
		})
		require.NoError(t, err)
	})

	t.Run("should invoke delete with correct itemId from primary keys when no itemIdTemplate is set", func(t *testing.T) {
		mockClient := &mockConsoleClient{
			DeleteAssert: func(ctx context.Context, tenantID string, itemID string) {
				require.Equal(t, "tenant-id", tenantID)
				require.Equal(t, "f52268d49ce3927826a9ed23465bf68e26282065", itemID)
			},
		}

		writer := &Writer[entities.PipelineEvent]{
			client: mockClient,
			log:    log,
			config: &Config{
				URL:              "http://example.com",
				TenantID:         "tenant-id",
				ItemType:         "item-type",
				ClientID:         "client-id",
				ClientSecret:     "secret",
				ItemIDTemplate:   "",
				ItemNameTemplate: "the-name-{{assetId}}",
			},
		}

		err := writer.WriteData(context.Background(), &entities.Event{
			OperationType: entities.Delete,
			PrimaryKeys: entities.PkFields{
				entities.PkField{
					Key:   "assetId",
					Value: "the-asset-id",
				},
				entities.PkField{
					Key:   "resourceId",
					Value: "/subscriptions/mysubscription/resourcegroups/myresourcegroup/providers/microsoft.web/sites/myappservice",
				},
			},
			OriginalRaw: []byte(`{"name": "The Name", "assetId": "the-asset-id"}`),
		})
		require.NoError(t, err)
	})
}

type mockConsoleClient struct {
	ApplyResult string
	ApplyError  error
	ApplyAssert func(ctx context.Context, item *consoleclient.MarketplaceResource[any])

	DeleteAssert func(ctx context.Context, tenantID string, itemID string)
	DeleteError  error
}

func (m *mockConsoleClient) Apply(ctx context.Context, item *consoleclient.MarketplaceResource[any]) (string, error) {
	if m.ApplyAssert != nil {
		m.ApplyAssert(ctx, item)
	}
	return m.ApplyResult, m.ApplyError
}

func (m *mockConsoleClient) Delete(ctx context.Context, tenantID string, itemID string) error {
	if m.DeleteAssert != nil {
		m.DeleteAssert(ctx, tenantID, itemID)
	}

	return m.DeleteError
}
