package creatoravatar

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/smithy-go"
	"github.com/google/uuid"

	medias3 "github.com/LinkLynx-AI/shorts-fans/backend/internal/s3"
)

const (
	defaultUploadTTL    = 15 * time.Minute
	maxFileSizeBytes    = 5_242_880
	uploadStateCreated  = "created"
	uploadStateComplete = "completed"
	uploadStateConsumed = "consumed"

	avatarUploadTokenPrefix = "vcupl_"
	avatarAssetIDPrefix     = "asset_creator_registration_avatar_"
)

var (
	// ErrUploadExpired は avatar upload token の期限切れを表します。
	ErrUploadExpired = errors.New("creator avatar upload has expired")
	// ErrUploadIncomplete は avatar upload object が未完了または不正なことを表します。
	ErrUploadIncomplete = errors.New("creator avatar upload is incomplete")
	// ErrUploadNotFound は viewer が解決できない avatar upload token を表します。
	ErrUploadNotFound = errors.New("creator avatar upload was not found")
	// ErrUploadConsumed は successful registration 後に消費済みの token を表します。
	ErrUploadConsumed = errors.New("creator avatar upload was already consumed")
	// ErrStorageFailure は S3 / Redis などの dependency failure を表します。
	ErrStorageFailure = errors.New("creator avatar upload storage failure")
)

// ValidationError は request payload の validation failure を表します。
type ValidationError struct {
	code    string
	message string
}

// NewValidationError は transport layer に返す validation error を構築します。
func NewValidationError(code string, message string) *ValidationError {
	return &ValidationError{
		code:    code,
		message: message,
	}
}

func (e *ValidationError) Error() string {
	return e.message
}

// Code は transport error code を返します。
func (e *ValidationError) Code() string {
	return e.code
}

// Message は transport error message を返します。
func (e *ValidationError) Message() string {
	return e.message
}

// CreateUploadInput は avatar upload initiation request を表します。
type CreateUploadInput struct {
	FileName      string
	FileSizeBytes int64
	MimeType      string
	ViewerUserID  uuid.UUID
}

// DirectUpload は presigned PUT request を表します。
type DirectUpload struct {
	Headers map[string]string
	Method  string
	URL     string
}

// UploadTarget は client が直接 PUT する target を表します。
type UploadTarget struct {
	FileName string
	MimeType string
	Upload   DirectUpload
}

// CreateUploadResult は initiation response を表します。
type CreateUploadResult struct {
	AvatarUploadToken string
	ExpiresAt         time.Time
	UploadTarget      UploadTarget
}

// CompleteUploadInput は avatar upload completion request を表します。
type CompleteUploadInput struct {
	AvatarUploadToken string
	ViewerUserID      uuid.UUID
}

// CompletedUpload は completed avatar upload を表します。
type CompletedUpload struct {
	AvatarAssetID     string
	AvatarUploadToken string
	AvatarURL         string
}

// CompleteUploadResult は completion response を表します。
type CompleteUploadResult struct {
	Avatar CompletedUpload
}

// ServiceConfig は avatar upload service の実行設定を表します。
type ServiceConfig struct {
	DeliveryBaseURL    string
	DeliveryBucketName string
	UploadBucketName   string
	UploadTTL          time.Duration
}

type uploadStore interface {
	DeleteUpload(ctx context.Context, avatarUploadToken string) error
	GetUpload(ctx context.Context, avatarUploadToken string) (storedUpload, error)
	SaveUpload(ctx context.Context, avatarUploadToken string, upload storedUpload, ttl time.Duration) error
}

// Storage は avatar upload に必要な S3 操作を表します。
type Storage interface {
	GetObject(ctx context.Context, bucket string, key string) (medias3.ObjectData, error)
	HeadObject(ctx context.Context, bucket string, key string) (medias3.ObjectMetadata, error)
	PresignPutObject(ctx context.Context, bucket string, key string, contentType string, expires time.Duration) (medias3.PresignedUpload, error)
	PutObject(ctx context.Context, bucket string, key string, body []byte, contentType string) error
}

