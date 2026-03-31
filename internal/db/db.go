package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

var newPool = func(ctx context.Context, config *pgxpool.Config) (Pool, error) {
	return pgxpool.NewWithConfig(ctx, config)
}

type Pool interface {
	Ping(ctx context.Context) error
	Close()
}

type DB struct {
	pool *pgxpool.Pool
}

func (d *DB) Pool() *pgxpool.Pool {
	return d.pool
}

func Connect(ctx context.Context, databaseURL string) (*DB, error) {
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	pool, err := newPool(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{pool: pool.(*pgxpool.Pool)}, nil
}

func (d *DB) Close() {
	d.pool.Close()
}
