package media

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	defaultProbePrefix    = "codex/media-smoke"
	defaultFetchAttempts  = 5
	defaultFetchDelay     = time.Second
	defaultHTTPTimeout    = 10 * time.Second
	defaultCleanupTimeout = 10 * time.Second
)

// ProbeConfig は media smoke の実行設定です。
type ProbeConfig struct {
	ShortPublicBucketName string
	MainPrivateBucketName string
	ProbePrefix           string
	SignedURLTTL          time.Duration
	FetchAttempts         int
	FetchDelay            time.Duration
	HTTPTimeout           time.Duration
	CleanupTimeout        time.Duration
}

type probeObjectManager interface {
	PutObject(ctx context.Context, bucket string, key string, body []byte, contentType string) error
	DeleteObject(ctx context.Context, bucket string, key string) error
}

type queueAccessChecker interface {
	CheckAccess(ctx context.Context) (string, error)
}

type mediaConvertAccessChecker interface {
	CheckAccess(ctx context.Context) (string, error)
}

// ProbeResult は media smoke の成功結果を表します。
type ProbeResult struct {
	ShortObjectKey        string
	MainObjectKey         string
	ShortPublicURL        string
	MainSignedURL         string
	QueueARN              string
	MediaConvertQueueName string
}

// ProbeRunner は dev AWS media sandbox の representative path を検証します。
type ProbeRunner struct {
	config              ProbeConfig
	delivery            *Delivery
	objects             probeObjectManager
	queueChecker        queueAccessChecker
	mediaConvertChecker mediaConvertAccessChecker
	httpClient          *http.Client
}

// NewProbeRunner は media smoke runner を構築します。
func NewProbeRunner(
	cfg ProbeConfig,
	delivery *Delivery,
	objects probeObjectManager,
	queueChecker queueAccessChecker,
	mediaConvertChecker mediaConvertAccessChecker,
	httpClient *http.Client,
) (*ProbeRunner, error) {
	if delivery == nil {
		return nil, fmt.Errorf("delivery is required")
	}
	if objects == nil {
		return nil, fmt.Errorf("object manager is required")
	}
	if queueChecker == nil {
		return nil, fmt.Errorf("queue checker is required")
	}
	if mediaConvertChecker == nil {
		return nil, fmt.Errorf("mediaconvert checker is required")
	}
	if strings.TrimSpace(cfg.ShortPublicBucketName) == "" {
		return nil, fmt.Errorf("short public bucket name is required")
	}
	if strings.TrimSpace(cfg.MainPrivateBucketName) == "" {
		return nil, fmt.Errorf("main private bucket name is required")
	}
	if cfg.SignedURLTTL <= 0 {
		cfg.SignedURLTTL = DefaultSignedURLTTL
	}
	if cfg.FetchAttempts <= 0 {
		cfg.FetchAttempts = defaultFetchAttempts
	}
	if cfg.FetchDelay <= 0 {
		cfg.FetchDelay = defaultFetchDelay
	}
	if cfg.HTTPTimeout <= 0 {
		cfg.HTTPTimeout = defaultHTTPTimeout
	}
	if cfg.CleanupTimeout <= 0 {
		cfg.CleanupTimeout = defaultCleanupTimeout
	}
	if strings.TrimSpace(cfg.ProbePrefix) == "" {
		cfg.ProbePrefix = defaultProbePrefix
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: cfg.HTTPTimeout}
	}

	return &ProbeRunner{
		config:              cfg,
		delivery:            delivery,
		objects:             objects,
		queueChecker:        queueChecker,
		mediaConvertChecker: mediaConvertChecker,
		httpClient:          httpClient,
	}, nil
}

