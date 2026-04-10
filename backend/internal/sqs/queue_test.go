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
	sendInputs    []*awssqs.SendMessageInput
	sendErr       error
	sendFunc      func(*awssqs.SendMessageInput) error
	receiveOutput *awssqs.ReceiveMessageOutput
	receiveErr    error
	deleteInput   *awssqs.DeleteMessageInput
	deleteInputs  []*awssqs.DeleteMessageInput
	deleteErr     error
}

func (s *stubQueueAPI) SendMessage(_ context.Context, params *awssqs.SendMessageInput, _ ...func(*awssqs.Options)) (*awssqs.SendMessageOutput, error) {
	s.sendInput = params
	s.sendInputs = append(s.sendInputs, params)
	if s.sendFunc != nil {
		return &awssqs.SendMessageOutput{}, s.sendFunc(params)
	}
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

func TestNewQueueValidatesAndTrimsQueueURL(t *testing.T) {
	t.Parallel()

	if _, err := NewQueue(nil, "https://example.com/queue"); err == nil {
		t.Fatal("NewQueue() error = nil, want client validation error")
	}
	if _, err := newQueue(&stubQueueAPI{}, " "); err == nil {
		t.Fatal("newQueue() error = nil, want queue url validation error")
	}

	queue, err := NewQueue(&awssqs.Client{}, " https://example.com/queue ")
	if err != nil {
		t.Fatalf("NewQueue() error = %v, want nil", err)
	}
	if got, want := queue.queueURL, "https://example.com/queue"; got != want {
		t.Fatalf("NewQueue() queue url got %q want %q", got, want)
	}
}

func TestQueuePublishMediaAssetIDsPublishesEachMessage(t *testing.T) {
	t.Parallel()

	api := &stubQueueAPI{}
	queue, err := newQueue(api, "https://example.com/queue")
	if err != nil {
		t.Fatalf("newQueue() error = %v, want nil", err)
	}

	ids := []uuid.UUID{
		uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		uuid.MustParse("22222222-2222-2222-2222-222222222222"),
	}
	if err := queue.PublishMediaAssetIDs(context.Background(), ids); err != nil {
		t.Fatalf("PublishMediaAssetIDs() error = %v, want nil", err)
	}
	if len(api.sendInputs) != len(ids) {
		t.Fatalf("PublishMediaAssetIDs() send calls got %d want %d", len(api.sendInputs), len(ids))
	}
	if got, want := aws.ToString(api.sendInputs[1].MessageBody), "{\"mediaAssetId\":\"22222222-2222-2222-2222-222222222222\"}"; got != want {
		t.Fatalf("PublishMediaAssetIDs() second message body got %q want %q", got, want)
	}
}

func TestQueueNotifyProcessingQueuedDelegatesToPublish(t *testing.T) {
	t.Parallel()

	api := &stubQueueAPI{}
	queue, err := newQueue(api, "https://example.com/queue")
	if err != nil {
		t.Fatalf("newQueue() error = %v, want nil", err)
	}

	ids := []uuid.UUID{uuid.MustParse("11111111-1111-1111-1111-111111111111")}
	if err := queue.NotifyProcessingQueued(context.Background(), ids); err != nil {
		t.Fatalf("NotifyProcessingQueued() error = %v, want nil", err)
	}
	if len(api.sendInputs) != 1 {
		t.Fatalf("NotifyProcessingQueued() send calls got %d want 1", len(api.sendInputs))
	}
}

func TestQueuePublishMediaAssetIDsStopsOnError(t *testing.T) {
	t.Parallel()

	sendCalls := 0
	api := &stubQueueAPI{
		sendFunc: func(*awssqs.SendMessageInput) error {
			sendCalls++
			if sendCalls == 2 {
				return errors.New("send failed")
			}
			return nil
		},
	}
	queue, err := newQueue(api, "https://example.com/queue")
	if err != nil {
		t.Fatalf("newQueue() error = %v, want nil", err)
	}

	err = queue.PublishMediaAssetIDs(context.Background(), []uuid.UUID{
		uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		uuid.MustParse("33333333-3333-3333-3333-333333333333"),
	})
	if err == nil {
		t.Fatal("PublishMediaAssetIDs() error = nil, want send failure")
	}
	if sendCalls != 2 {
		t.Fatalf("PublishMediaAssetIDs() send calls got %d want 2", sendCalls)
	}
}

func TestQueuePublishMediaAssetIDValidatesInput(t *testing.T) {
	t.Parallel()

	var nilQueue *Queue
	if err := nilQueue.PublishMediaAssetID(context.Background(), uuid.MustParse("11111111-1111-1111-1111-111111111111")); err == nil {
		t.Fatal("PublishMediaAssetID() error = nil, want nil receiver error")
	}

	queue, err := newQueue(&stubQueueAPI{}, "https://example.com/queue")
	if err != nil {
		t.Fatalf("newQueue() error = %v, want nil", err)
	}

	if err := queue.PublishMediaAssetID(context.Background(), uuid.Nil); err == nil {
		t.Fatal("PublishMediaAssetID() error = nil, want media asset id validation error")
	}
}

func TestQueueReceiveWakeMessagesValidatesReceiverAndErrors(t *testing.T) {
	t.Parallel()

	var nilQueue *Queue
	if _, err := nilQueue.ReceiveWakeMessages(context.Background()); err == nil {
		t.Fatal("ReceiveWakeMessages() error = nil, want nil receiver error")
	}

	receiveErr := errors.New("receive failed")
	queue, err := newQueue(&stubQueueAPI{receiveErr: receiveErr}, "https://example.com/queue")
	if err != nil {
		t.Fatalf("newQueue() error = %v, want nil", err)
	}

	if _, err := queue.ReceiveWakeMessages(context.Background()); !errors.Is(err, receiveErr) {
		t.Fatalf("ReceiveWakeMessages() error got %v want %v", err, receiveErr)
	}
}

func TestQueueDeleteMessageValidatesInputAndSuccess(t *testing.T) {
	t.Parallel()

	var nilQueue *Queue
	if err := nilQueue.DeleteMessage(context.Background(), "receipt-1"); err == nil {
		t.Fatal("DeleteMessage() error = nil, want nil receiver error")
	}

	api := &stubQueueAPI{}
	queue, err := newQueue(api, "https://example.com/queue")
	if err != nil {
		t.Fatalf("newQueue() error = %v, want nil", err)
	}

	if err := queue.DeleteMessage(context.Background(), " "); err == nil {
		t.Fatal("DeleteMessage() error = nil, want receipt handle validation error")
	}
	if err := queue.DeleteMessage(context.Background(), " receipt-1 "); err != nil {
		t.Fatalf("DeleteMessage() error = %v, want nil", err)
	}
	if got, want := aws.ToString(api.deleteInput.ReceiptHandle), "receipt-1"; got != want {
		t.Fatalf("DeleteMessage() receipt handle got %q want %q", got, want)
	}
}

func TestParseReceivedWakeMessageValidatesPayload(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		message awssqstypes.Message
	}{
		{
			name:    "missing body",
			message: awssqstypes.Message{ReceiptHandle: aws.String("receipt-1")},
		},
		{
			name:    "blank body",
			message: awssqstypes.Message{Body: aws.String(" "), ReceiptHandle: aws.String("receipt-1")},
		},
		{
			name:    "missing receipt handle",
			message: awssqstypes.Message{Body: aws.String("{\"mediaAssetId\":\"11111111-1111-1111-1111-111111111111\"}")},
		},
		{
			name:    "invalid uuid",
			message: awssqstypes.Message{Body: aws.String("{\"mediaAssetId\":\"not-a-uuid\"}"), ReceiptHandle: aws.String("receipt-1")},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := parseReceivedWakeMessage(tt.message); err == nil {
				t.Fatal("parseReceivedWakeMessage() error = nil, want validation error")
			}
		})
	}
}
