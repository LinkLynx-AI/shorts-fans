package sqs

import (
	"context"
	"testing"
)

func TestNewClientDisabledReturnsNil(t *testing.T) {
	t.Parallel()

	client, err := NewClient(context.Background(), Config{})
	if err != nil {
		t.Fatalf("NewClient() error = %v, want nil", err)
	}
	if client != nil {
		t.Fatalf("NewClient() client got %v want nil", client)
	}
}

func TestNewClientRejectsInvalidConfig(t *testing.T) {
	t.Parallel()

	client, err := NewClient(context.Background(), Config{QueueURL: "https://example.com/queue"})
	if err == nil {
		t.Fatal("NewClient() error = nil, want validation error")
	}
	if client != nil {
		t.Fatalf("NewClient() client got %v want nil", client)
	}
}

func TestNewClientSuccess(t *testing.T) {
	t.Parallel()

	client, err := NewClient(context.Background(), Config{
		Region:   "ap-northeast-1",
		QueueURL: "https://example.com/queue",
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v, want nil", err)
	}
	if client == nil {
		t.Fatal("NewClient() client = nil, want non-nil")
	}
}
