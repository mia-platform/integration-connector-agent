// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package gcpclient

import "context"

const (
	InventoryEventBucketPrefix   = "//storage.googleapis.com/"
	InventoryEventFunctionPrefix = "//run.googleapis.com/"
)

type ListenerFunc func(ctx context.Context, data []byte) error

type GCP interface {
	ListBuckets(ctx context.Context) ([]*Bucket, error)
	ListFunctions(ctx context.Context) ([]*Function, error)
	Listen(ctx context.Context, handler ListenerFunc) error
	Close() error
}

type Bucket struct {
	Name string
}

func (b *Bucket) AssetName() string {
	return InventoryEventBucketPrefix + b.Name
}

type Function struct {
	Name string `json:"name"`
}

func (f *Function) AssetName() string {
	return InventoryEventFunctionPrefix + f.Name
}
