package creatorregistration

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
	defaultEvidenceUploadTTL       = 15 * time.Minute
	evidenceUploadStateCreated     = "created"
	evidenceUploadStateCompleted   = "completed"
	evidenceUploadMaxFileSizeBytes = 10_485_760

	evidenceUploadTokenPrefix = "vcevd_"
)

var (
	ErrEvidenceUploadExpired    = errors.New("creator registration evidence upload has expired")
	ErrEvidenceUploadIncomplete = errors.New("creator registration evidence upload is incomplete")
	ErrEvidenceUploadNotFound   = errors.New("creator registration evidence upload was not found")
	ErrEvidenceUploadStorage    = errors.New("creator registration evidence upload storage failure")
)

// ValidationError は evidence upload request payload の validation failure を表します。
type ValidationError struct {
	code    string
	message string
}

func (e *ValidationError) Error() string {
	return e.message
}

func (e *ValidationError) Code() string {
	return e.code
}

func (e *ValidationError) Message() string {
	return e.message
}

func newValidationError(code string, message string) *ValidationError {
	return &ValidationError{
		code:    code,
		message: message,
	}
}

type CreateEvidenceUploadInput struct {
	FileName      string
	FileSizeBytes int64
	Kind          string
	MimeType      string
	ViewerUserID  uuid.UUID
}

type CompleteEvidenceUploadInput struct {
	EvidenceUploadToken string
	ViewerUserID        uuid.UUID
}

type DirectUpload struct {
	Headers map[string]string
	Method  string
	URL     string
}

type EvidenceUploadTarget struct {
	FileName string
	MimeType string
	Upload   DirectUpload
}

type CreateEvidenceUploadResult struct {
	EvidenceKind        string
	EvidenceUploadToken string
	ExpiresAt           time.Time
	UploadTarget        EvidenceUploadTarget
}

type CompleteEvidenceUploadResult struct {
	Evidence            Evidence
	EvidenceKind        string
	EvidenceUploadToken string
}

type ServiceConfig struct {
	EvidenceBucketName string
	UploadTTL          time.Duration
}

type uploadStore interface {
	DeleteUpload(ctx context.Context, evidenceUploadToken string) error
	GetUpload(ctx context.Context, evidenceUploadToken string) (storedEvidenceUpload, error)
	SaveUpload(ctx context.Context, evidenceUploadToken string, upload storedEvidenceUpload, ttl time.Duration) error
}

type storage interface {
	CopyObject(ctx context.Context, sourceBucket string, sourceKey string, destinationBucket string, destinationKey string) error
	DeleteObject(ctx context.Context, bucket string, key string) error
	GetObject(ctx context.Context, bucket string, key string) (medias3.ObjectData, error)
	HeadObject(ctx context.Context, bucket string, key string) (medias3.ObjectMetadata, error)
	PresignPutObjectWithLength(ctx context.Context, bucket string, key string, contentType string, contentLength int64, expires time.Duration) (medias3.PresignedUpload, error)
}

type evidenceRepository interface {
	PrepareEvidenceUpload(ctx context.Context, userID uuid.UUID) error
	SaveEvidence(ctx context.Context, input SaveEvidenceInput) (SaveEvidenceResult, error)
}

// Service は creator registration evidence upload を扱います。
type Service struct {
	bucketName  string
	now         func() time.Time
	newOpaqueID func(prefix string) (string, error)
	repository  evidenceRepository
	storage     storage
	store       uploadStore
	uploadTTL   time.Duration
}

// NewEvidenceUploadService は evidence upload service を構築します。
func NewEvidenceUploadService(cfg ServiceConfig, storage storage, store uploadStore, repository evidenceRepository) (*Service, error) {
	if storage == nil {
		return nil, fmt.Errorf("creator registration evidence storage is required")
	}
	if store == nil {
		return nil, fmt.Errorf("creator registration evidence upload store is required")
	}
	if repository == nil {
		return nil, fmt.Errorf("creator registration evidence repository is required")
	}
	if strings.TrimSpace(cfg.EvidenceBucketName) == "" {
		return nil, fmt.Errorf("creator registration evidence bucket name is required")
	}
	if cfg.UploadTTL <= 0 {
		cfg.UploadTTL = defaultEvidenceUploadTTL
	}

	return &Service{
		bucketName:  strings.TrimSpace(cfg.EvidenceBucketName),
		now:         time.Now,
		newOpaqueID: generateOpaqueID,
		repository:  repository,
		storage:     storage,
		store:       store,
		uploadTTL:   cfg.UploadTTL,
	}, nil
}