// Run は queue / MediaConvert / short public / main private の代表疎通を検証します。
func (r *ProbeRunner) Run(ctx context.Context) (result ProbeResult, runErr error) {
	queueARN, err := r.queueChecker.CheckAccess(ctx)
	if err != nil {
		return result, fmt.Errorf("check media queue access: %w", err)
	}
	result.QueueARN = queueARN

	queueName, err := r.mediaConvertChecker.CheckAccess(ctx)
	if err != nil {
		return result, fmt.Errorf("check mediaconvert access: %w", err)
	}
	result.MediaConvertQueueName = queueName

	probeID := uuid.NewString()
	shortKey := path.Join(r.config.ProbePrefix, probeID, "short-probe.m3u8")
	mainKey := path.Join(r.config.ProbePrefix, probeID, "main-probe.m3u8")
	result.ShortObjectKey = shortKey
	result.MainObjectKey = mainKey

	shortBody := []byte("#EXTM3U\n#EXT-X-VERSION:3\n# shorts-fans short probe\n")
	mainBody := []byte("#EXTM3U\n#EXT-X-VERSION:3\n# shorts-fans main probe\n")

	type uploadedObject struct {
		bucket string
		key    string
	}
	var uploaded []uploadedObject
	defer func() {
		if len(uploaded) == 0 {
			return
		}

		cleanupCtx, cancel := context.WithTimeout(context.Background(), r.config.CleanupTimeout)
		defer cancel()

		var cleanupErr error
		for _, object := range uploaded {
			err := r.objects.DeleteObject(cleanupCtx, object.bucket, object.key)
			if err != nil {
				cleanupErr = errors.Join(cleanupErr, fmt.Errorf("cleanup bucket=%s key=%s: %w", object.bucket, object.key, err))
			}
		}
		if cleanupErr != nil {
			runErr = errors.Join(runErr, cleanupErr)
		}
	}()

	if err := r.objects.PutObject(ctx, r.config.ShortPublicBucketName, shortKey, shortBody, "application/vnd.apple.mpegurl"); err != nil {
		return result, fmt.Errorf("upload short probe object: %w", err)
	}
	uploaded = append(uploaded, uploadedObject{bucket: r.config.ShortPublicBucketName, key: shortKey})

	if err := r.objects.PutObject(ctx, r.config.MainPrivateBucketName, mainKey, mainBody, "application/vnd.apple.mpegurl"); err != nil {
		return result, fmt.Errorf("upload main probe object: %w", err)
	}
	uploaded = append(uploaded, uploadedObject{bucket: r.config.MainPrivateBucketName, key: mainKey})

	result.ShortPublicURL, err = r.delivery.ShortPublicURL(shortKey)
	if err != nil {
		return result, fmt.Errorf("resolve short public url: %w", err)
	}
	if err := r.fetchAndMatch(ctx, result.ShortPublicURL, shortBody); err != nil {
		return result, fmt.Errorf("fetch short public probe: %w", err)
	}

	result.MainSignedURL, err = r.delivery.MainSignedURL(ctx, mainKey, r.config.SignedURLTTL)
	if err != nil {
		return result, fmt.Errorf("resolve main signed url: %w", err)
	}
	if err := r.fetchAndMatch(ctx, result.MainSignedURL, mainBody); err != nil {
		return result, fmt.Errorf("fetch main signed probe: %w", err)
	}

	return result, nil
}

func (r *ProbeRunner) fetchAndMatch(ctx context.Context, targetURL string, wantBody []byte) error {
	var lastErr error

	for attempt := 1; attempt <= r.config.FetchAttempts; attempt++ {
		lastErr = r.fetchOnce(ctx, targetURL, wantBody)
		if lastErr == nil {
			return nil
		}
		if attempt == r.config.FetchAttempts {
			break
		}

		timer := time.NewTimer(r.config.FetchDelay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("wait for retry url=%s: %w", redactURL(targetURL), ctx.Err())
		case <-timer.C:
		}
	}

	return lastErr
}

func (r *ProbeRunner) fetchOnce(ctx context.Context, targetURL string, wantBody []byte) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return fmt.Errorf("build get request url=%s: %w", redactURL(targetURL), err)
	}

	response, err := r.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("get url=%s: %w", redactURL(targetURL), err)
	}

	body, readErr := io.ReadAll(response.Body)
	closeErr := response.Body.Close()
	if readErr != nil {
		return fmt.Errorf("read response body url=%s: %w", redactURL(targetURL), readErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close response body url=%s: %w", redactURL(targetURL), closeErr)
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status url=%s: got %d want %d", redactURL(targetURL), response.StatusCode, http.StatusOK)
	}
	if !bytes.Equal(body, wantBody) {
		return fmt.Errorf("response body mismatch url=%s", redactURL(targetURL))
	}

	return nil
}

func redactURL(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	parsedURL.RawQuery = ""

	return parsedURL.String()
}
