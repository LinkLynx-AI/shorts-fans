package sqs

import (
	"context"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
)

// Config describes the minimal SQS client configuration.
type Config struct {
	Region   string
	QueueURL string
}

// Enabled reports whether SQS configuration is fully supplied.
func (c Config) Enabled() bool {
	return c.Region != "" && c.QueueURL != ""
}

// Validate checks that SQS configuration is internally consistent.
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

// NewClient builds an AWS SQS client when configuration is present.
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
