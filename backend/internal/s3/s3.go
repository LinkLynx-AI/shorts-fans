package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
)

const defaultContentType = "application/octet-stream"

// PresignedUpload は direct upload 用の presigned request を表します。
type PresignedUpload struct {
	URL     string
	Headers map[string]string
}

// ObjectMetadata は object verification に必要な最小メタデータです。
type ObjectMetadata struct {
	ContentLength int64
	ContentType   string
}

// ObjectData は object body 読み出しに必要なデータを表します。
type ObjectData struct {
	Body        []byte
	ContentType string
}

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
	CopyObject(ctx context.Context, params *awss3.CopyObjectInput, optFns ...func(*awss3.Options)) (*awss3.CopyObjectOutput, error)
	DeleteObject(ctx context.Context, params *awss3.DeleteObjectInput, optFns ...func(*awss3.Options)) (*awss3.DeleteObjectOutput, error)
	GetObject(ctx context.Context, params *awss3.GetObjectInput, optFns ...func(*awss3.Options)) (*awss3.GetObjectOutput, error)
	HeadObject(ctx context.Context, params *awss3.HeadObjectInput, optFns ...func(*awss3.Options)) (*awss3.HeadObjectOutput, error)
}

type presignAPI interface {
	PresignGetObject(ctx context.Context, params *awss3.GetObjectInput, optFns ...func(*awss3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
	PresignPutObject(ctx context.Context, params *awss3.PutObjectInput, optFns ...func(*awss3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
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

// CopyObject は object を別 key へ複製します。
func (c *Client) CopyObject(ctx context.Context, sourceBucket string, sourceKey string, destinationBucket string, destinationKey string) error {
	if c == nil {
		return fmt.Errorf("s3 client is nil")
	}
	if err := validateBucketAndKey(sourceBucket, sourceKey); err != nil {
		return fmt.Errorf("source: %w", err)
	}
	if err := validateBucketAndKey(destinationBucket, destinationKey); err != nil {
		return fmt.Errorf("destination: %w", err)
	}

	copySource := strings.TrimLeft(strings.TrimSpace(sourceBucket)+"/"+strings.TrimLeft(strings.TrimSpace(sourceKey), "/"), "/")
	_, err := c.api.CopyObject(ctx, &awss3.CopyObjectInput{
		Bucket:     aws.String(destinationBucket),
		Key:        aws.String(destinationKey),
		CopySource: aws.String(copySource),
	})
	if err != nil {
		return fmt.Errorf(
			"copy object source_bucket=%s source_key=%s destination_bucket=%s destination_key=%s: %w",
			sourceBucket,
			sourceKey,
			destinationBucket,
			destinationKey,
			err,
		)
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

// PresignPutObject は direct upload 用の signed PUT request を生成します。
func (c *Client) PresignPutObject(ctx context.Context, bucket string, key string, contentType string, expires time.Duration) (PresignedUpload, error) {
	if c == nil {
		return PresignedUpload{}, fmt.Errorf("s3 client is nil")
	}
	if err := validateBucketAndKey(bucket, key); err != nil {
		return PresignedUpload{}, err
	}
	if strings.TrimSpace(contentType) == "" {
		return PresignedUpload{}, fmt.Errorf("content type is required")
	}
	if expires <= 0 {
		return PresignedUpload{}, fmt.Errorf("expires must be greater than zero")
	}

	request, err := c.presigner.PresignPutObject(ctx, &awss3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, func(options *awss3.PresignOptions) {
		options.Expires = expires
	})
	if err != nil {
		return PresignedUpload{}, fmt.Errorf("presign put object bucket=%s key=%s: %w", bucket, key, err)
	}

	return PresignedUpload{
		URL: request.URL,
		Headers: withRequiredContentType(
			normalizeSignedHeaders(request.SignedHeader),
			contentType,
		),
	}, nil
}

// HeadObject は object verification 用の metadata を返します。
func (c *Client) HeadObject(ctx context.Context, bucket string, key string) (ObjectMetadata, error) {
	if c == nil {
		return ObjectMetadata{}, fmt.Errorf("s3 client is nil")
	}
	if err := validateBucketAndKey(bucket, key); err != nil {
		return ObjectMetadata{}, err
	}

	output, err := c.api.HeadObject(ctx, &awss3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return ObjectMetadata{}, fmt.Errorf("head object bucket=%s key=%s: %w", bucket, key, err)
	}

	var contentLength int64
	if output.ContentLength != nil {
		contentLength = *output.ContentLength
	}

	return ObjectMetadata{
		ContentLength: contentLength,
		ContentType:   strings.TrimSpace(aws.ToString(output.ContentType)),
	}, nil
}

// GetObject は object body を読み出します。
func (c *Client) GetObject(ctx context.Context, bucket string, key string) (ObjectData, error) {
	if c == nil {
		return ObjectData{}, fmt.Errorf("s3 client is nil")
	}
	if err := validateBucketAndKey(bucket, key); err != nil {
		return ObjectData{}, err
	}

	output, err := c.api.GetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return ObjectData{}, fmt.Errorf("get object bucket=%s key=%s: %w", bucket, key, err)
	}
	defer func() {
		_ = output.Body.Close()
	}()

	body, err := io.ReadAll(output.Body)
	if err != nil {
		return ObjectData{}, fmt.Errorf("read object bucket=%s key=%s: %w", bucket, key, err)
	}

	return ObjectData{
		Body:        body,
		ContentType: strings.TrimSpace(aws.ToString(output.ContentType)),
	}, nil
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

func normalizeSignedHeaders(headers http.Header) map[string]string {
	if len(headers) == 0 {
		return map[string]string{}
	}

	normalized := make(map[string]string, len(headers))
	for key, values := range headers {
		if strings.EqualFold(key, "host") || len(values) == 0 {
			continue
		}

		normalized[http.CanonicalHeaderKey(key)] = strings.Join(values, ",")
	}

	return normalized
}

func withRequiredContentType(headers map[string]string, contentType string) map[string]string {
	if headers == nil {
		headers = make(map[string]string, 1)
	}
	if _, ok := headers["Content-Type"]; !ok {
		headers["Content-Type"] = contentType
	}

	return headers
}
