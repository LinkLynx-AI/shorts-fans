package sqs

import (
	"context"
	"fmt"
	"strings"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	awssqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// Config は最小限の SQS client 設定を表します。
type Config struct {
	Region   string
	QueueURL string
}

type queueAttributesAPI interface {
	GetQueueAttributes(ctx context.Context, params *awssqs.GetQueueAttributesInput, optFns ...func(*awssqs.Options)) (*awssqs.GetQueueAttributesOutput, error)
}

// AccessChecker は queue attribute を読んで access 前提を検証します。
type AccessChecker struct {
	client   queueAttributesAPI
	queueURL string
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

// NewAccessChecker は queue attribute 読み取り用の checker を構築します。
func NewAccessChecker(ctx context.Context, cfg Config) (*AccessChecker, error) {
	client, err := NewClient(ctx, cfg)
	if err != nil || client == nil {
		return nil, err
	}

	return newAccessChecker(client, cfg.QueueURL), nil
}

func newAccessChecker(client queueAttributesAPI, queueURL string) *AccessChecker {
	return &AccessChecker{
		client:   client,
		queueURL: queueURL,
	}
}

// CheckAccess は QueueArn を読めることを検証し、取得した ARN を返します。
func (c *AccessChecker) CheckAccess(ctx context.Context) (string, error) {
	if c == nil {
		return "", fmt.Errorf("sqs access checker is nil")
	}
	if strings.TrimSpace(c.queueURL) == "" {
		return "", fmt.Errorf("queue url is required")
	}

	output, err := c.client.GetQueueAttributes(ctx, &awssqs.GetQueueAttributesInput{
		QueueUrl: &c.queueURL,
		AttributeNames: []awssqstypes.QueueAttributeName{
			awssqstypes.QueueAttributeNameQueueArn,
		},
	})
	if err != nil {
		return "", fmt.Errorf("get queue attributes: %w", err)
	}

	queueARN := strings.TrimSpace(output.Attributes[string(awssqstypes.QueueAttributeNameQueueArn)])
	if queueARN == "" {
		return "", fmt.Errorf("queue arn attribute is missing")
	}

	return queueARN, nil
}
