package cache

import (
	"golang.org/x/net/context"
)

type Cache interface {
	SetComments(ctx context.Context, itemId int32, comments []byte) error
	GetComments(ctx context.Context, itemId int32) ([]byte, error)
	Close() error
}