// Service は creator registration avatar upload を扱います。
type Service struct {
	deliveryBaseURL    string
	deliveryBucketName string
	now                func() time.Time
	newOpaqueID        func(prefix string) (string, error)
	store              uploadStore
	storage            Storage
	uploadBucketName   string
	uploadTTL          time.Duration
}

// NewService は creator avatar upload service を構築します。
func NewService(cfg ServiceConfig, storage Storage, store uploadStore) (*Service, error) {
	if storage == nil {
		return nil, fmt.Errorf("creator avatar storage is required")
	}
	if store == nil {
		return nil, fmt.Errorf("creator avatar upload store is required")
	}
	if strings.TrimSpace(cfg.UploadBucketName) == "" {
		return nil, fmt.Errorf("creator avatar upload bucket name is required")
	}
	if strings.TrimSpace(cfg.DeliveryBucketName) == "" {
		return nil, fmt.Errorf("creator avatar delivery bucket name is required")
	}
	if strings.TrimSpace(cfg.DeliveryBaseURL) == "" {
		return nil, fmt.Errorf("creator avatar delivery base url is required")
	}
	if cfg.UploadTTL <= 0 {
		cfg.UploadTTL = defaultUploadTTL
	}

	return &Service{
		deliveryBaseURL:    strings.TrimSpace(cfg.DeliveryBaseURL),
		deliveryBucketName: strings.TrimSpace(cfg.DeliveryBucketName),
		now:                time.Now,
		newOpaqueID:        generateOpaqueID,
		store:              store,
		storage:            storage,
		uploadBucketName:   strings.TrimSpace(cfg.UploadBucketName),
		uploadTTL:          cfg.UploadTTL,
	}, nil
}

// CreateUpload は avatar upload target を発行します。
func (s *Service) CreateUpload(ctx context.Context, input CreateUploadInput) (CreateUploadResult, error) {
	if s == nil {
		return CreateUploadResult{}, fmt.Errorf("creator avatar service is nil")
	}

	metadata, err := normalizeFileMetadata(input.FileName, input.MimeType, input.FileSizeBytes)
	if err != nil {
		return CreateUploadResult{}, err
	}

	avatarUploadToken, err := s.newOpaqueID(avatarUploadTokenPrefix)
	if err != nil {
		return CreateUploadResult{}, fmt.Errorf("generate avatar upload token: %w", err)
	}

	now := s.now().UTC()
	expiresAt := now.Add(s.uploadTTL)
	uploadKey := buildUploadObjectKey(input.ViewerUserID, avatarUploadToken, metadata.FileName)
	expiresIn := expiresAt.Sub(now)

	presignedUpload, err := s.storage.PresignPutObject(ctx, s.uploadBucketName, uploadKey, metadata.MimeType, expiresIn)
	if err != nil {
		return CreateUploadResult{}, fmt.Errorf("%w: presign avatar upload token=%s: %v", ErrStorageFailure, avatarUploadToken, err)
	}

	if err := s.store.SaveUpload(ctx, avatarUploadToken, storedUpload{
		ExpiresAt:     expiresAt,
		FileName:      metadata.FileName,
		FileSizeBytes: metadata.FileSizeBytes,
		MimeType:      metadata.MimeType,
		State:         uploadStateCreated,
		UploadKey:     uploadKey,
		ViewerUserID:  input.ViewerUserID.String(),
	}, s.uploadTTL); err != nil {
		return CreateUploadResult{}, fmt.Errorf("%w: save avatar upload token=%s: %v", ErrStorageFailure, avatarUploadToken, err)
	}

	return CreateUploadResult{
		AvatarUploadToken: avatarUploadToken,
		ExpiresAt:         expiresAt,
		UploadTarget: UploadTarget{
			FileName: metadata.FileName,
			MimeType: metadata.MimeType,
			Upload: DirectUpload{
				Headers: presignedUpload.Headers,
				Method:  "PUT",
				URL:     presignedUpload.URL,
			},
		},
	}, nil
}

