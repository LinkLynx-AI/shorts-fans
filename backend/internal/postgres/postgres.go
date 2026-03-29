package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ReadinessChecker は pgxpool を通じて PostgreSQL の readiness を検証します。
type ReadinessChecker struct {
	pool *pgxpool.Pool
}

// NewPool は pgxpool.Pool を初期化し、接続確認まで行います。
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

// NewReadinessChecker は readiness check 用に pool を包みます。
func NewReadinessChecker(pool *pgxpool.Pool) ReadinessChecker {
	return ReadinessChecker{pool: pool}
}

// CheckReadiness は PostgreSQL が ping に応答できるか検証します。
func (c ReadinessChecker) CheckReadiness(ctx context.Context) error {
	if c.pool == nil {
		return fmt.Errorf("postgres pool is nil")
	}

	return c.pool.Ping(ctx)
}
