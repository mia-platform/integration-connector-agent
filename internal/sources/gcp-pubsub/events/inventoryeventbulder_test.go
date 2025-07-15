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

package gcppubsubevents

import (
	"testing"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/stretchr/testify/require"
)

func TestGetPipelineEvent(t *testing.T) {
	testCases := []struct {
		name          string
		data          []byte
		expectedEvent *entities.Event
		expectedErr   error
	}{
		{
			name:        "invalid non-json data",
			data:        []byte("#### invalid data"),
			expectedErr: ErrMalformedEvent,
		},
		{
			name: "bucket creation event",
			data: []byte(bucketCreationEvent),
			expectedEvent: &entities.Event{
				PrimaryKeys: entities.PkFields{
					{Key: "resourceName", Value: "//storage.googleapis.com/testbucketname"},
					{Key: "resourceType", Value: "storage.googleapis.com/Bucket"},
				},
				Type:          RealtimeSyncEventType,
				OperationType: entities.Write,
				OriginalRaw:   []byte(bucketCreationEvent),
			},
		},
		{
			name: "bucket update event",
			data: []byte(bucketUpdateEvent),
			expectedEvent: &entities.Event{
				PrimaryKeys: entities.PkFields{
					{Key: "resourceName", Value: "//storage.googleapis.com/testbucketname"},
					{Key: "resourceType", Value: "storage.googleapis.com/Bucket"},
				},
				Type:          RealtimeSyncEventType,
				OperationType: entities.Write,
				OriginalRaw:   []byte(bucketUpdateEvent),
			},
		},
		{
			name: "bucket deletion event",
			data: []byte(bucketDeletionEvent),
			expectedEvent: &entities.Event{
				PrimaryKeys: entities.PkFields{
					{Key: "resourceName", Value: "//storage.googleapis.com/testbucketname"},
					{Key: "resourceType", Value: "storage.googleapis.com/Bucket"},
				},
				Type:          RealtimeSyncEventType,
				OperationType: entities.Delete,
				OriginalRaw:   []byte(bucketDeletionEvent),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			builder := NewInventoryEventBuilder[InventoryEvent]()

			event, err := builder.GetPipelineEvent(t.Context(), tc.data)

			if tc.expectedErr != nil {
				require.Error(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
			}

			if tc.expectedEvent == nil {
				require.Nil(t, event, "Expected event to be nil")
			} else {
				require.Equal(t, tc.expectedEvent, event, "Expected event does not match the returned event")
			}
		})
	}
}

const bucketCreationEvent = `{
	"asset": {
		"ancestors": [
			"projects/123123123",
			"folders/abcababc",
			"organizations/orgorgorg"
		],
		"assetType": "storage.googleapis.com/Bucket",
		"name": "//storage.googleapis.com/testbucketname",
		"resource": {},
		"updateTime": "2025-06-17T15:33:46.875574Z"
	},
	"priorAssetState": "DOES_NOT_EXIST",
	"window": {
		"startTime": "2025-06-17T15:33:46.875574Z"
	}
}`

const bucketUpdateEvent = `{
	"asset": {
		"ancestors": [
			"projects/20318464073",
			"folders/918177713511",
			"organizations/566099688680"
		],
		"assetType": "storage.googleapis.com/Bucket",
		"name": "//storage.googleapis.com/testbucketname",
		"resource": {},
		"updateTime": "2025-06-17T15:35:48.938746Z"
	},
	"priorAsset": {
		"ancestors": [
			"projects/20318464073",
			"folders/918177713511",
			"organizations/566099688680"
		],
		"assetType": "storage.googleapis.com/Bucket",
		"name": "//storage.googleapis.com/testbucketname",
		"resource": {},
		"updateTime": "2025-06-17T15:33:46.875574Z"
	},
	"priorAssetState": "PRESENT",
	"window": {
		"startTime": "2025-06-17T15:35:48.938746Z"
	}
}`

const bucketDeletionEvent = `{
	"asset": {
		"ancestors": [
			"projects/20318464073",
			"folders/918177713511",
			"organizations/566099688680"
		],
		"assetType": "storage.googleapis.com/Bucket",
		"name": "//storage.googleapis.com/testbucketname",
		"resource": {},
		"updateTime": "2025-06-17T15:36:25.534906Z"
	},
	"deleted": true,
	"priorAsset": {
		"ancestors": [
			"projects/20318464073",
			"folders/918177713511",
			"organizations/566099688680"
		],
		"assetType": "storage.googleapis.com/Bucket",
		"name": "//storage.googleapis.com/testbucketname",
		"resource": {},
		"updateTime": "2025-06-17T15:35:48.938746Z"
	},
	"priorAssetState": "PRESENT",
	"window": {
		"startTime": "2025-06-17T15:36:25.534906Z"
	}
}`