// CompleteUpload は uploaded avatar object を検証し stable URL を確定します。
func (s *Service) CompleteUpload(ctx context.Context, input CompleteUploadInput) (CompleteUploadResult, error) {
	if s == nil {
		return CompleteUploadResult{}, fmt.Errorf("creator avatar service is nil")
	}

	upload, avatarUploadToken, err := s.loadOwnedUpload(ctx, input.ViewerUserID, input.AvatarUploadToken)
	if err != nil {
		return CompleteUploadResult{}, err
	}

	if upload.State == uploadStateComplete {
		completed, completedErr := buildCompletedUpload(upload, avatarUploadToken)
		if completedErr != nil {
			return CompleteUploadResult{}, completedErr
		}

		return CompleteUploadResult{
			Avatar: completed,
		}, nil
	}
	if upload.State == uploadStateConsumed {
		return CompleteUploadResult{}, ErrUploadNotFound
	}

	objectMetadata, err := s.storage.HeadObject(ctx, s.uploadBucketName, upload.UploadKey)
	if err != nil {
		if isObjectMissingError(err) {
			return CompleteUploadResult{}, ErrUploadIncomplete
		}
		return CompleteUploadResult{}, fmt.Errorf("%w: head avatar upload token=%s: %v", ErrStorageFailure, avatarUploadToken, err)
	}

	if objectMetadata.ContentLength != upload.FileSizeBytes {
		return CompleteUploadResult{}, ErrUploadIncomplete
	}

	expectedMimeType, err := normalizeImageMimeType(upload.MimeType)
	if err != nil {
		return CompleteUploadResult{}, ErrUploadIncomplete
	}
	actualMimeType, err := normalizeImageMimeType(objectMetadata.ContentType)
	if err != nil || actualMimeType != expectedMimeType {
		return CompleteUploadResult{}, ErrUploadIncomplete
	}

	objectData, err := s.storage.GetObject(ctx, s.uploadBucketName, upload.UploadKey)
	if err != nil {
		if isObjectMissingError(err) {
			return CompleteUploadResult{}, ErrUploadIncomplete
		}
		return CompleteUploadResult{}, fmt.Errorf("%w: get avatar upload token=%s: %v", ErrStorageFailure, avatarUploadToken, err)
	}

	if err := validateImageBytes(objectData.Body, expectedMimeType); err != nil {
		return CompleteUploadResult{}, ErrUploadIncomplete
	}

	remainingTTL := upload.ExpiresAt.Sub(s.now().UTC())
	if remainingTTL <= 0 {
		return CompleteUploadResult{}, ErrUploadExpired
	}

	// Reuse a deterministic delivery key so repeated completion requests stay idempotent.
	avatarAssetID := buildAvatarAssetID(avatarUploadToken)
	deliveryKey := buildDeliveryObjectKey(input.ViewerUserID, avatarAssetID, upload.FileName)
	if err := s.storage.PutObject(ctx, s.deliveryBucketName, deliveryKey, objectData.Body, expectedMimeType); err != nil {
		return CompleteUploadResult{}, fmt.Errorf("%w: store delivery avatar token=%s: %v", ErrStorageFailure, avatarUploadToken, err)
	}

	upload.AvatarAssetID = avatarAssetID
	upload.AvatarURL = buildDeliveryURL(s.deliveryBaseURL, deliveryKey)
	upload.DeliveryKey = deliveryKey
	upload.State = uploadStateComplete

	remainingTTL = upload.ExpiresAt.Sub(s.now().UTC())
	if remainingTTL <= 0 {
		return CompleteUploadResult{}, ErrUploadExpired
	}

	if err := s.store.SaveUpload(ctx, avatarUploadToken, upload, remainingTTL); err != nil {
		return CompleteUploadResult{}, fmt.Errorf("%w: save completed avatar token=%s: %v", ErrStorageFailure, avatarUploadToken, err)
	}

	completed, err := buildCompletedUpload(upload, avatarUploadToken)
	if err != nil {
		return CompleteUploadResult{}, err
	}

	return CompleteUploadResult{
		Avatar: completed,
	}, nil
}

