package media

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type stubObjectManager struct {
	putErr    error
	deleteErr error
	uploads   []probeUpload
	deletes   []probeUpload
}

type probeUpload struct {
	bucket string
	key    string
	body   []byte
}

func (s *stubObjectManager) PutObject(_ context.Context, bucket string, key string, body []byte, _ string) error {
	if s.putErr != nil {
		return s.putErr
	}

	s.uploads = append(s.uploads, probeUpload{bucket: bucket, key: key, body: append([]byte(nil), body...)})
	return nil
}

func (s *stubObjectManager) DeleteObject(_ context.Context, bucket string, key string) error {
	s.deletes = append(s.deletes, probeUpload{bucket: bucket, key: key})
	return s.deleteErr
}

type stubQueueChecker struct {
	arn string
	err error
}

func (s stubQueueChecker) CheckAccess(context.Context) (string, error) {
	if s.err != nil {
		return "", s.err
	}

	return s.arn, nil
}

type stubMediaConvertChecker struct {
	queueName string
	err       error
}

func (s stubMediaConvertChecker) CheckAccess(context.Context) (string, error) {
	if s.err != nil {
		return "", s.err
	}

	return s.queueName, nil
}

func TestProbeRunnerRun(t *testing.T) {
	t.Parallel()

	shortBody := []byte("#EXTM3U\n#EXT-X-VERSION:3\n# shorts-fans short probe\n")
	mainBody := []byte("#EXTM3U\n#EXT-X-VERSION:3\n# shorts-fans main probe\n")

	var mainSignedURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/signed/main-probe":
			_, _ = w.Write(mainBody)
		default:
			if strings.HasPrefix(r.URL.Path, "/public/probe-prefix/probe-id/") && strings.HasSuffix(r.URL.Path, "/short-probe.m3u8") {
				_, _ = w.Write(shortBody)
				return
			}

			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	delivery, err := NewDelivery(DeliveryConfig{
		ShortPublicBaseURL:    server.URL + "/public",
		MainPrivateBucketName: "main-bucket",
	}, &stubMainURLSigner{url: server.URL + "/signed/main-probe"})
	if err != nil {
		t.Fatalf("NewDelivery() error = %v, want nil", err)
	}

	objects := &stubObjectManager{}
	runner, err := NewProbeRunner(ProbeConfig{
		ShortPublicBucketName: "short-bucket",
		MainPrivateBucketName: "main-bucket",
		ProbePrefix:           "probe-prefix/probe-id",
		FetchAttempts:         1,
		FetchDelay:            time.Millisecond,
	}, delivery, objects, stubQueueChecker{
		arn: "arn:aws:sqs:ap-northeast-1:123456789012:media-jobs",
	}, stubMediaConvertChecker{
		queueName: "Default",
	}, server.Client())
	if err != nil {
		t.Fatalf("NewProbeRunner() error = %v, want nil", err)
	}

	result, err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	mainSignedURL = result.MainSignedURL

	if result.QueueARN != "arn:aws:sqs:ap-northeast-1:123456789012:media-jobs" {
		t.Fatalf("Run() queue arn got %q want %q", result.QueueARN, "arn:aws:sqs:ap-northeast-1:123456789012:media-jobs")
	}
	if result.MediaConvertQueueName != "Default" {
		t.Fatalf("Run() mediaconvert queue got %q want %q", result.MediaConvertQueueName, "Default")
	}
	if len(objects.uploads) != 2 {
		t.Fatalf("Run() upload count got %d want %d", len(objects.uploads), 2)
	}
	if len(objects.deletes) != 2 {
		t.Fatalf("Run() delete count got %d want %d", len(objects.deletes), 2)
	}
	if !strings.Contains(result.ShortPublicURL, "/public/probe-prefix/probe-id/") || !strings.HasSuffix(result.ShortPublicURL, "/short-probe.m3u8") {
		t.Fatalf("Run() short url got %q want probe prefix/suffix match", result.ShortPublicURL)
	}
	if mainSignedURL != server.URL+"/signed/main-probe" {
		t.Fatalf("Run() main signed url got %q want %q", mainSignedURL, server.URL+"/signed/main-probe")
	}
}

func TestProbeRunnerPropagatesQueueErrors(t *testing.T) {
	t.Parallel()

	delivery, err := NewDelivery(DeliveryConfig{
		ShortPublicBaseURL:    "https://cdn.example.com/public",
		MainPrivateBucketName: "main-bucket",
	}, &stubMainURLSigner{url: "https://signed.example.com/main"})
	if err != nil {
		t.Fatalf("NewDelivery() error = %v, want nil", err)
	}

	queueErr := errors.New("queue denied")
	runner, err := NewProbeRunner(ProbeConfig{
		ShortPublicBucketName: "short-bucket",
		MainPrivateBucketName: "main-bucket",
	}, delivery, &stubObjectManager{}, stubQueueChecker{err: queueErr}, stubMediaConvertChecker{queueName: "Default"}, nil)
	if err != nil {
		t.Fatalf("NewProbeRunner() error = %v, want nil", err)
	}

	if _, err := runner.Run(context.Background()); !errors.Is(err, queueErr) {
		t.Fatalf("Run() error got %v want %v", err, queueErr)
	}
}

func TestProbeRunnerCleansUpOnFetchFailure(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	delivery, err := NewDelivery(DeliveryConfig{
		ShortPublicBaseURL:    server.URL + "/public",
		MainPrivateBucketName: "main-bucket",
	}, &stubMainURLSigner{url: server.URL + "/signed/main"})
	if err != nil {
		t.Fatalf("NewDelivery() error = %v, want nil", err)
	}

	objects := &stubObjectManager{}
	runner, err := NewProbeRunner(ProbeConfig{
		ShortPublicBucketName: "short-bucket",
		MainPrivateBucketName: "main-bucket",
		ProbePrefix:           "probe-prefix/probe-id",
		FetchAttempts:         1,
	}, delivery, objects, stubQueueChecker{arn: "arn:aws:sqs:queue"}, stubMediaConvertChecker{queueName: "Default"}, server.Client())
	if err != nil {
		t.Fatalf("NewProbeRunner() error = %v, want nil", err)
	}

	if _, err := runner.Run(context.Background()); err == nil {
		t.Fatal("Run() error = nil, want error")
	}
	if len(objects.deletes) != 2 {
		t.Fatalf("Run() delete count got %d want %d", len(objects.deletes), 2)
	}
}
