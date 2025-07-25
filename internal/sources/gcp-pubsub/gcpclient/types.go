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