// ResolveCompletedUpload は registration 前に completed avatar upload を解決します。
func (s *Service) ResolveCompletedUpload(ctx context.Context, viewerUserID uuid.UUID, avatarUploadToken string) (CompletedUpload, error) {
	if s == nil {
		return CompletedUpload{}, fmt.Errorf("creator avatar service is nil")
	}

	upload, normalizedToken, err := s.loadOwnedUpload(ctx, viewerUserID, avatarUploadToken)
	if err != nil {
		return CompletedUpload{}, err
	}
	if upload.State == uploadStateConsumed {
		return CompletedUpload{}, ErrUploadConsumed
	}
	if upload.State != uploadStateComplete {
		return CompletedUpload{}, ErrUploadIncomplete
	}

	return buildCompletedUpload(upload, normalizedToken)
}

// ConsumeCompletedUpload は successful registration 後に token を消費します。
func (s *Service) ConsumeCompletedUpload(ctx context.Context, viewerUserID uuid.UUID, avatarUploadToken string) error {
	if s == nil {
		return fmt.Errorf("creator avatar service is nil")
	}

	upload, normalizedToken, err := s.loadOwnedUpload(ctx, viewerUserID, avatarUploadToken)
	if err != nil {
		return err
	}
	if upload.State == uploadStateConsumed {
		return nil
	}
	if upload.State != uploadStateComplete {
		return ErrUploadIncomplete
	}

	if err := s.store.DeleteUpload(ctx, normalizedToken); err == nil {
		return nil
	} else {
		consumedAt := s.now().UTC()
		upload.ConsumedAt = &consumedAt
		upload.State = uploadStateConsumed

		remainingTTL := upload.ExpiresAt.Sub(s.now().UTC())
		if remainingTTL <= 0 {
			remainingTTL = time.Second
		}

		if saveErr := s.store.SaveUpload(ctx, normalizedToken, upload, remainingTTL); saveErr != nil {
			return fmt.Errorf("%w: consume avatar upload token=%s: delete: %v; mark consumed: %v", ErrStorageFailure, normalizedToken, err, saveErr)
		}
	}

	return nil
}

func (s *Service) loadOwnedUpload(ctx context.Context, viewerUserID uuid.UUID, avatarUploadToken string) (storedUpload, string, error) {
	normalizedToken := strings.TrimSpace(avatarUploadToken)
	if normalizedToken == "" {
		return storedUpload{}, "", ErrUploadNotFound
	}

	upload, err := s.store.GetUpload(ctx, normalizedToken)
	if err != nil {
		if errors.Is(err, ErrUploadNotFound) {
			return storedUpload{}, "", ErrUploadNotFound
		}
		return storedUpload{}, "", fmt.Errorf("%w: load avatar upload token=%s: %v", ErrStorageFailure, normalizedToken, err)
	}

	now := s.now().UTC()
	if now.After(upload.ExpiresAt) {
		_ = s.store.DeleteUpload(ctx, normalizedToken)
		return storedUpload{}, "", ErrUploadExpired
	}
	if upload.ViewerUserID != viewerUserID.String() {
		return storedUpload{}, "", ErrUploadNotFound
	}

	return upload, normalizedToken, nil
}

func buildCompletedUpload(upload storedUpload, avatarUploadToken string) (CompletedUpload, error) {
	if strings.TrimSpace(upload.AvatarAssetID) == "" || strings.TrimSpace(upload.AvatarURL) == "" {
		return CompletedUpload{}, ErrUploadIncomplete
	}

	return CompletedUpload{
		AvatarAssetID:     strings.TrimSpace(upload.AvatarAssetID),
		AvatarUploadToken: avatarUploadToken,
		AvatarURL:         strings.TrimSpace(upload.AvatarURL),
	}, nil
}

