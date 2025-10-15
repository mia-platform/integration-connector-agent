// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package storage

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/storage"
	gcpOptions "google.golang.org/api/option"
)

type Bucket struct {
	Name     string            `json:"name"`
	Location string            `json:"location"`
	Labels   map[string]string `json:"labels"`
}

type Client interface {
	GetBucket(ctx context.Context, name string) (*Bucket, error)
	Close() error
}

type gcpStorageClient struct {
	client *storage.Client
}

func NewClient(ctx context.Context, options gcpOptions.ClientOption) (Client, error) {
	client, err := storage.NewClient(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCP storage client: %w", err)
	}

	return &gcpStorageClient{
		client: client,
	}, nil
}

func (c *gcpStorageClient) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

func (c *gcpStorageClient) GetBucket(ctx context.Context, name string) (*Bucket, error) {
	attrs, err := c.client.Bucket(bucketName(name)).Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket attributes: %w", err)
	}

	return &Bucket{
		Name:     attrs.Name,
		Labels:   attrs.Labels,
		Location: attrs.Location,
	}, nil
}

func bucketName(name string) string {
	return strings.TrimPrefix(name, "//storage.googleapis.com/")
}
