package sqs

import (
	"context"
	"errors"
	"testing"

	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
)

type stubQueueAttributesAPI struct {
	output *awssqs.GetQueueAttributesOutput
	err    error
}

func (s stubQueueAttributesAPI) GetQueueAttributes(context.Context, *awssqs.GetQueueAttributesInput, ...func(*awssqs.Options)) (*awssqs.GetQueueAttributesOutput, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.output, nil
}

func TestAccessCheckerCheckAccess(t *testing.T) {
	t.Parallel()

	checker := newAccessChecker(stubQueueAttributesAPI{
		output: &awssqs.GetQueueAttributesOutput{
			Attributes: map[string]string{
				"QueueArn": "arn:aws:sqs:ap-northeast-1:123456789012:media-jobs",
			},
		},
	}, "https://example.com/queue")

	got, err := checker.CheckAccess(context.Background())
	if err != nil {
		t.Fatalf("CheckAccess() error = %v, want nil", err)
	}
	if got != "arn:aws:sqs:ap-northeast-1:123456789012:media-jobs" {
		t.Fatalf("CheckAccess() queue arn got %q want %q", got, "arn:aws:sqs:ap-northeast-1:123456789012:media-jobs")
	}
}

func TestAccessCheckerRejectsMissingQueueARN(t *testing.T) {
	t.Parallel()

	checker := newAccessChecker(stubQueueAttributesAPI{
		output: &awssqs.GetQueueAttributesOutput{},
	}, "https://example.com/queue")
	if _, err := checker.CheckAccess(context.Background()); err == nil {
		t.Fatal("CheckAccess() error = nil, want error")
	}
}

func TestAccessCheckerPropagatesErrors(t *testing.T) {
	t.Parallel()

	getErr := errors.New("get queue attributes failed")
	checker := newAccessChecker(stubQueueAttributesAPI{err: getErr}, "https://example.com/queue")
	if _, err := checker.CheckAccess(context.Background()); !errors.Is(err, getErr) {
		t.Fatalf("CheckAccess() error got %v want %v", err, getErr)
	}
}
