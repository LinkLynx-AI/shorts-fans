package s3

import (
	"context"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
)

type stubObjectAPI struct {
	putInput    *awss3.PutObjectInput
	putErr      error
	copyInput   *awss3.CopyObjectInput
	copyErr     error
	deleteInput *awss3.DeleteObjectInput
	deleteErr   error
	getInput    *awss3.GetObjectInput
	getOutput   *awss3.GetObjectOutput
	getErr      error
	headInput   *awss3.HeadObjectInput
	headOutput  *awss3.HeadObjectOutput
	headErr     error
}

func (s *stubObjectAPI) PutObject(_ context.Context, params *awss3.PutObjectInput, _ ...func(*awss3.Options)) (*awss3.PutObjectOutput, error) {
	s.putInput = params
	return &awss3.PutObjectOutput{}, s.putErr
}

func (s *stubObjectAPI) CopyObject(_ context.Context, params *awss3.CopyObjectInput, _ ...func(*awss3.Options)) (*awss3.CopyObjectOutput, error) {
	s.copyInput = params
	return &awss3.CopyObjectOutput{}, s.copyErr
}

func (s *stubObjectAPI) DeleteObject(_ context.Context, params *awss3.DeleteObjectInput, _ ...func(*awss3.Options)) (*awss3.DeleteObjectOutput, error) {
	s.deleteInput = params
	return &awss3.DeleteObjectOutput{}, s.deleteErr
}

func (s *stubObjectAPI) GetObject(_ context.Context, params *awss3.GetObjectInput, _ ...func(*awss3.Options)) (*awss3.GetObjectOutput, error) {
	s.getInput = params
	if s.getOutput == nil {
		s.getOutput = &awss3.GetObjectOutput{
			Body: io.NopCloser(strings.NewReader("")),
		}
	}
	return s.getOutput, s.getErr
}

func (s *stubObjectAPI) HeadObject(_ context.Context, params *awss3.HeadObjectInput, _ ...func(*awss3.Options)) (*awss3.HeadObjectOutput, error) {
	s.headInput = params
	if s.headOutput == nil {
		s.headOutput = &awss3.HeadObjectOutput{}
	}
	return s.headOutput, s.headErr
}

type stubPresignAPI struct {
	getInput *awss3.GetObjectInput
	putInput *awss3.PutObjectInput
	expires  time.Duration
	url      string
	headers  http.Header
	err      error
}

