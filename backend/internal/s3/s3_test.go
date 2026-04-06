package s3

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
)

type stubObjectAPI struct {
	putInput    *awss3.PutObjectInput
	putErr      error
	deleteInput *awss3.DeleteObjectInput
	deleteErr   error
}

func (s *stubObjectAPI) PutObject(_ context.Context, params *awss3.PutObjectInput, _ ...func(*awss3.Options)) (*awss3.PutObjectOutput, error) {
	s.putInput = params
	return &awss3.PutObjectOutput{}, s.putErr
}

func (s *stubObjectAPI) DeleteObject(_ context.Context, params *awss3.DeleteObjectInput, _ ...func(*awss3.Options)) (*awss3.DeleteObjectOutput, error) {
	s.deleteInput = params
	return &awss3.DeleteObjectOutput{}, s.deleteErr
}

type stubPresignAPI struct {
	input   *awss3.GetObjectInput
	expires time.Duration
	url     string
	err     error
}

func (s *stubPresignAPI) PresignGetObject(_ context.Context, params *awss3.GetObjectInput, optFns ...func(*awss3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
	s.input = params
	options := &awss3.PresignOptions{}
	for _, optFn := range optFns {
		optFn(options)
	}
	s.expires = options.Expires
	if s.err != nil {
		return nil, s.err
	}

	return &v4.PresignedHTTPRequest{URL: s.url}, nil
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

func TestPutObject(t *testing.T) {
	t.Parallel()

	api := &stubObjectAPI{}
	client := newClient(api, &stubPresignAPI{})

	if err := client.PutObject(context.Background(), "short-bucket", "probe/short.m3u8", []byte("body"), "application/vnd.apple.mpegurl"); err != nil {
		t.Fatalf("PutObject() error = %v, want nil", err)
	}

	if got, want := *api.putInput.Bucket, "short-bucket"; got != want {
		t.Fatalf("PutObject() bucket got %q want %q", got, want)
	}
	if got, want := *api.putInput.Key, "probe/short.m3u8"; got != want {
		t.Fatalf("PutObject() key got %q want %q", got, want)
	}
	if got, want := *api.putInput.ContentType, "application/vnd.apple.mpegurl"; got != want {
		t.Fatalf("PutObject() content type got %q want %q", got, want)
	}
}

func TestDeleteObject(t *testing.T) {
	t.Parallel()

	api := &stubObjectAPI{}
	client := newClient(api, &stubPresignAPI{})

	if err := client.DeleteObject(context.Background(), "main-bucket", "probe/main.m3u8"); err != nil {
		t.Fatalf("DeleteObject() error = %v, want nil", err)
	}

	if got, want := *api.deleteInput.Bucket, "main-bucket"; got != want {
		t.Fatalf("DeleteObject() bucket got %q want %q", got, want)
	}
	if got, want := *api.deleteInput.Key, "probe/main.m3u8"; got != want {
		t.Fatalf("DeleteObject() key got %q want %q", got, want)
	}
}

func TestPresignGetObject(t *testing.T) {
	t.Parallel()

	presigner := &stubPresignAPI{url: "https://signed.example.com/object"}
	client := newClient(&stubObjectAPI{}, presigner)

	got, err := client.PresignGetObject(context.Background(), "main-bucket", "probe/main.m3u8", 15*time.Minute)
	if err != nil {
		t.Fatalf("PresignGetObject() error = %v, want nil", err)
	}
	if got != "https://signed.example.com/object" {
		t.Fatalf("PresignGetObject() url got %q want %q", got, "https://signed.example.com/object")
	}
	if got, want := *presigner.input.Bucket, "main-bucket"; got != want {
		t.Fatalf("PresignGetObject() bucket got %q want %q", got, want)
	}
	if got, want := *presigner.input.Key, "probe/main.m3u8"; got != want {
		t.Fatalf("PresignGetObject() key got %q want %q", got, want)
	}
	if got, want := presigner.expires, 15*time.Minute; got != want {
		t.Fatalf("PresignGetObject() expires got %s want %s", got, want)
	}
}

func TestClientMethodsValidateInput(t *testing.T) {
	t.Parallel()

	client := newClient(&stubObjectAPI{}, &stubPresignAPI{})
	if err := client.PutObject(context.Background(), "", "key", nil, ""); err == nil {
		t.Fatal("PutObject() error = nil, want error")
	}
	if err := client.DeleteObject(context.Background(), "bucket", ""); err == nil {
		t.Fatal("DeleteObject() error = nil, want error")
	}
	if _, err := client.PresignGetObject(context.Background(), "bucket", "key", 0); err == nil {
		t.Fatal("PresignGetObject() error = nil, want error")
	}
}

func TestClientPropagatesErrors(t *testing.T) {
	t.Parallel()

	putErr := errors.New("put failed")
	deleteErr := errors.New("delete failed")
	presignErr := errors.New("presign failed")

	client := newClient(
		&stubObjectAPI{putErr: putErr, deleteErr: deleteErr},
		&stubPresignAPI{err: presignErr},
	)

	if err := client.PutObject(context.Background(), "bucket", "key", nil, ""); !errors.Is(err, putErr) {
		t.Fatalf("PutObject() error got %v want %v", err, putErr)
	}
	if err := client.DeleteObject(context.Background(), "bucket", "key"); !errors.Is(err, deleteErr) {
		t.Fatalf("DeleteObject() error got %v want %v", err, deleteErr)
	}
	if _, err := client.PresignGetObject(context.Background(), "bucket", "key", time.Minute); !errors.Is(err, presignErr) {
		t.Fatalf("PresignGetObject() error got %v want %v", err, presignErr)
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
	if got, want := reflect.TypeOf(client).String(), "*s3.Client"; got != want {
		t.Fatalf("NewClient() type got %q want %q", got, want)
	}
}
