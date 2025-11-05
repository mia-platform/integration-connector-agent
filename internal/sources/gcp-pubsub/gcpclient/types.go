// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package gcpclient

import (
	"context"

	"cloud.google.com/go/asset/apiv1/assetpb"
)

const (
	InventoryEventBucketPrefix   = "//storage.googleapis.com/"
	InventoryEventFunctionPrefix = "//run.googleapis.com/"
)

type ListenerFunc func(ctx context.Context, data []byte) error

type GCP interface {
	ListAssets(ctx context.Context) ([]*assetpb.Asset, error)
	Listen(ctx context.Context, handler ListenerFunc) error
	Close() error
}
