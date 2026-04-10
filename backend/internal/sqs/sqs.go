package sqs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	awssqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
)

// Config は最小限の SQS client 設定を表します。
type Config struct {
	Region   string
	QueueURL string
}

const (
	defaultMaxMessages       = int32(10)
	defaultWaitTimeSeconds   = int32(10)
	defaultVisibilityTimeout = int32(30)
)

type queueAttributesAPI interface {
	GetQueueAttributes(ctx context.Context, params *awssqs.GetQueueAttributesInput, optFns ...func(*awssqs.Options)) (*awssqs.GetQueueAttributesOutput, error)
}

type queueAPI interface {
	SendMessage(ctx context.Context, params *awssqs.SendMessageInput, optFns ...func(*awssqs.Options)) (*awssqs.SendMessageOutput, error)
	ReceiveMessage(ctx context.Context, params *awssqs.ReceiveMessageInput, optFns ...func(*awssqs.Options)) (*awssqs.ReceiveMessageOutput, error)
	DeleteMessage(ctx context.Context, params *awssqs.DeleteMessageInput, optFns ...func(*awssqs.Options)) (*awssqs.DeleteMessageOutput, error)
}

// AccessChecker は queue attribute を読んで access 前提を検証します。
type AccessChecker struct {
	client   queueAttributesAPI
	queueURL string
}

type wakeMessage struct {
	MediaAssetID string `json:"mediaAssetId"`
}

// Queue は media job wake-up 用の最小 SQS 操作を包みます。
type Queue struct {
	client            queueAPI
	queueURL          string
	maxMessages       int32
	waitTimeSeconds   int32
	visibilityTimeout int32
}

