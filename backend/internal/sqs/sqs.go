package sqs

import (
	"context"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
)

// Config は最小限の SQS client 設定を表します。
type Config struct {
	Region   string
	QueueURL string
}

// Enabled は SQS 設定が一式そろっているか返します。
func (c Config) Enabled() bool {
	return c.Region != "" && c.QueueURL != ""
}

// Validate は SQS 設定の整合性を検証します。
func (c Config) Validate() error {
	switch {
	case c.Region == "" && c.QueueURL == "":
		return nil
	case c.Region == "":
		return fmt.Errorf("region is required when queue url is set")
	case c.QueueURL == "":
		return fmt.Errorf("queue url is required when region is set")
	default:
		return nil
	}
}

// NewClient は設定がそろっている場合に AWS SQS client を構築します。
func NewClient(ctx context.Context, cfg Config) (*awssqs.Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if !cfg.Enabled() {
		return nil, nil
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	return awssqs.NewFromConfig(awsCfg), nil
}
