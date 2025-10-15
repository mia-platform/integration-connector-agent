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

package runservice

import (
	"context"
	"fmt"

	run "cloud.google.com/go/run/apiv2"
	"cloud.google.com/go/run/apiv2/runpb"
	gcpOptions "google.golang.org/api/option"
)

type Service struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
}

type Client interface {
	GetService(ctx context.Context, name string) (*Service, error)
	Close() error
}

type gcpRunServiceClient struct {
	client *run.ServicesClient
}

func NewClient(ctx context.Context, options gcpOptions.ClientOption) (Client, error) {
	client, err := run.NewServicesClient(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCP run service client: %w", err)
	}
	return &gcpRunServiceClient{
		client: client,
	}, nil
}

func (c *gcpRunServiceClient) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

func (c *gcpRunServiceClient) GetService(ctx context.Context, name string) (*Service, error) {
	service, err := c.client.GetService(ctx, &runpb.GetServiceRequest{
		Name: name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get GCP run service: %w", err)
	}

	return &Service{
		Name:   service.GetName(),
		Labels: service.GetLabels(),
	}, nil
}
