package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

// ReadinessChecker は Redis client を通じて Redis の readiness を検証します。
type ReadinessChecker struct {
	client *goredis.Client
}

// NewClient は Redis client を初期化し、接続確認まで行います。
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

// NewReadinessChecker は readiness check 用に Redis client を包みます。
func NewReadinessChecker(client *goredis.Client) ReadinessChecker {
	return ReadinessChecker{client: client}
}

// CheckReadiness は Redis が ping に応答できるか検証します。
func (c ReadinessChecker) CheckReadiness(ctx context.Context) error {
	if c.client == nil {
		return fmt.Errorf("redis client is nil")
	}

	return c.client.Ping(ctx).Err()
}
