// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package sources

const (
	Jira                     = "jira"
	Console                  = "console"
	Github                   = "github"
	Gitlab                   = "gitlab"
	Confluence               = "confluence"
	GCPInventoryPubSub       = "gcp-inventory-pubsub"
	AzureActivityLogEventHub = "azure-activity-log-event-hub"
	AzureDevOps              = "azure-devops"
	AWSCloudTrailSQS         = "aws-cloudtrail-sqs"
	JBoss                    = "jboss"
)

type CloseableSource interface {
	Close() error
}
