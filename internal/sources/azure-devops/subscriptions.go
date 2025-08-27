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
	"context"
	"fmt"
	"net/url"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/forminput"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/servicehooks"
)

const (
	publisherID = "tfs"
)

var (
	supportedSubscriptions = []servicehooks.Subscription{
		{
			EventType:   to.Ptr(repositoryCreated),
			PublisherId: to.Ptr(publisherID),
		},
		{
			EventType:   to.Ptr(repositoryRenamed),
			PublisherId: to.Ptr(publisherID),
		},
		{
			EventType:   to.Ptr(repositoryDeleted),
			PublisherId: to.Ptr(publisherID),
		},
	}
)

func createSubscriptionsForProject(ctx context.Context, connection *azuredevops.Connection, devopsConfig *Config, projectID string) error {
	// get all subscription on project
	client := servicehooks.NewClient(ctx, connection)
	subscriptions, err := client.CreateSubscriptionsQuery(ctx, queryArgs(projectID))
	if err != nil {
		return fmt.Errorf("failed to get existing subscriptions for project %s: %w", projectID, err)
	}

	webhookURL, _ := url.Parse(devopsConfig.WebhookHost) // can skip error check, done in config validation
	webhookURL.Path = devopsConfig.WebhookPath

	subscriptionToDelete := make([]servicehooks.Subscription, 0)
	subscriptionToCreate := make([]servicehooks.Subscription, 0)
	for _, supportedSubscription := range supportedSubscriptions {
		subscription := newSubscription(supportedSubscription, projectID, webhookURL.String(), devopsConfig)

		existringSubscription := subscriptionAlreadyExists(subscription, subscriptions.Results)
		if existringSubscription == nil {
			subscriptionToCreate = append(subscriptionToCreate, subscription)
		} else if existringSubscription.Status != nil && *existringSubscription.Status != servicehooks.SubscriptionStatusValues.Enabled {
			subscriptionToCreate = append(subscriptionToCreate, subscription)
			subscriptionToDelete = append(subscriptionToDelete, *existringSubscription)
		}
	}

	if len(subscriptionToDelete) > 0 {
		for _, subscription := range subscriptionToDelete {
			// TODO: print error
			_ = client.DeleteSubscription(ctx, servicehooks.DeleteSubscriptionArgs{
				SubscriptionId: subscription.Id,
			})
		}
	}

	if len(subscriptionToCreate) > 0 {
		for _, subscription := range subscriptionToCreate {
			if _, err := client.CreateSubscription(ctx, servicehooks.CreateSubscriptionArgs{
				Subscription: &subscription,
			}); err != nil {
				return fmt.Errorf("failed to create subscription for project %s: %w", projectID, err)
			}
		}
	}

	return nil
}

func subscriptionAlreadyExists(subscription servicehooks.Subscription, subscriptions *[]servicehooks.Subscription) *servicehooks.Subscription {
	for _, existingSubscription := range *subscriptions {
		if *existingSubscription.EventType != *subscription.EventType ||
			*existingSubscription.PublisherId != *subscription.PublisherId ||
			existingSubscription.ConsumerInputs == nil {
			continue
		}

		if (*subscription.ConsumerInputs)["url"] != (*existingSubscription.ConsumerInputs)["url"] {
			continue
		}
		if existingSubscription.PublisherInputs == nil && subscription.PublisherInputs == nil {
			return &existingSubscription
		}

		if existingSubscription.PublisherInputs != nil && subscription.PublisherInputs != nil && (*subscription.PublisherInputs)["projectId"] == (*existingSubscription.PublisherInputs)["projectId"] {
			return &existingSubscription
		}
	}

	return nil
}

func queryArgs(projectID string) servicehooks.CreateSubscriptionsQueryArgs {
	query := servicehooks.CreateSubscriptionsQueryArgs{
		Query: &servicehooks.SubscriptionsQuery{
			PublisherId: to.Ptr(publisherID),
		},
	}

	if projectID != "" {
		query.Query.PublisherInputFilters = &[]forminput.InputFilter{
			{
				Conditions: &[]forminput.InputFilterCondition{
					{
						InputId:    to.Ptr("projectId"),
						Operator:   to.Ptr(forminput.InputFilterOperatorValues.Equals),
						InputValue: to.Ptr(projectID),
					},
				},
			},
		}
	}

	return query
}

func newSubscription(supportedSubscription servicehooks.Subscription, projectID string, webhookURL string, devopsConfig *Config) servicehooks.Subscription {
	subscription := servicehooks.Subscription{
		EventType:   supportedSubscription.EventType,
		PublisherId: supportedSubscription.PublisherId,
		ConsumerInputs: &map[string]string{
			"url": webhookURL,
		},
	}
	if projectID != "" {
		subscription.PublisherInputs = &map[string]string{
			"projectId": projectID,
		}
	}

	if devopsConfig.Authentication != nil && len(devopsConfig.Authentication.Username) > 0 {
		(*subscription.ConsumerInputs)["basicAuthUsername"] = devopsConfig.Authentication.Username
	}
	if devopsConfig.Authentication != nil && len(devopsConfig.Authentication.Secret.String()) > 0 {
		(*subscription.ConsumerInputs)["basicAuthPassword"] = devopsConfig.Authentication.Secret.String()
	}
	return subscription
}
