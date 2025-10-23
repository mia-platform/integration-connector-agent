// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package gcpclient

import (
	"context"
	"errors"
	"fmt"
	"log"

	asset "cloud.google.com/go/asset/apiv1"
	"cloud.google.com/go/asset/apiv1/assetpb"
	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	run "cloud.google.com/go/run/apiv2"
	"cloud.google.com/go/run/apiv2/runpb"
	"cloud.google.com/go/storage"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type concrete struct {
	config GCPConfig
	log    *logrus.Logger

	p *pubsub.Client
	s *storage.Client
	f *run.ServicesClient
	a *asset.Client
}

type GCPConfig struct {
	ProjectID          string
	AckDeadlineSeconds int
	TopicName          string
	SubscriptionID     string
	CredentialsJSON    string
}

const (
	BucketAPI      = "storage.googleapis.com/Bucket"
	JobAPI         = "run.googleapis.com/Job"
	RevisionAPI    = "run.googleapis.com/Revision"
	ServiceAPI     = "run.googleapis.com/Service"
	InstanceAPI    = "compute.googleapis.com/Instance"
	DiskAPI        = "compute.googleapis.com/Disk"
	NetworkAPI     = "compute.googleapis.com/Network"
	FirewallAPI    = "compute.googleapis.com/Firewall"
	ClusterAPI     = "container.googleapis.com/Cluster"
	NodePoolAPI    = "container.googleapis.com/NodePool"
	SQLInstanceAPI = "sqladmin.googleapis.com/Instance"
	FolderAPI      = "cloudresourcemanager.googleapis.com/Folder"
)

var allAssetTypes = []string{
	BucketAPI,
	NetworkAPI,
}

func New(ctx context.Context, log *logrus.Logger, config GCPConfig) (GCP, error) {
	options := make([]option.ClientOption, 0)
	if config.CredentialsJSON != "" {
		log.Debug("using credentials JSON for Pub/Sub client")
		options = append(options, option.WithCredentialsJSON([]byte(config.CredentialsJSON)))
	}

	pubSubClient, err := pubsub.NewClient(ctx, config.ProjectID, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub client: %w", err)
	}

	storageClient, err := storage.NewClient(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}

	functionServiceClient, err := run.NewServicesClient(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCP run service client: %w", err)
	}

	assetClient, err := asset.NewClient(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCP asset client: %w", err)
	}
	fmt.Println("GCP clients created successfully")

	return &concrete{
		log:    log,
		config: config,
		p:      pubSubClient,
		s:      storageClient,
		f:      functionServiceClient,
		a:      assetClient,
	}, nil
}

func (p *concrete) Listen(ctx context.Context, handler ListenerFunc) error {
	subscriber := p.p.Subscriber(p.config.SubscriptionID)
	p.log.WithFields(logrus.Fields{
		"projectId":      p.config.ProjectID,
		"topicName":      p.config.TopicName,
		"subscriptionId": p.config.SubscriptionID,
	}).Debug("starting to listen to Pub/Sub messages")

	err := subscriber.Receive(ctx, handlerPubSubMessage(p, handler))
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound {
				// If the subscription does not exist, then create the subscription.
				subscription, err := p.p.SubscriptionAdminClient.CreateSubscription(ctx, &pubsubpb.Subscription{
					Name:  p.config.SubscriptionID,
					Topic: p.config.TopicName,
				})
				if err != nil {
					return err
				}

				subscriber = p.p.Subscriber(subscription.GetName())
				return subscriber.Receive(ctx, handlerPubSubMessage(p, handler))
			}
		}
	}

	return nil
}

func handlerPubSubMessage(p *concrete, handler ListenerFunc) func(ctx context.Context, msg *pubsub.Message) {
	return func(ctx context.Context, msg *pubsub.Message) {
		p.log.WithFields(logrus.Fields{
			"projectId":      p.config.ProjectID,
			"topicName":      p.config.TopicName,
			"subscriptionId": p.config.SubscriptionID,
			"messageId":      msg.ID,
		}).Trace("received message from Pub/Sub")

		if err := handler(ctx, msg.Data); err != nil {
			p.log.
				WithFields(logrus.Fields{
					"projectId":      p.config.ProjectID,
					"topicName":      p.config.TopicName,
					"subscriptionId": p.config.SubscriptionID,
					"messageId":      msg.ID,
				}).
				WithError(err).
				Error("error handling message")

			msg.Nack()
			return
		}

		// TODO: message is Acked here once the pipelines have received the message for processing.
		// This means that if the pipeline fails after this point, the message will not be
		// retried. Consider implementing, either:
		// - a dead-letter queue or similar mechanism.
		// - a way to be notified here if all the pipelins have processed the
		//   message successfully in order to correctly ack/nack it.
		msg.Ack()
	}
}

func (p *concrete) Close() error {
	if err := p.p.Close(); err != nil {
		return fmt.Errorf("failed to close pubsub client: %w", err)
	}

	if err := p.s.Close(); err != nil {
		return fmt.Errorf("failed to close storage client: %w", err)
	}
	return nil
}

func (p *concrete) ListAssets(ctx context.Context) ([]*assetpb.Asset, error) {
	req := &assetpb.ListAssetsRequest{
		Parent:      fmt.Sprintf("projects/%s", p.config.ProjectID),
		AssetTypes:  allAssetTypes,
		ContentType: assetpb.ContentType_RESOURCE,
	}

	it := p.a.ListAssets(ctx, req)

	assets := make([]*assetpb.Asset, 0)

	for {
		response, err := it.Next()
		if err == iterator.Done {
			fmt.Println("End of asset list. Import completed.")
			break
		}
		if err != nil {
			fmt.Println("MY FATAL ERROR")
			log.Fatal(err)
		}
		assets = append(assets, response)
	}
	return assets, nil
}

func (p *concrete) ListBuckets(ctx context.Context) ([]*Bucket, error) {
	buckets := make([]*Bucket, 0)

	it := p.s.Buckets(ctx, p.config.ProjectID)
	for {
		bucket, err := it.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return nil, fmt.Errorf("failed to list buckets: %w", err)
		}

		buckets = append(buckets, &Bucket{
			Name: bucket.Name,
		})
	}
	return buckets, nil
}

func (p *concrete) ListFunctions(ctx context.Context) ([]*Function, error) {
	functions := make([]*Function, 0)

	it := p.f.ListServices(ctx, &runpb.ListServicesRequest{
		Parent: fmt.Sprintf("projects/%s/locations/-", p.config.ProjectID),
	})
	for {
		function, err := it.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return nil, fmt.Errorf("failed to list functions: %w", err)
		}

		functions = append(functions, &Function{
			Name: function.GetName(),
		})
	}
	return functions, nil
}
