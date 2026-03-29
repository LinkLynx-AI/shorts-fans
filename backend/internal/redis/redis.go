package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

// ReadinessChecker validates Redis readiness through a redis client.
type ReadinessChecker struct {
	client *goredis.Client
}

// NewClient constructs and validates a Redis client.
func NewClient(ctx context.Context, addr string) (*goredis.Client, error) {
	client := goredis.NewClient(&goredis.Options{
		Addr: addr,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		if closeErr := client.Close(); closeErr != nil {
			return nil, fmt.Errorf("ping redis: %w (close: %v)", err, closeErr)
		}

		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}

// NewReadinessChecker wraps a Redis client for readiness checks.
func NewReadinessChecker(client *goredis.Client) ReadinessChecker {
	return ReadinessChecker{client: client}
}

// CheckReadiness verifies that Redis responds to ping.
func (c ReadinessChecker) CheckReadiness(ctx context.Context) error {
	if c.client == nil {
		return fmt.Errorf("redis client is nil")
	}

	return c.client.Ping(ctx).Err()
}
