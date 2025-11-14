// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package awssqsevents

import (
	"os"
	"testing"

	"github.com/mia-platform/integration-connector-agent/entities"

	"github.com/stretchr/testify/require"
)

func TestCloudTrailEventBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		dataFilePath  string
		expectedEvent *entities.Event
		expectedErr   error
	}{
		{
			name:         "create bucket event",
			dataFilePath: "testdata/bucket-create.json",
			expectedEvent: &entities.Event{
				PrimaryKeys: entities.PkFields{
					{Key: "resourceName", Value: "the-bucket-name"},
					{Key: "eventSource", Value: "s3.amazonaws.com"},
				},
				Type:          RealtimeSyncEventType,
				OperationType: entities.Write,
			},
			expectedErr: nil,
		},
		{
			name:         "delete bucket event",
			dataFilePath: "testdata/bucket-delete.json",
			expectedEvent: &entities.Event{
				PrimaryKeys: entities.PkFields{
					{Key: "resourceName", Value: "the-bucket-name"},
					{Key: "eventSource", Value: "s3.amazonaws.com"},
				},
				Type:          RealtimeSyncEventType,
				OperationType: entities.Delete,
			},
			expectedErr: nil,
		},
		{
			name:         "update bucket tags",
			dataFilePath: "testdata/bucket-update-tags.json",
			expectedEvent: &entities.Event{
				PrimaryKeys: entities.PkFields{
					{Key: "resourceName", Value: "the-bucket-name"},
					{Key: "eventSource", Value: "s3.amazonaws.com"},
				},
				Type:          RealtimeSyncEventType,
				OperationType: entities.Write,
			},
			expectedErr: nil,
		},
		{
			name:         "update bucket ownership",
			dataFilePath: "testdata/bucket-update-ownership.json",
			expectedEvent: &entities.Event{
				PrimaryKeys: entities.PkFields{
					{Key: "resourceName", Value: "the-bucket-name"},
					{Key: "eventSource", Value: "s3.amazonaws.com"},
				},
				Type:          RealtimeSyncEventType,
				OperationType: entities.Write,
			},
			expectedErr: nil,
		},
		{
			name:         "create lambda function",
			dataFilePath: "testdata/lambda-create.json",
			expectedEvent: &entities.Event{
				PrimaryKeys: entities.PkFields{
					{Key: "resourceName", Value: "the-function-name"},
					{Key: "eventSource", Value: "lambda.amazonaws.com"},
				},
				Type:          RealtimeSyncEventType,
				OperationType: entities.Write,
			},
			expectedErr: nil,
		},
		{
			name:         "delete lambda function",
			dataFilePath: "testdata/lambda-delete.json",
			expectedEvent: &entities.Event{
				PrimaryKeys: entities.PkFields{
					{Key: "resourceName", Value: "the-function-name"},
					{Key: "eventSource", Value: "lambda.amazonaws.com"},
				},
				Type:          RealtimeSyncEventType,
				OperationType: entities.Delete,
			},
			expectedErr: nil,
		},
		{
			name:         "update lambda function code",
			dataFilePath: "testdata/lambda-update-code.json",
			expectedEvent: &entities.Event{
				PrimaryKeys: entities.PkFields{
					{Key: "resourceName", Value: "the-function-name"},
					{Key: "eventSource", Value: "lambda.amazonaws.com"},
				},
				Type:          RealtimeSyncEventType,
				OperationType: entities.Write,
			},
			expectedErr: nil,
		},
		{
			name:         "update lambda publish version",
			dataFilePath: "testdata/lambda-publish-version.json",
			expectedEvent: &entities.Event{
				PrimaryKeys: entities.PkFields{
					{Key: "resourceName", Value: "the-function-name"},
					{Key: "eventSource", Value: "lambda.amazonaws.com"},
				},
				Type:          RealtimeSyncEventType,
				OperationType: entities.Write,
			},
			expectedErr: nil,
		},
		{
			name:         "update lambda tags",
			dataFilePath: "testdata/lambda-update-tags.json",
			expectedEvent: &entities.Event{
				PrimaryKeys: entities.PkFields{
					{Key: "resourceName", Value: "the-function-name"},
					{Key: "eventSource", Value: "lambda.amazonaws.com"},
				},
				Type:          RealtimeSyncEventType,
				OperationType: entities.Write,
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			builder := NewCloudTrailEventBuilder[*CloudTrailEvent]()

			data, err := os.ReadFile(tc.dataFilePath)
			require.NoError(t, err, "Failed to read data file: %s", tc.dataFilePath)

			event, err := builder.GetPipelineEvent(t.Context(), data)

			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
			}

			if tc.expectedEvent == nil {
				require.Nil(t, event, "Expected event to be nil")
			} else {
				// patch expected event original raw data
				tc.expectedEvent.OriginalRaw = data

				require.Equal(t, tc.expectedEvent, event, "Expected event does not match the returned event")
			}
		})
	}
}