// CreateUpload は evidence upload target を発行します。
func (s *Service) CreateUpload(ctx context.Context, input CreateEvidenceUploadInput) (CreateEvidenceUploadResult, error) {
	if s == nil {
		return CreateEvidenceUploadResult{}, fmt.Errorf("creator registration evidence service is nil")
	}

	upload, err := normalizeEvidenceMetadata(input.Kind, input.FileName, input.MimeType, input.FileSizeBytes)
	if err != nil {
		return CreateEvidenceUploadResult{}, err
	}
	if err := s.repository.PrepareEvidenceUpload(ctx, input.ViewerUserID); err != nil {
		return CreateEvidenceUploadResult{}, err
	}

	evidenceUploadToken, err := s.newOpaqueID(evidenceUploadTokenPrefix)
	if err != nil {
		return CreateEvidenceUploadResult{}, fmt.Errorf("generate evidence upload token: %w", err)
	}

	now := s.now().UTC()
	expiresAt := now.Add(s.uploadTTL)
	uploadKey := buildEvidenceUploadObjectKey(input.ViewerUserID, upload.Kind, evidenceUploadToken, upload.FileName)

	presignedUpload, err := s.storage.PresignPutObjectWithLength(
		ctx,
		s.bucketName,
		uploadKey,
		upload.MimeType,
		upload.FileSizeBytes,
		s.uploadTTL,
	)
	if err != nil {
		return CreateEvidenceUploadResult{}, fmt.Errorf("%w: presign evidence upload token=%s: %v", ErrEvidenceUploadStorage, evidenceUploadToken, err)
	}

	upload.ExpiresAt = expiresAt
	upload.State = evidenceUploadStateCreated
	upload.UploadKey = uploadKey
	upload.ViewerUserID = input.ViewerUserID.String()
	if err := s.store.SaveUpload(ctx, evidenceUploadToken, upload, s.uploadTTL); err != nil {
		return CreateEvidenceUploadResult{}, fmt.Errorf("%w: save evidence upload token=%s: %v", ErrEvidenceUploadStorage, evidenceUploadToken, err)
	}

	return CreateEvidenceUploadResult{
		EvidenceKind:        upload.Kind,
		EvidenceUploadToken: evidenceUploadToken,
		ExpiresAt:           expiresAt,
		UploadTarget: EvidenceUploadTarget{
			FileName: upload.FileName,
			MimeType: upload.MimeType,
			Upload: DirectUpload{
				Headers: presignedUpload.Headers,
				Method:  "PUT",
				URL:     presignedUpload.URL,
			},
		},
	}, nil
}

