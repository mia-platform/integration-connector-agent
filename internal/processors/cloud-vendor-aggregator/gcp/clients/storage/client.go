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