func normalizeFileMetadata(fileName string, mimeType string, fileSizeBytes int64) (storedUpload, error) {
	trimmedFileName := strings.TrimSpace(fileName)
	if trimmedFileName == "" {
		return storedUpload{}, NewValidationError("invalid_request", "avatar fileName is required")
	}

	normalizedMimeType, err := normalizeImageMimeType(mimeType)
	if err != nil {
		return storedUpload{}, NewValidationError("invalid_avatar_mime_type", "avatar mime type is invalid")
	}
	if fileSizeBytes <= 0 {
		return storedUpload{}, NewValidationError("invalid_avatar_file_size", "avatar file size is invalid")
	}
	if fileSizeBytes > maxFileSizeBytes {
		return storedUpload{}, NewValidationError("avatar_file_too_large", "avatar file size exceeds the maximum allowed size")
	}

	return storedUpload{
		FileName:      trimmedFileName,
		FileSizeBytes: fileSizeBytes,
		MimeType:      normalizedMimeType,
	}, nil
}

func normalizeImageMimeType(value string) (string, error) {
	mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(value))
	if err != nil {
		return "", err
	}

	normalized := strings.ToLower(strings.TrimSpace(mediaType))
	switch normalized {
	case "image/jpeg", "image/png", "image/webp":
		return normalized, nil
	default:
		return "", fmt.Errorf("mime type must be one of image/jpeg, image/png, image/webp")
	}
}

func validateImageBytes(body []byte, expectedMimeType string) error {
	if len(body) == 0 {
		return fmt.Errorf("image body is empty")
	}

	detectedMimeType, err := normalizeImageMimeType(http.DetectContentType(body))
	if err != nil {
		return err
	}
	if detectedMimeType != expectedMimeType {
		return fmt.Errorf("detected mime type %s does not match expected %s", detectedMimeType, expectedMimeType)
	}

	return nil
}

func buildUploadObjectKey(viewerUserID uuid.UUID, avatarUploadToken string, fileName string) string {
	return strings.Join([]string{
		"creator-avatar-upload",
		viewerUserID.String(),
		avatarUploadToken,
		sanitizeStorageSegment(fileName),
	}, "/")
}

func buildDeliveryObjectKey(viewerUserID uuid.UUID, avatarAssetID string, fileName string) string {
	return strings.Join([]string{
		"creator-avatar",
		viewerUserID.String(),
		avatarAssetID,
		sanitizeStorageSegment(fileName),
	}, "/")
}

func buildDeliveryURL(baseURL string, objectKey string) string {
	trimmedBaseURL := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	trimmedKey := strings.TrimLeft(strings.TrimSpace(objectKey), "/")
	return trimmedBaseURL + "/" + trimmedKey
}

func buildAvatarAssetID(avatarUploadToken string) string {
	normalizedToken := strings.TrimSpace(avatarUploadToken)
	if strings.HasPrefix(normalizedToken, avatarUploadTokenPrefix) {
		return avatarAssetIDPrefix + strings.TrimPrefix(normalizedToken, avatarUploadTokenPrefix)
	}

	return avatarAssetIDPrefix + normalizedToken
}

func sanitizeStorageSegment(fileName string) string {
	baseName := filepath.Base(strings.TrimSpace(fileName))
	if baseName == "." || baseName == string(filepath.Separator) || baseName == "" {
		baseName = "upload"
	}

	var builder strings.Builder
	for _, char := range baseName {
		switch {
		case char >= 'a' && char <= 'z':
			builder.WriteRune(char)
		case char >= 'A' && char <= 'Z':
			builder.WriteRune(char)
		case char >= '0' && char <= '9':
			builder.WriteRune(char)
		case char == '.' || char == '-' || char == '_':
			builder.WriteRune(char)
		default:
			builder.WriteRune('_')
		}
	}

	sanitized := strings.Trim(builder.String(), "._-")
	if sanitized == "" {
		return "upload"
	}

	return sanitized
}

func generateOpaqueID(prefix string) (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate opaque id: %w", err)
	}

	return prefix + hex.EncodeToString(randomBytes), nil
}

func isObjectMissingError(err error) bool {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "NotFound", "NoSuchKey", "NoSuchUpload":
			return true
		}
	}

	return false
}
