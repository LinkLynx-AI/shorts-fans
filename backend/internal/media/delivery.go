package media

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// DefaultSignedURLTTL は main private object の signed URL に使う既定の有効期限です。
const DefaultSignedURLTTL = 15 * time.Minute

// MainURLSigner は private main object の signed URL を生成します。
type MainURLSigner interface {
	PresignGetObject(ctx context.Context, bucket string, key string, expires time.Duration) (string, error)
}

// DeliveryConfig は short/main の delivery ref 解決に必要な設定です。
type DeliveryConfig struct {
	ShortPublicBaseURL    string
	MainPrivateBucketName string
}

// Delivery は media asset の delivery ref を解決します。
type Delivery struct {
	shortPublicBaseURL    string
	mainPrivateBucketName string
	signer                MainURLSigner
}

// NewDelivery は delivery helper を構築します。
func NewDelivery(cfg DeliveryConfig, signer MainURLSigner) (*Delivery, error) {
	if strings.TrimSpace(cfg.ShortPublicBaseURL) == "" {
		return nil, fmt.Errorf("short public base url is required")
	}
	if _, err := url.Parse(cfg.ShortPublicBaseURL); err != nil {
		return nil, fmt.Errorf("parse short public base url: %w", err)
	}
	if strings.TrimSpace(cfg.MainPrivateBucketName) == "" {
		return nil, fmt.Errorf("main private bucket name is required")
	}
	if signer == nil {
		return nil, fmt.Errorf("main url signer is required")
	}

	return &Delivery{
		shortPublicBaseURL:    strings.TrimSpace(cfg.ShortPublicBaseURL),
		mainPrivateBucketName: strings.TrimSpace(cfg.MainPrivateBucketName),
		signer:                signer,
	}, nil
}

// ShortPublicURL は short public object の再生 URL を返します。
func (d *Delivery) ShortPublicURL(storageKey string) (string, error) {
	if d == nil {
		return "", fmt.Errorf("delivery is nil")
	}

	return joinObjectURL(d.shortPublicBaseURL, storageKey)
}

// MainSignedURL は main private object の signed URL を返します。
func (d *Delivery) MainSignedURL(ctx context.Context, storageKey string, ttl time.Duration) (string, error) {
	if d == nil {
		return "", fmt.Errorf("delivery is nil")
	}
	if strings.TrimSpace(storageKey) == "" {
		return "", fmt.Errorf("storage key is required")
	}
	if ttl <= 0 {
		ttl = DefaultSignedURLTTL
	}

	signedURL, err := d.signer.PresignGetObject(ctx, d.mainPrivateBucketName, strings.TrimSpace(storageKey), ttl)
	if err != nil {
		return "", fmt.Errorf("generate main signed url key=%s: %w", storageKey, err)
	}

	return signedURL, nil
}

func joinObjectURL(baseURL string, storageKey string) (string, error) {
	trimmedKey := strings.Trim(strings.TrimSpace(storageKey), "/")
	if trimmedKey == "" {
		return "", fmt.Errorf("storage key is required")
	}

	parsedBaseURL, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return "", fmt.Errorf("parse base url: %w", err)
	}

	segments := strings.Split(trimmedKey, "/")
	cleanSegments := make([]string, 0, len(segments))
	for _, segment := range segments {
		if segment == "" {
			continue
		}

		cleanSegments = append(cleanSegments, segment)
	}
	if len(cleanSegments) == 0 {
		return "", fmt.Errorf("storage key is required")
	}

	parsedBaseURL.Path = strings.TrimRight(parsedBaseURL.Path, "/") + "/" + strings.Join(cleanSegments, "/")

	return parsedBaseURL.String(), nil
}
