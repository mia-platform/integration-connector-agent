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

package azuredevops

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/google/uuid"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/servicehooks"
	"github.com/stretchr/testify/assert"
)

func TestAlreadyExists(t *testing.T) {
	t.Parallel()

	testSubscriptions := &[]servicehooks.Subscription{
		{
			Id:          to.Ptr(uuid.MustParse("00000000-0000-0000-0000-000000000001")),
			EventType:   to.Ptr(repositoryCreated),
			PublisherId: to.Ptr("tfs"),
		},
		{
			Id:          to.Ptr(uuid.MustParse("00000000-0000-0000-0000-000000000002")),
			EventType:   to.Ptr(repositoryDeleted),
			PublisherId: to.Ptr("tfs"),
			ConsumerInputs: &map[string]string{
				"url": "",
			},
		},
		{
			Id:          to.Ptr(uuid.MustParse("00000000-0000-0000-0000-000000000003")),
			EventType:   to.Ptr(repositoryRenamed),
			PublisherId: to.Ptr("tfs"),
			ConsumerInputs: &map[string]string{
				"url": "",
			},
			PublisherInputs: &map[string]string{
				"projectId": "00000000-0000-0000-0000-00000000000",
			},
		},
	}
	testCases := map[string]struct {
		subscription   servicehooks.Subscription
		expectedResult *servicehooks.Subscription
	}{
		"complete subscription exists": {
			subscription: servicehooks.Subscription{
				EventType:   to.Ptr(repositoryRenamed),
				PublisherId: to.Ptr("tfs"),
				ConsumerInputs: &map[string]string{
					"url": "",
				},
				PublisherInputs: &map[string]string{
					"projectId": "00000000-0000-0000-0000-00000000000",
				},
			},
			expectedResult: &servicehooks.Subscription{
				Id:          to.Ptr(uuid.MustParse("00000000-0000-0000-0000-000000000003")),
				EventType:   to.Ptr(repositoryRenamed),
				PublisherId: to.Ptr("tfs"),
				ConsumerInputs: &map[string]string{
					"url": "",
				},
				PublisherInputs: &map[string]string{
					"projectId": "00000000-0000-0000-0000-00000000000",
				},
			},
		},
		"without publisher inputs exits": {
			subscription: servicehooks.Subscription{
				EventType:   to.Ptr(repositoryDeleted),
				PublisherId: to.Ptr("tfs"),
				ConsumerInputs: &map[string]string{
					"url": "",
				},
			},
			expectedResult: &servicehooks.Subscription{
				Id:          to.Ptr(uuid.MustParse("00000000-0000-0000-0000-000000000002")),
				EventType:   to.Ptr(repositoryDeleted),
				PublisherId: to.Ptr("tfs"),
				ConsumerInputs: &map[string]string{
					"url": "",
				},
			},
		},
		"without consumer & publisher inputs return nil": {
			subscription: servicehooks.Subscription{
				EventType:   to.Ptr(repositoryCreated),
				PublisherId: to.Ptr("tfs"),
				ConsumerInputs: &map[string]string{
					"url": "",
				},
			},
		},
		"with different event type return nil": {
			subscription: servicehooks.Subscription{
				EventType:   to.Ptr("git.pull.created"),
				PublisherId: to.Ptr("tfs"),
				ConsumerInputs: &map[string]string{
					"url": "",
				},
			},
		},
		"with different publisher id return nil": {
			subscription: servicehooks.Subscription{
				EventType:   to.Ptr("git.pull.deleted"),
				PublisherId: to.Ptr("test"),
				ConsumerInputs: &map[string]string{
					"url": "",
				},
			},
		},
		"with different url return nil": {
			subscription: servicehooks.Subscription{
				EventType:   to.Ptr(repositoryRenamed),
				PublisherId: to.Ptr("tfs"),
				ConsumerInputs: &map[string]string{
					"url": "http://example.com",
				},
				PublisherInputs: &map[string]string{
					"projectId": "00000000-0000-0000-0000-00000000000",
				},
			},
		},
		"with nil publisher input return nil": {
			subscription: servicehooks.Subscription{
				EventType:   to.Ptr(repositoryDeleted),
				PublisherId: to.Ptr("tfs"),
				ConsumerInputs: &map[string]string{
					"url": "",
				},
				PublisherInputs: &map[string]string{
					"projectId": "00000000-0000-0000-0000-00000000000",
				},
			},
		},
		"with different publisher input return nil": {
			subscription: servicehooks.Subscription{
				EventType:   to.Ptr(repositoryRenamed),
				PublisherId: to.Ptr("tfs"),
				ConsumerInputs: &map[string]string{
					"url": "",
				},
				PublisherInputs: &map[string]string{
					"projectId": "00000000-0000-0000-0000-00000000001",
				},
			},
		},
	}

	for testName, test := range testCases {
		t.Run(testName, func(t *testing.T) {
			existingSubscription := subscriptionAlreadyExists(test.subscription, testSubscriptions)
			assert.Equal(t, test.expectedResult, existingSubscription)
		})
	}
}