func (s *stubPresignAPI) PresignGetObject(_ context.Context, params *awss3.GetObjectInput, optFns ...func(*awss3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
	s.getInput = params
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

func (s *stubPresignAPI) PresignPutObject(_ context.Context, params *awss3.PutObjectInput, optFns ...func(*awss3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
	s.putInput = params
	options := &awss3.PresignOptions{}
	for _, optFn := range optFns {
		optFn(options)
	}
	s.expires = options.Expires
	if s.err != nil {
		return nil, s.err
	}

	return &v4.PresignedHTTPRequest{
		URL:          s.url,
		SignedHeader: s.headers,
	}, nil
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

func TestCopyObject(t *testing.T) {
	t.Parallel()

	api := &stubObjectAPI{}
	client := newClient(api, &stubPresignAPI{})

	if err := client.CopyObject(context.Background(), "main-bucket", "mains/source.jpg", "main-bucket", "mains/target.jpg"); err != nil {
		t.Fatalf("CopyObject() error = %v, want nil", err)
	}

	if got, want := aws.ToString(api.copyInput.Bucket), "main-bucket"; got != want {
		t.Fatalf("CopyObject() destination bucket got %q want %q", got, want)
	}
	if got, want := aws.ToString(api.copyInput.Key), "mains/target.jpg"; got != want {
		t.Fatalf("CopyObject() destination key got %q want %q", got, want)
	}
	if got, want := aws.ToString(api.copyInput.CopySource), "main-bucket/mains/source.jpg"; got != want {
		t.Fatalf("CopyObject() copy source got %q want %q", got, want)
	}
}

func TestCopyObjectEscapesCopySourceKey(t *testing.T) {
	t.Parallel()

	api := &stubObjectAPI{}
	client := newClient(api, &stubPresignAPI{})

	if err := client.CopyObject(context.Background(), "main-bucket", "mains/source poster+#1.jpg", "main-bucket", "mains/target.jpg"); err != nil {
		t.Fatalf("CopyObject() error = %v, want nil", err)
	}

	if got, want := aws.ToString(api.copyInput.CopySource), "main-bucket/mains/source%20poster+%231.jpg"; got != want {
		t.Fatalf("CopyObject() copy source got %q want %q", got, want)
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
	if got, want := *presigner.getInput.Bucket, "main-bucket"; got != want {
		t.Fatalf("PresignGetObject() bucket got %q want %q", got, want)
	}
	if got, want := *presigner.getInput.Key, "probe/main.m3u8"; got != want {
		t.Fatalf("PresignGetObject() key got %q want %q", got, want)
	}
	if got, want := presigner.expires, 15*time.Minute; got != want {
		t.Fatalf("PresignGetObject() expires got %s want %s", got, want)
	}
}

func TestPresignPutObject(t *testing.T) {
	t.Parallel()

	presigner := &stubPresignAPI{
		url: "https://signed.example.com/upload",
		headers: http.Header{
			"Content-Type": []string{"video/mp4"},
			"Host":         []string{"signed.example.com"},
		},
	}
	client := newClient(&stubObjectAPI{}, presigner)

	got, err := client.PresignPutObject(context.Background(), "raw-bucket", "creator-upload/key.mp4", "video/mp4", 15*time.Minute)
	if err != nil {
		t.Fatalf("PresignPutObject() error = %v, want nil", err)
	}
	if got.URL != "https://signed.example.com/upload" {
		t.Fatalf("PresignPutObject() url got %q want %q", got.URL, "https://signed.example.com/upload")
	}
	if got.Headers["Content-Type"] != "video/mp4" {
		t.Fatalf("PresignPutObject() headers got %#v want Content-Type", got.Headers)
	}
	if _, ok := got.Headers["Host"]; ok {
		t.Fatalf("PresignPutObject() headers got %#v want no Host header", got.Headers)
	}
	if got, want := aws.ToString(presigner.putInput.Bucket), "raw-bucket"; got != want {
		t.Fatalf("PresignPutObject() bucket got %q want %q", got, want)
	}
	if got, want := aws.ToString(presigner.putInput.Key), "creator-upload/key.mp4"; got != want {
		t.Fatalf("PresignPutObject() key got %q want %q", got, want)
	}
	if got, want := aws.ToString(presigner.putInput.ContentType), "video/mp4"; got != want {
		t.Fatalf("PresignPutObject() content type got %q want %q", got, want)
	}
	if got, want := presigner.expires, 15*time.Minute; got != want {
		t.Fatalf("PresignPutObject() expires got %s want %s", got, want)
	}
}

func TestPresignPutObjectAddsContentTypeHeaderWhenMissing(t *testing.T) {
	t.Parallel()

	client := newClient(&stubObjectAPI{}, &stubPresignAPI{
		url:     "https://signed.example.com/upload",
		headers: http.Header{},
	})

	got, err := client.PresignPutObject(context.Background(), "raw-bucket", "creator-upload/key.mp4", "video/mp4", 15*time.Minute)
	if err != nil {
		t.Fatalf("PresignPutObject() error = %v, want nil", err)
	}
	if got.Headers["Content-Type"] != "video/mp4" {
		t.Fatalf("PresignPutObject() headers got %#v want injected Content-Type", got.Headers)
	}
}

func TestHeadObject(t *testing.T) {
	t.Parallel()

	contentLength := int64(42)
	client := newClient(&stubObjectAPI{
		headOutput: &awss3.HeadObjectOutput{
			ContentLength: &contentLength,
			ContentType:   aws.String("video/mp4"),
		},
	}, &stubPresignAPI{})

	got, err := client.HeadObject(context.Background(), "raw-bucket", "creator-upload/key.mp4")
	if err != nil {
		t.Fatalf("HeadObject() error = %v, want nil", err)
	}
	if got.ContentLength != 42 {
		t.Fatalf("HeadObject() content length got %d want %d", got.ContentLength, 42)
	}
	if got.ContentType != "video/mp4" {
		t.Fatalf("HeadObject() content type got %q want %q", got.ContentType, "video/mp4")
	}
}

func TestGetObject(t *testing.T) {
	t.Parallel()

	client := newClient(&stubObjectAPI{
		getOutput: &awss3.GetObjectOutput{
			Body:        io.NopCloser(strings.NewReader("avatar-bytes")),
			ContentType: aws.String("image/png"),
		},
	}, &stubPresignAPI{})

	got, err := client.GetObject(context.Background(), "avatar-bucket", "creator-avatar/key.png")
	if err != nil {
		t.Fatalf("GetObject() error = %v, want nil", err)
	}
	if string(got.Body) != "avatar-bytes" {
		t.Fatalf("GetObject() body got %q want %q", string(got.Body), "avatar-bytes")
	}
	if got.ContentType != "image/png" {
		t.Fatalf("GetObject() content type got %q want %q", got.ContentType, "image/png")
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
	if _, err := client.PresignPutObject(context.Background(), "bucket", "key", "", time.Minute); err == nil {
		t.Fatal("PresignPutObject() error = nil, want error")
	}
	if _, err := client.PresignGetObject(context.Background(), "bucket", "key", 0); err == nil {
		t.Fatal("PresignGetObject() error = nil, want error")
	}
	if _, err := client.HeadObject(context.Background(), "", "key"); err == nil {
		t.Fatal("HeadObject() error = nil, want error")
	}
	if _, err := client.GetObject(context.Background(), "", "key"); err == nil {
		t.Fatal("GetObject() error = nil, want error")
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
	if _, err := client.PresignPutObject(context.Background(), "bucket", "key", "video/mp4", time.Minute); !errors.Is(err, presignErr) {
		t.Fatalf("PresignPutObject() error got %v want %v", err, presignErr)
	}
	if _, err := client.PresignGetObject(context.Background(), "bucket", "key", time.Minute); !errors.Is(err, presignErr) {
		t.Fatalf("PresignGetObject() error got %v want %v", err, presignErr)
	}
	if _, err := newClient(&stubObjectAPI{headErr: errors.New("head failed")}, &stubPresignAPI{}).HeadObject(context.Background(), "bucket", "key"); err == nil {
		t.Fatal("HeadObject() error = nil, want error")
	}
	if _, err := newClient(&stubObjectAPI{getErr: errors.New("get failed")}, &stubPresignAPI{}).GetObject(context.Background(), "bucket", "key"); err == nil {
		t.Fatal("GetObject() error = nil, want error")
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
