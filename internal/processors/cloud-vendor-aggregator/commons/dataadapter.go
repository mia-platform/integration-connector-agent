// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

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
