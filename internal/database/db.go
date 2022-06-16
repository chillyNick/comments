package database

import (
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/net/context"
)

func NewPgxPool(ctx context.Context, adr string) (*pgxpool.Pool, error) {
	return pgxpool.Connect(ctx, adr)
}
