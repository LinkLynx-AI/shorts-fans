package sqs

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	awssqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
)

type stubQueueAPI struct {
	sendInput     *awssqs.SendMessageInput
	sendErr       error
	receiveOutput *awssqs.ReceiveMessageOutput
	receiveErr    error
	deleteInput   *awssqs.DeleteMessageInput
	deleteInputs  []*awssqs.DeleteMessageInput
	deleteErr     error
}

func (s *stubQueueAPI) SendMessage(_ context.Context, params *awssqs.SendMessageInput, _ ...func(*awssqs.Options)) (*awssqs.SendMessageOutput, error) {
	s.sendInput = params
	return &awssqs.SendMessageOutput{}, s.sendErr
}

func (s *stubQueueAPI) ReceiveMessage(_ context.Context, params *awssqs.ReceiveMessageInput, _ ...func(*awssqs.Options)) (*awssqs.ReceiveMessageOutput, error) {
	if s.receiveOutput == nil {
		s.receiveOutput = &awssqs.ReceiveMessageOutput{}
	}
	return s.receiveOutput, s.receiveErr
}

func (s *stubQueueAPI) DeleteMessage(_ context.Context, params *awssqs.DeleteMessageInput, _ ...func(*awssqs.Options)) (*awssqs.DeleteMessageOutput, error) {
	s.deleteInput = params
	s.deleteInputs = append(s.deleteInputs, params)
	return &awssqs.DeleteMessageOutput{}, s.deleteErr
}

func TestQueuePublishMediaAssetID(t *testing.T) {
	t.Parallel()

	api := &stubQueueAPI{}
	queue, err := newQueue(api, "https://example.com/queue")
	if err != nil {
		t.Fatalf("newQueue() error = %v, want nil", err)
	}

	mediaAssetID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	if err := queue.PublishMediaAssetID(context.Background(), mediaAssetID); err != nil {
		t.Fatalf("PublishMediaAssetID() error = %v, want nil", err)
	}
	if got, want := aws.ToString(api.sendInput.QueueUrl), "https://example.com/queue"; got != want {
		t.Fatalf("PublishMediaAssetID() queue url got %q want %q", got, want)
	}
	if got, want := aws.ToString(api.sendInput.MessageBody), "{\"mediaAssetId\":\"11111111-1111-1111-1111-111111111111\"}"; got != want {
		t.Fatalf("PublishMediaAssetID() message body got %q want %q", got, want)
	}
}

func TestQueueReceiveWakeMessages(t *testing.T) {
	t.Parallel()

	api := &stubQueueAPI{
		receiveOutput: &awssqs.ReceiveMessageOutput{
			Messages: []awssqstypes.Message{
				{
					Body:          aws.String("{\"mediaAssetId\":\"11111111-1111-1111-1111-111111111111\"}"),
					ReceiptHandle: aws.String("receipt-1"),
				},
			},
		},
	}
	queue, err := newQueue(api, "https://example.com/queue")
	if err != nil {
		t.Fatalf("newQueue() error = %v, want nil", err)
	}

	messages, err := queue.ReceiveWakeMessages(context.Background())
	if err != nil {
		t.Fatalf("ReceiveWakeMessages() error = %v, want nil", err)
	}
	if len(messages) != 1 {
		t.Fatalf("ReceiveWakeMessages() len got %d want 1", len(messages))
	}
	if got, want := messages[0].ReceiptHandle, "receipt-1"; got != want {
		t.Fatalf("ReceiveWakeMessages() receipt handle got %q want %q", got, want)
	}
	if got, want := messages[0].MediaAssetID.String(), "11111111-1111-1111-1111-111111111111"; got != want {
		t.Fatalf("ReceiveWakeMessages() media asset id got %q want %q", got, want)
	}
}

func TestQueueReceiveWakeMessagesSkipsMalformedPayload(t *testing.T) {
	t.Parallel()

	api := &stubQueueAPI{
		receiveOutput: &awssqs.ReceiveMessageOutput{
			Messages: []awssqstypes.Message{
				{
					Body:          aws.String("not-json"),
					ReceiptHandle: aws.String("receipt-bad"),
				},
				{
					Body:          aws.String("{\"mediaAssetId\":\"11111111-1111-1111-1111-111111111111\"}"),
					ReceiptHandle: aws.String("receipt-good"),
				},
			},
		},
	}
	queue, err := newQueue(api, "https://example.com/queue")
	if err != nil {
		t.Fatalf("newQueue() error = %v, want nil", err)
	}

	messages, err := queue.ReceiveWakeMessages(context.Background())
	if err != nil {
		t.Fatalf("ReceiveWakeMessages() error = %v, want nil", err)
	}
	if len(messages) != 1 {
		t.Fatalf("ReceiveWakeMessages() len got %d want 1", len(messages))
	}
	if got, want := messages[0].ReceiptHandle, "receipt-good"; got != want {
		t.Fatalf("ReceiveWakeMessages() receipt handle got %q want %q", got, want)
	}
	if len(api.deleteInputs) != 1 {
		t.Fatalf("ReceiveWakeMessages() poison delete calls got %d want 1", len(api.deleteInputs))
	}
	if got, want := aws.ToString(api.deleteInputs[0].ReceiptHandle), "receipt-bad"; got != want {
		t.Fatalf("ReceiveWakeMessages() poison delete handle got %q want %q", got, want)
	}
}

func TestQueueReceiveWakeMessagesIgnoresPoisonDeleteFailure(t *testing.T) {
	t.Parallel()

	api := &stubQueueAPI{
		deleteErr: errors.New("delete failed"),
		receiveOutput: &awssqs.ReceiveMessageOutput{
			Messages: []awssqstypes.Message{
				{
					Body:          aws.String("not-json"),
					ReceiptHandle: aws.String("receipt-bad"),
				},
			},
		},
	}
	queue, err := newQueue(api, "https://example.com/queue")
	if err != nil {
		t.Fatalf("newQueue() error = %v, want nil", err)
	}

	messages, err := queue.ReceiveWakeMessages(context.Background())
	if err != nil {
		t.Fatalf("ReceiveWakeMessages() error = %v, want nil", err)
	}
	if len(messages) != 0 {
		t.Fatalf("ReceiveWakeMessages() len got %d want 0", len(messages))
	}
	if len(api.deleteInputs) != 1 {
		t.Fatalf("ReceiveWakeMessages() poison delete calls got %d want 1", len(api.deleteInputs))
	}
}

func TestQueueDeleteMessagePropagatesError(t *testing.T) {
	t.Parallel()

	deleteErr := errors.New("delete failed")
	queue, err := newQueue(&stubQueueAPI{deleteErr: deleteErr}, "https://example.com/queue")
	if err != nil {
		t.Fatalf("newQueue() error = %v, want nil", err)
	}

	if err := queue.DeleteMessage(context.Background(), "receipt-1"); !errors.Is(err, deleteErr) {
		t.Fatalf("DeleteMessage() error got %v want %v", err, deleteErr)
	}
}