// CompleteUpload は uploaded evidence object を検証して metadata を保存します。
func (s *Service) CompleteUpload(ctx context.Context, input CompleteEvidenceUploadInput) (CompleteEvidenceUploadResult, error) {
	if s == nil {
		return CompleteEvidenceUploadResult{}, fmt.Errorf("creator registration evidence service is nil")
	}

	upload, normalizedToken, err := s.loadOwnedUpload(ctx, input.ViewerUserID, input.EvidenceUploadToken)
	if err != nil {
		return CompleteEvidenceUploadResult{}, err
	}
	if upload.State == evidenceUploadStateCompleted && upload.CompletedEvidence != nil {
		if err := s.cleanupCompletedArtifacts(ctx, normalizedToken, &upload); err != nil {
			return CompleteEvidenceUploadResult{}, err
		}
		return CompleteEvidenceUploadResult{
			Evidence:            *upload.CompletedEvidence,
			EvidenceKind:        upload.Kind,
			EvidenceUploadToken: normalizedToken,
		}, nil
	}

	objectMetadata, err := s.storage.HeadObject(ctx, s.bucketName, upload.UploadKey)
	if err != nil {
		if isObjectMissingError(err) {
			return CompleteEvidenceUploadResult{}, ErrEvidenceUploadIncomplete
		}
		return CompleteEvidenceUploadResult{}, fmt.Errorf("%w: head evidence upload token=%s: %v", ErrEvidenceUploadStorage, normalizedToken, err)
	}

	if objectMetadata.ContentLength != upload.FileSizeBytes {
		if cleanupErr := s.cleanupUploadedObject(ctx, upload); cleanupErr != nil {
			return CompleteEvidenceUploadResult{}, cleanupErr
		}
		return CompleteEvidenceUploadResult{}, ErrEvidenceUploadIncomplete
	}
	if normalizedContentType, err := normalizeEvidenceMimeType(objectMetadata.ContentType); err != nil || normalizedContentType != upload.MimeType {
		if cleanupErr := s.cleanupUploadedObject(ctx, upload); cleanupErr != nil {
			return CompleteEvidenceUploadResult{}, cleanupErr
		}
		return CompleteEvidenceUploadResult{}, ErrEvidenceUploadIncomplete
	}

	objectData, err := s.storage.GetObject(ctx, s.bucketName, upload.UploadKey)
	if err != nil {
		if isObjectMissingError(err) {
			return CompleteEvidenceUploadResult{}, ErrEvidenceUploadIncomplete
		}
		return CompleteEvidenceUploadResult{}, fmt.Errorf("%w: get evidence upload token=%s: %v", ErrEvidenceUploadStorage, normalizedToken, err)
	}
	if err := validateEvidenceBytes(objectData.Body, upload.MimeType); err != nil {
		if cleanupErr := s.cleanupUploadedObject(ctx, upload); cleanupErr != nil {
			return CompleteEvidenceUploadResult{}, cleanupErr
		}
		return CompleteEvidenceUploadResult{}, ErrEvidenceUploadIncomplete
	}
	if remainingTTL := upload.ExpiresAt.Sub(s.now().UTC()); remainingTTL <= 0 {
		if cleanupErr := s.cleanupUploadedObject(ctx, upload); cleanupErr != nil {
			return CompleteEvidenceUploadResult{}, cleanupErr
		}
		return CompleteEvidenceUploadResult{}, ErrEvidenceUploadExpired
	}

	finalStorageKey := buildCompletedEvidenceObjectKey(input.ViewerUserID, upload.Kind, normalizedToken, upload.FileName)
	if err := s.storage.CopyObject(ctx, s.bucketName, upload.UploadKey, s.bucketName, finalStorageKey); err != nil {
		return CompleteEvidenceUploadResult{}, fmt.Errorf("%w: finalize evidence upload token=%s: %v", ErrEvidenceUploadStorage, normalizedToken, err)
	}

	savedEvidence, err := s.repository.SaveEvidence(ctx, SaveEvidenceInput{
		FileName:      upload.FileName,
		FileSizeBytes: upload.FileSizeBytes,
		Kind:          upload.Kind,
		MimeType:      upload.MimeType,
		StorageBucket: s.bucketName,
		StorageKey:    finalStorageKey,
		UploadedAt:    s.now().UTC(),
		UserID:        input.ViewerUserID,
	})
	if err != nil {
		_ = s.storage.DeleteObject(ctx, s.bucketName, finalStorageKey)
		if errors.Is(err, ErrRegistrationStateConflict) {
			if cleanupErr := s.cleanupUploadedObject(ctx, upload); cleanupErr != nil {
				return CompleteEvidenceUploadResult{}, cleanupErr
			}
		}
		return CompleteEvidenceUploadResult{}, err
	}

	upload.CompletedEvidence = &savedEvidence.Evidence
	upload.PendingDelete = savedEvidence.ReplacedObject
	upload.State = evidenceUploadStateCompleted
	remainingTTL := upload.ExpiresAt.Sub(s.now().UTC())
	if remainingTTL <= 0 {
		return CompleteEvidenceUploadResult{}, ErrEvidenceUploadExpired
	}
	if err := s.store.SaveUpload(ctx, normalizedToken, upload, remainingTTL); err != nil {
		_ = s.cleanupUploadedObject(ctx, upload)
		return CompleteEvidenceUploadResult{}, fmt.Errorf("%w: save completed evidence upload token=%s: %v", ErrEvidenceUploadStorage, normalizedToken, err)
	}
	if err := s.cleanupCompletedArtifacts(ctx, normalizedToken, &upload); err != nil {
		return CompleteEvidenceUploadResult{}, err
	}

	return CompleteEvidenceUploadResult{
		Evidence:            savedEvidence.Evidence,
		EvidenceKind:        upload.Kind,
		EvidenceUploadToken: normalizedToken,
	}, nil
}

func buildEvidenceUploadObjectKey(viewerUserID uuid.UUID, kind string, evidenceUploadToken string, fileName string) string {
	return strings.Join([]string{
		"creator-registration-evidence",
		viewerUserID.String(),
		kind,
		evidenceUploadToken,
		sanitizeStorageSegment(fileName),
	}, "/")
}

