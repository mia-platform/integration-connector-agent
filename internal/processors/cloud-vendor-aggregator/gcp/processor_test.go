// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package gcpaggregator

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/commons"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/gcp/clients/runservice"
	storageclient "github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/gcp/clients/storage"
	gcppubsubevents "github.com/mia-platform/integration-connector-agent/internal/sources/gcp-pubsub/events"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestProcessWithStandardAssetInventoryEvent(t *testing.T) {
	l, _ := test.NewNullLogger()

	testCases := []struct {
		name                 string
		eventFilePath        string
		expectedAsset        commons.Asset
		mockRunServiceClient runservice.Client
		mockStorageClient    storageclient.Client
	}{
		{
			name:          "Create run service event",
			eventFilePath: "testdata/func-create.json",
			expectedAsset: commons.Asset{
				Name:     "test-function",
				Type:     "run.googleapis.com/Service",
				Provider: commons.GCPAssetProvider,
				Location: "europe-west1",
				Relationships: []string{
					"projects/123123123123",
					"folders/456456456456",
					"organizations/789789789789",
				},
				Tags: commons.Tags{"label1": "value1", "label2": "value2"},
			},
			mockRunServiceClient: &MockFnService{
				GetServiceResult: &runservice.Service{
					Name:   "projects/the-project/locations/europe-west1/services/test-function",
					Labels: map[string]string{"label1": "value1", "label2": "value2"},
				},
				GetServiceAssert: func(_ context.Context, name string) {
					require.Equal(t, "projects/the-project/locations/europe-west1/services/test-function", name)
				},
			},
		},
		{
			name:          "Create bucket event",
			eventFilePath: "testdata/bucket-create.json",
			expectedAsset: commons.Asset{
				Name:     "test-bucket",
				Type:     "storage.googleapis.com/Bucket",
				Provider: commons.GCPAssetProvider,
				Location: "europe-west1",
				Relationships: []string{
					"projects/123123123123",
					"folders/456456456456",
					"organizations/789789789789",
				},
				Tags: commons.Tags{"label1": "value1", "label2": "value2"},
			},
			mockStorageClient: &MockStorageClient{
				GetBucketResult: &storageclient.Bucket{
					Name:     "test-bucket",
					Location: "europe-west1",
					Labels:   map[string]string{"label1": "value1", "label2": "value2"},
				},
				GetBucketAssert: func(_ context.Context, name string) {
					require.Equal(t, "//storage.googleapis.com/test-bucket", name)
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := &GCPCloudVendorAggregator{
				logger:  l,
				options: nil,
				s:       tc.mockStorageClient,
				f:       tc.mockRunServiceClient,
			}

			d, err := os.ReadFile(tc.eventFilePath)
			require.NoError(t, err)

			builder := gcppubsubevents.NewInventoryEventBuilder[gcppubsubevents.InventoryEvent]()
			event, err := builder.GetPipelineEvent(t.Context(), d)
			require.NoError(t, err)

			result, err := p.Process(event)
			require.NoError(t, err)
			require.NotNil(t, result)

			var asset commons.Asset
			require.NoError(t, json.Unmarshal(result.Data(), &asset))

			require.Equal(t, tc.expectedAsset.Name, asset.Name)
			require.Equal(t, tc.expectedAsset.Type, asset.Type)
			require.Equal(t, tc.expectedAsset.Provider, asset.Provider)
			require.Equal(t, tc.expectedAsset.Location, asset.Location)
			require.Len(t, asset.Relationships, len(tc.expectedAsset.Relationships))
			for i, rel := range tc.expectedAsset.Relationships {
				require.Equal(t, rel, asset.Relationships[i])
			}
			require.Equal(t, tc.expectedAsset.Tags, asset.Tags)
		})
	}
}

func TestProcessWithImportAssetInventoryEvent(t *testing.T) {
	l, _ := test.NewNullLogger()

	testCases := []struct {
		name                 string
		importEvent          gcppubsubevents.InventoryImportEvent
		expectedAsset        commons.Asset
		mockRunServiceClient runservice.Client
		mockStorageClient    storageclient.Client
	}{
		{
			name: "import run service event",
			importEvent: gcppubsubevents.InventoryImportEvent{
				AssetName: "//run.googleapis.com/projects/test-project/locations/europe-west1/services/test-function",
				Type:      "run.googleapis.com/Service",
			},
			expectedAsset: commons.Asset{
				Name:     "test-function",
				Type:     "run.googleapis.com/Service",
				Provider: commons.GCPAssetProvider,
				Location: "europe-west1",
				Tags:     commons.Tags{"label1": "value1", "label2": "value2"},
			},
			mockRunServiceClient: &MockFnService{
				GetServiceResult: &runservice.Service{
					Name:   "projects/test-project/locations/europe-west1/services/test-function",
					Labels: map[string]string{"label1": "value1", "label2": "value2"},
				},
				GetServiceAssert: func(_ context.Context, name string) {
					require.Equal(t, "projects/test-project/locations/europe-west1/services/test-function", name)
				},
			},
		},
		{
			name: "import bucket event",
			importEvent: gcppubsubevents.InventoryImportEvent{
				AssetName: "//storage.googleapis.com/bucket1",
				Type:      "storage.googleapis.com/Bucket",
			},
			expectedAsset: commons.Asset{
				Name:     "bucket1",
				Type:     "storage.googleapis.com/Bucket",
				Provider: commons.GCPAssetProvider,
				Location: "europe-west1",
				Tags:     commons.Tags{"label1": "value1", "label2": "value2"},
			},
			mockStorageClient: &MockStorageClient{
				GetBucketResult: &storageclient.Bucket{
					Name:     "bucket1",
					Location: "europe-west1",
					Labels:   map[string]string{"label1": "value1", "label2": "value2"},
				},
				GetBucketAssert: func(_ context.Context, name string) {
					require.Equal(t, "//storage.googleapis.com/bucket1", name)
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := &GCPCloudVendorAggregator{
				logger:  l,
				options: nil,
				s:       tc.mockStorageClient,
				f:       tc.mockRunServiceClient,
			}

			data, err := json.Marshal(tc.importEvent)
			require.NoError(t, err)

			builder := gcppubsubevents.NewInventoryEventBuilder[gcppubsubevents.InventoryImportEvent]()
			event, err := builder.GetPipelineEvent(t.Context(), data)
			require.NoError(t, err)

			result, err := p.Process(event)
			require.NoError(t, err)
			require.NotNil(t, result)

			var asset commons.Asset
			require.NoError(t, json.Unmarshal(result.Data(), &asset))

			require.Equal(t, tc.expectedAsset.Name, asset.Name)
			require.Equal(t, tc.expectedAsset.Type, asset.Type)
			require.Equal(t, tc.expectedAsset.Provider, asset.Provider)
			require.Equal(t, tc.expectedAsset.Location, asset.Location)
			require.Len(t, asset.Relationships, len(tc.expectedAsset.Relationships))
			for i, rel := range tc.expectedAsset.Relationships {
				require.Equal(t, rel, asset.Relationships[i])
			}
			require.Equal(t, tc.expectedAsset.Tags, asset.Tags)
		})
	}
}
