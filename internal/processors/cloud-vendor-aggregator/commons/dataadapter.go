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