func buildCompletedEvidenceObjectKey(viewerUserID uuid.UUID, kind string, evidenceUploadToken string, fileName string) string {
	return strings.Join([]string{
		"creator-registration-evidence-final",
		viewerUserID.String(),
		kind,
		evidenceUploadToken,
		sanitizeStorageSegment(fileName),
	}, "/")
}

func generateOpaqueID(prefix string) (string, error) {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("generate opaque id entropy: %w", err)
	}

	return prefix + hex.EncodeToString(buffer), nil
}

func (s *Service) loadOwnedUpload(ctx context.Context, viewerUserID uuid.UUID, evidenceUploadToken string) (storedEvidenceUpload, string, error) {
	normalizedToken := strings.TrimSpace(evidenceUploadToken)
	if normalizedToken == "" {
		return storedEvidenceUpload{}, "", ErrEvidenceUploadNotFound
	}

	upload, err := s.store.GetUpload(ctx, normalizedToken)
	if err != nil {
		if errors.Is(err, ErrEvidenceUploadNotFound) {
			return storedEvidenceUpload{}, "", ErrEvidenceUploadNotFound
		}
		return storedEvidenceUpload{}, "", fmt.Errorf("%w: load evidence upload token=%s: %v", ErrEvidenceUploadStorage, normalizedToken, err)
	}

	if upload.ViewerUserID != viewerUserID.String() {
		return storedEvidenceUpload{}, "", ErrEvidenceUploadNotFound
	}
	if s.now().UTC().After(upload.ExpiresAt) {
		if cleanupErr := s.cleanupExpiredUpload(ctx, upload); cleanupErr != nil {
			return storedEvidenceUpload{}, "", cleanupErr
		}
		_ = s.store.DeleteUpload(ctx, normalizedToken)
		return storedEvidenceUpload{}, "", ErrEvidenceUploadExpired
	}

	return upload, normalizedToken, nil
}

func (s *Service) cleanupPendingDelete(ctx context.Context, evidenceUploadToken string, upload *storedEvidenceUpload) error {
	if upload == nil || upload.PendingDelete == nil {
		return nil
	}

	object := upload.PendingDelete
	if err := s.storage.DeleteObject(ctx, object.Bucket, object.Key); err != nil && !isObjectMissingError(err) {
		return fmt.Errorf("%w: delete replaced evidence object bucket=%s key=%s: %v", ErrEvidenceUploadStorage, object.Bucket, object.Key, err)
	}
	upload.PendingDelete = nil

	remainingTTL := upload.ExpiresAt.Sub(s.now().UTC())
	if remainingTTL <= 0 {
		return ErrEvidenceUploadExpired
	}
	if err := s.store.SaveUpload(ctx, evidenceUploadToken, *upload, remainingTTL); err != nil {
		return fmt.Errorf("%w: save cleanup-complete evidence upload token=%s: %v", ErrEvidenceUploadStorage, evidenceUploadToken, err)
	}

	return nil
}

func (s *Service) cleanupCompletedArtifacts(ctx context.Context, evidenceUploadToken string, upload *storedEvidenceUpload) error {
	if upload == nil {
		return nil
	}
	if strings.TrimSpace(upload.UploadKey) != "" {
		if err := s.cleanupUploadedObject(ctx, *upload); err != nil {
			return err
		}
		upload.UploadKey = ""

		remainingTTL := upload.ExpiresAt.Sub(s.now().UTC())
		if remainingTTL <= 0 {
			return ErrEvidenceUploadExpired
		}
		if err := s.store.SaveUpload(ctx, evidenceUploadToken, *upload, remainingTTL); err != nil {
			return fmt.Errorf("%w: save upload-cleanup evidence upload token=%s: %v", ErrEvidenceUploadStorage, evidenceUploadToken, err)
		}
	}

	return s.cleanupPendingDelete(ctx, evidenceUploadToken, upload)
}

func (s *Service) cleanupExpiredUpload(ctx context.Context, upload storedEvidenceUpload) error {
	if upload.State == evidenceUploadStateCompleted && upload.CompletedEvidence != nil {
		if err := s.cleanupUploadedObject(ctx, upload); err != nil {
			return err
		}
		if upload.PendingDelete == nil {
			return nil
		}

		object := upload.PendingDelete
		if err := s.storage.DeleteObject(ctx, object.Bucket, object.Key); err != nil && !isObjectMissingError(err) {
			return fmt.Errorf("%w: delete expired pending evidence object bucket=%s key=%s: %v", ErrEvidenceUploadStorage, object.Bucket, object.Key, err)
		}

		return nil
	}

	return s.cleanupUploadedObject(ctx, upload)
}

