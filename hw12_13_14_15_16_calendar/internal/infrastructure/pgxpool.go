package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const pingTimeout = 2 * time.Second

type PgxPool struct {
	*pgxpool.Pool
}

func NewPgxPool(ctx context.Context, dsn string) (*PgxPool, error) {
	conf, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.ParseConfig: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.NewWithConfig: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("pgxpool.Ping: %w", err)
	}

	return &PgxPool{pool}, nil
}
