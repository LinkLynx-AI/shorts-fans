package s3

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
)

const defaultContentType = "application/octet-stream"

// Config は S3 client の最小設定を表します。
type Config struct {
	Region string
}

// Validate は S3 設定の整合性を検証します。
func (c Config) Validate() error {
	if strings.TrimSpace(c.Region) == "" {
		return fmt.Errorf("region is required")
	}

	return nil
}

type objectAPI interface {
	PutObject(ctx context.Context, params *awss3.PutObjectInput, optFns ...func(*awss3.Options)) (*awss3.PutObjectOutput, error)
	DeleteObject(ctx context.Context, params *awss3.DeleteObjectInput, optFns ...func(*awss3.Options)) (*awss3.DeleteObjectOutput, error)
}

type presignAPI interface {
	PresignGetObject(ctx context.Context, params *awss3.GetObjectInput, optFns ...func(*awss3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

// Client は media sandbox で使う最小限の S3 操作を包みます。
type Client struct {
	api       objectAPI
	presigner presignAPI
}

// NewClient は AWS SDK を使って S3 client を構築します。
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	rawClient := awss3.NewFromConfig(awsCfg)

	return newClient(rawClient, awss3.NewPresignClient(rawClient)), nil
}

func newClient(api objectAPI, presigner presignAPI) *Client {
	return &Client{
		api:       api,
		presigner: presigner,
	}
}

// PutObject は指定 bucket/key に object を配置します。
func (c *Client) PutObject(ctx context.Context, bucket string, key string, body []byte, contentType string) error {
	if c == nil {
		return fmt.Errorf("s3 client is nil")
	}
	if err := validateBucketAndKey(bucket, key); err != nil {
		return err
	}
	if contentType == "" {
		contentType = defaultContentType
	}

	_, err := c.api.PutObject(ctx, &awss3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(body),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("put object bucket=%s key=%s: %w", bucket, key, err)
	}

	return nil
}

// DeleteObject は指定 bucket/key の object を削除します。
func (c *Client) DeleteObject(ctx context.Context, bucket string, key string) error {
	if c == nil {
		return fmt.Errorf("s3 client is nil")
	}
	if err := validateBucketAndKey(bucket, key); err != nil {
		return err
	}

	_, err := c.api.DeleteObject(ctx, &awss3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete object bucket=%s key=%s: %w", bucket, key, err)
	}

	return nil
}

// PresignGetObject は private object 取得用の signed URL を生成します。
func (c *Client) PresignGetObject(ctx context.Context, bucket string, key string, expires time.Duration) (string, error) {
	if c == nil {
		return "", fmt.Errorf("s3 client is nil")
	}
	if err := validateBucketAndKey(bucket, key); err != nil {
		return "", err
	}
	if expires <= 0 {
		return "", fmt.Errorf("expires must be greater than zero")
	}

	request, err := c.presigner.PresignGetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, func(options *awss3.PresignOptions) {
		options.Expires = expires
	})
	if err != nil {
		return "", fmt.Errorf("presign get object bucket=%s key=%s: %w", bucket, key, err)
	}

	return request.URL, nil
}

func validateBucketAndKey(bucket string, key string) error {
	switch {
	case strings.TrimSpace(bucket) == "":
		return fmt.Errorf("bucket is required")
	case strings.TrimSpace(key) == "":
		return fmt.Errorf("key is required")
	default:
		return nil
	}
}