// ReceivedWakeMessage は queue から取得した wake-up payload を表します。
type ReceivedWakeMessage struct {
	MediaAssetID  uuid.UUID
	ReceiptHandle string
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

// NewQueue は media processing wake-up 用 queue helper を構築します。
func NewQueue(client *awssqs.Client, queueURL string) (*Queue, error) {
	if client == nil {
		return nil, fmt.Errorf("sqs client is required")
	}

	return newQueue(client, queueURL)
}

func newQueue(client queueAPI, queueURL string) (*Queue, error) {
	if client == nil {
		return nil, fmt.Errorf("sqs queue client is required")
	}
	if strings.TrimSpace(queueURL) == "" {
		return nil, fmt.Errorf("queue url is required")
	}

	return &Queue{
		client:            client,
		queueURL:          strings.TrimSpace(queueURL),
		maxMessages:       defaultMaxMessages,
		waitTimeSeconds:   defaultWaitTimeSeconds,
		visibilityTimeout: defaultVisibilityTimeout,
	}, nil
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

// PublishMediaAssetIDs は media asset ごとの wake-up message を enqueue します。
func (q *Queue) PublishMediaAssetIDs(ctx context.Context, mediaAssetIDs []uuid.UUID) error {
	if q == nil {
		return fmt.Errorf("sqs queue is nil")
	}

	for _, mediaAssetID := range mediaAssetIDs {
		if err := q.PublishMediaAssetID(ctx, mediaAssetID); err != nil {
			return err
		}
	}

	return nil
}

// NotifyProcessingQueued は creatorupload.ProcessingNotifier 互換の best-effort notify を提供します。
func (q *Queue) NotifyProcessingQueued(ctx context.Context, mediaAssetIDs []uuid.UUID) error {
	return q.PublishMediaAssetIDs(ctx, mediaAssetIDs)
}

// PublishMediaAssetID は media asset 1 件分の wake-up message を enqueue します。
func (q *Queue) PublishMediaAssetID(ctx context.Context, mediaAssetID uuid.UUID) error {
	if q == nil {
		return fmt.Errorf("sqs queue is nil")
	}
	if mediaAssetID == uuid.Nil {
		return fmt.Errorf("media asset id is required")
	}

	body, err := json.Marshal(wakeMessage{MediaAssetID: mediaAssetID.String()})
	if err != nil {
		return fmt.Errorf("marshal wake message media_asset_id=%s: %w", mediaAssetID, err)
	}

	_, err = q.client.SendMessage(ctx, &awssqs.SendMessageInput{
		QueueUrl:    &q.queueURL,
		MessageBody: stringPtr(string(body)),
	})
	if err != nil {
		return fmt.Errorf("send wake message media_asset_id=%s: %w", mediaAssetID, err)
	}

	return nil
}

// ReceiveWakeMessages は queue から wake-up message を取得します。
func (q *Queue) ReceiveWakeMessages(ctx context.Context) ([]ReceivedWakeMessage, error) {
	if q == nil {
		return nil, fmt.Errorf("sqs queue is nil")
	}

	output, err := q.client.ReceiveMessage(ctx, &awssqs.ReceiveMessageInput{
		QueueUrl:            &q.queueURL,
		MaxNumberOfMessages: q.maxMessages,
		WaitTimeSeconds:     q.waitTimeSeconds,
		VisibilityTimeout:   q.visibilityTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("receive wake messages: %w", err)
	}

	messages := make([]ReceivedWakeMessage, 0, len(output.Messages))
	for _, raw := range output.Messages {
		parsed, parseErr := parseReceivedWakeMessage(raw)
		if parseErr != nil {
			q.discardMalformedWakeMessage(ctx, raw)
			continue
		}
		messages = append(messages, parsed)
	}

	return messages, nil
}

// DeleteMessage は処理済み wake-up message を queue から削除します。
func (q *Queue) DeleteMessage(ctx context.Context, receiptHandle string) error {
	if q == nil {
		return fmt.Errorf("sqs queue is nil")
	}
	if strings.TrimSpace(receiptHandle) == "" {
		return fmt.Errorf("receipt handle is required")
	}

	_, err := q.client.DeleteMessage(ctx, &awssqs.DeleteMessageInput{
		QueueUrl:      &q.queueURL,
		ReceiptHandle: stringPtr(strings.TrimSpace(receiptHandle)),
	})
	if err != nil {
		return fmt.Errorf("delete wake message: %w", err)
	}

	return nil
}

func parseReceivedWakeMessage(message awssqstypes.Message) (ReceivedWakeMessage, error) {
	if message.Body == nil || strings.TrimSpace(*message.Body) == "" {
		return ReceivedWakeMessage{}, fmt.Errorf("wake message body is empty")
	}
	if message.ReceiptHandle == nil || strings.TrimSpace(*message.ReceiptHandle) == "" {
		return ReceivedWakeMessage{}, fmt.Errorf("wake message receipt handle is empty")
	}

	var payload wakeMessage
	if err := json.Unmarshal([]byte(*message.Body), &payload); err != nil {
		return ReceivedWakeMessage{}, fmt.Errorf("unmarshal wake message: %w", err)
	}

	mediaAssetID, err := uuid.Parse(strings.TrimSpace(payload.MediaAssetID))
	if err != nil {
		return ReceivedWakeMessage{}, fmt.Errorf("parse wake message media asset id: %w", err)
	}

	return ReceivedWakeMessage{
		MediaAssetID:  mediaAssetID,
		ReceiptHandle: strings.TrimSpace(*message.ReceiptHandle),
	}, nil
}

func (q *Queue) discardMalformedWakeMessage(ctx context.Context, message awssqstypes.Message) {
	if q == nil || message.ReceiptHandle == nil {
		return
	}

	receiptHandle := strings.TrimSpace(*message.ReceiptHandle)
	if receiptHandle == "" {
		return
	}

	// SQS は wake-up 用なので、poison message は best-effort に除去して DB reconciliation を優先する。
	_ = q.DeleteMessage(ctx, receiptHandle)
}

func stringPtr(value string) *string {
	return &value
}
