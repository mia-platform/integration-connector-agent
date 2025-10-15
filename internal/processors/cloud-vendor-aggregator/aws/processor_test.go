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

package aws

import (
	"encoding/json"
	"testing"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/commons"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/config"
	awssqsevents "github.com/mia-platform/integration-connector-agent/internal/sources/aws-sqs/events"
	"github.com/sirupsen/logrus/hooks/test"

	"github.com/stretchr/testify/require"
)

func TestProcess(t *testing.T) {
	l, _ := test.NewNullLogger()
	p := New(l, config.AuthOptions{})

	e := &entities.Event{
		Type: awssqsevents.RealtimeSyncEventType,
	}
	e.WithData([]byte(`{
  "version": "0",
  "id": "77dd1b93-f3df-ac5c-ea6d-3b487ca64730",
  "detail-type": "AWS API Call via CloudTrail",
  "source": "aws.s3",
  "account": "accountid",
  "time": "2025-07-07T09:03:08Z",
  "region": "eu-north-1",
  "resources": [],
  "detail": {
    "eventVersion": "1.11",
    "userIdentity": {},
    "eventTime": "2025-07-07T09:03:08Z",
    "eventSource": "s3.amazonaws.com",
    "eventName": "PutBucketTagging",
    "awsRegion": "eu-north-1",
    "sourceIPAddress": "80.182.24.115",
    "userAgent": "[Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:139.0) Gecko/20100101 Firefox/139.0]",
    "requestParameters": {
      "Tagging": {
        "xmlns": "http://s3.amazonaws.com/doc/2006-03-01/",
        "TagSet": {
          "Tag": [
            {
              "Value": "tv2",
              "Key": "t1"
            },
            {
              "Value": "tv3",
              "Key": "t2"
            },
            {
              "Value": "tv4",
              "Key": "t3"
            }
          ]
        }
      },
      "tagging": "",
      "bucketName": "thebucketname",
      "Host": "s3.eu-north-1.amazonaws.com"
    },
    "responseElements": null,
    "additionalEventData": {},
    "requestID": "9PKQ42N01GSQCW64",
    "eventID": "681257c9-8cf7-4aac-8cba-143e127410e4",
    "readOnly": false,
    "resources": [{"accountId": "accountid","type": "AWS::S3::Bucket","ARN": "arn:aws:s3:::thebucketname"}],
    "eventType": "AwsApiCall",
    "managementEvent": true,
    "recipientAccountId": "accountid",
    "eventCategory": "Management",
    "tlsDetails": {"tlsVersion": "TLSv1.3","cipherSuite": "TLS_AES_128_GCM_SHA256","clientProvidedHostHeader": "s3.eu-north-1.amazonaws.com"}
  }
}`))

	res, err := p.Process(e)
	require.NoError(t, err, "Process should not return an error")
	require.NotNil(t, res, "Process should return a non-nil result")

	var asset commons.Asset
	require.NoError(t, json.Unmarshal(res.Data(), &asset))

	require.Equal(t, "thebucketname", asset.Name, "Processed data should match the input data")
	require.Equal(t, "s3.amazonaws.com", asset.Type, "Processed data type should be s3.amazonaws.com")
	require.Equal(t, commons.AWSAssetProvider, asset.Provider, "Processed data provider should be aws")
	require.Equal(t, "eu-north-1", asset.Location, "Processed data location should be eu-north-1")
	require.Len(t, asset.Relationships, 1, "Processed data should have one relationship")
	require.Equal(t, "account/accountid", asset.Relationships[0], "Processed data relationship should match the account ID")
	require.Empty(t, asset.Tags, "Processed data tags should be empty")
}
