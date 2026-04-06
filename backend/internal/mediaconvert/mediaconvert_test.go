package mediaconvert

import (
	"context"
	"errors"
	"testing"

	awsmc "github.com/aws/aws-sdk-go-v2/service/mediaconvert"
	awsmctypes "github.com/aws/aws-sdk-go-v2/service/mediaconvert/types"
)

type stubQueueLister struct {
	output *awsmc.ListQueuesOutput
	err    error
}

func (s stubQueueLister) ListQueues(context.Context, *awsmc.ListQueuesInput, ...func(*awsmc.Options)) (*awsmc.ListQueuesOutput, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.output, nil
}

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	if err := (Config{Region: "ap-northeast-1"}).Validate(); err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}
	if err := (Config{}).Validate(); err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
}

func TestCheckAccess(t *testing.T) {
	t.Parallel()

	client := newClient(stubQueueLister{
		output: &awsmc.ListQueuesOutput{
			Queues: []awsmctypes.Queue{
				{Name: stringPtr("Default")},
			},
		},
	})

	got, err := client.CheckAccess(context.Background())
	if err != nil {
		t.Fatalf("CheckAccess() error = %v, want nil", err)
	}
	if got != "Default" {
		t.Fatalf("CheckAccess() queue name got %q want %q", got, "Default")
	}
}

func TestCheckAccessRejectsMissingQueues(t *testing.T) {
	t.Parallel()

	client := newClient(stubQueueLister{output: &awsmc.ListQueuesOutput{}})
	if _, err := client.CheckAccess(context.Background()); err == nil {
		t.Fatal("CheckAccess() error = nil, want error")
	}
}

func TestCheckAccessPropagatesErrors(t *testing.T) {
	t.Parallel()

	listErr := errors.New("list failed")
	client := newClient(stubQueueLister{err: listErr})
	if _, err := client.CheckAccess(context.Background()); !errors.Is(err, listErr) {
		t.Fatalf("CheckAccess() error got %v want %v", err, listErr)
	}
}

func TestNewClientSuccess(t *testing.T) {
	t.Parallel()

	client, err := NewClient(context.Background(), Config{Region: "ap-northeast-1"})
	if err != nil {
		t.Fatalf("NewClient() error = %v, want nil", err)
	}
	if client == nil {
		t.Fatal("NewClient() client = nil, want non-nil")
	}
}

func TestNewClientRejectsInvalidConfig(t *testing.T) {
	t.Parallel()

	client, err := NewClient(context.Background(), Config{})
	if err == nil {
		t.Fatal("NewClient() error = nil, want error")
	}
	if client != nil {
		t.Fatalf("NewClient() client got %#v want nil", client)
	}
}

func stringPtr(value string) *string {
	return &value
}
