// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package commons

import (
	"context"
)

type DataAdapter[T any] interface {
	GetData(ctx context.Context, event T) ([]byte, error)
}

type Closable interface {
	Close() error
}
