package mediaconvert

import (
	"context"
	"fmt"
	"strings"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awsmc "github.com/aws/aws-sdk-go-v2/service/mediaconvert"
)

// Config は MediaConvert client の最小設定を表します。
type Config struct {
	Region string
}

// Validate は MediaConvert 設定の整合性を検証します。
func (c Config) Validate() error {
	if strings.TrimSpace(c.Region) == "" {
		return fmt.Errorf("region is required")
	}

	return nil
}

type queueLister interface {
	ListQueues(ctx context.Context, params *awsmc.ListQueuesInput, optFns ...func(*awsmc.Options)) (*awsmc.ListQueuesOutput, error)
}

// Client は MediaConvert への軽量 access check を包みます。
type Client struct {
	api queueLister
}

// NewClient は AWS SDK を使って MediaConvert client を構築します。
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	return newClient(awsmc.NewFromConfig(awsCfg)), nil
}

func newClient(api queueLister) *Client {
	return &Client{api: api}
}

// CheckAccess は queue 一覧取得で MediaConvert への接続前提を検証し、代表 queue 名を返します。
func (c *Client) CheckAccess(ctx context.Context) (string, error) {
	if c == nil {
		return "", fmt.Errorf("mediaconvert client is nil")
	}

	output, err := c.api.ListQueues(ctx, &awsmc.ListQueuesInput{})
	if err != nil {
		return "", fmt.Errorf("list mediaconvert queues: %w", err)
	}
	if len(output.Queues) == 0 {
		return "", fmt.Errorf("list mediaconvert queues returned no queues")
	}
	if output.Queues[0].Name == nil {
		return "", fmt.Errorf("mediaconvert queue name is empty")
	}

	queueName := strings.TrimSpace(*output.Queues[0].Name)
	if queueName == "" {
		return "", fmt.Errorf("mediaconvert queue name is empty")
	}

	return queueName, nil
}
