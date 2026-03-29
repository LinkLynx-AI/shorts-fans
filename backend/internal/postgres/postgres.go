package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ReadinessChecker validates PostgreSQL readiness through pgxpool.
type ReadinessChecker struct {
	pool *pgxpool.Pool
}

// NewPool constructs and validates a pgxpool.Pool.
func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return pool, nil
}

// NewReadinessChecker wraps a pool for readiness checks.
func NewReadinessChecker(pool *pgxpool.Pool) ReadinessChecker {
	return ReadinessChecker{pool: pool}
}

// CheckReadiness verifies that PostgreSQL responds to ping.
func (c ReadinessChecker) CheckReadiness(ctx context.Context) error {
	if c.pool == nil {
		return fmt.Errorf("postgres pool is nil")
	}

	return c.pool.Ping(ctx)
}