func (s *Service) cleanupUploadedObject(ctx context.Context, upload storedEvidenceUpload) error {
	if strings.TrimSpace(upload.UploadKey) == "" {
		return nil
	}

	if err := s.storage.DeleteObject(ctx, s.bucketName, upload.UploadKey); err != nil && !isObjectMissingError(err) {
		return fmt.Errorf("%w: delete evidence upload key=%s: %v", ErrEvidenceUploadStorage, upload.UploadKey, err)
	}

	return nil
}

func validateEvidenceBytes(body []byte, expectedMimeType string) error {
	if len(body) == 0 {
		return fmt.Errorf("evidence body is empty")
	}

	switch expectedMimeType {
	case "image/jpeg", "image/png", "image/webp":
		detectedMimeType, err := normalizeEvidenceMimeType(http.DetectContentType(body))
		if err != nil {
			return err
		}
		if detectedMimeType != expectedMimeType {
			return fmt.Errorf("detected mime type %s does not match expected %s", detectedMimeType, expectedMimeType)
		}
		return nil
	case "application/pdf":
		if http.DetectContentType(body) != "application/pdf" {
			return fmt.Errorf("detected mime type does not match expected pdf")
		}
		return nil
	default:
		return fmt.Errorf("unexpected evidence mime type %s", expectedMimeType)
	}
}

func isObjectMissingError(err error) bool {
	if err == nil {
		return false
	}

	var apiErr smithy.APIError
	if !errors.As(err, &apiErr) {
		return false
	}

	code := strings.TrimSpace(apiErr.ErrorCode())
	return code == "NotFound" || code == "NoSuchKey"
}

func normalizeEvidenceMetadata(kind string, fileName string, mimeType string, fileSizeBytes int64) (storedEvidenceUpload, error) {
	normalizedKind, err := normalizeEvidenceKind(kind)
	if err != nil {
		return storedEvidenceUpload{}, err
	}
	trimmedFileName := strings.TrimSpace(fileName)
	if trimmedFileName == "" {
		return storedEvidenceUpload{}, newValidationError("invalid_request", "evidence fileName is required")
	}
	normalizedMimeType, err := normalizeEvidenceMimeType(mimeType)
	if err != nil {
		return storedEvidenceUpload{}, newValidationError("invalid_evidence_mime_type", "evidence mime type is invalid")
	}
	if fileSizeBytes <= 0 {
		return storedEvidenceUpload{}, newValidationError("invalid_evidence_file_size", "evidence file size is invalid")
	}
	if fileSizeBytes > evidenceUploadMaxFileSizeBytes {
		return storedEvidenceUpload{}, newValidationError("evidence_file_too_large", "evidence file size exceeds the maximum allowed size")
	}

	return storedEvidenceUpload{
		FileName:      trimmedFileName,
		FileSizeBytes: fileSizeBytes,
		Kind:          normalizedKind,
		MimeType:      normalizedMimeType,
	}, nil
}

func normalizeEvidenceKind(value string) (string, error) {
	normalized := strings.TrimSpace(value)
	switch normalized {
	case EvidenceKindGovernmentID, EvidenceKindPayoutProof:
		return normalized, nil
	default:
		return "", newValidationError("invalid_evidence_kind", "evidence kind is invalid")
	}
}

func normalizeEvidenceMimeType(value string) (string, error) {
	mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(value))
	if err != nil {
		return "", err
	}

	normalized := strings.ToLower(strings.TrimSpace(mediaType))
	switch normalized {
	case "image/jpeg", "image/png", "image/webp", "application/pdf":
		return normalized, nil
	default:
		return "", fmt.Errorf("mime type must be one of image/jpeg, image/png, image/webp, application/pdf")
	}
}

func sanitizeStorageSegment(fileName string) string {
	baseName := filepath.Base(strings.TrimSpace(fileName))
	if baseName == "." || baseName == string(filepath.Separator) || baseName == "" {
		return "upload"
	}

	replacer := strings.NewReplacer(" ", "-", "..", ".", "\\", "-", "/", "-")
	sanitized := replacer.Replace(baseName)
	sanitized = strings.Trim(sanitized, ".-")
	if sanitized == "" {
		return "upload"
	}

	return sanitized
}
