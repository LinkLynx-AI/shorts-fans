package creatorupload

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/smithy-go"
	"github.com/google/uuid"

	medias3 "github.com/LinkLynx-AI/shorts-fans/backend/internal/s3"
)

const (
	defaultPackageTTL         = 15 * time.Minute
	roleMain                  = "main"
	roleShort                 = "short"
	stateDraft                = "draft"
	stateUploaded             = "uploaded"
	storageProviderS3         = "s3"
	currencyJPY               = "JPY"
	processingJobStatusQueued = "queued"

	packageTokenPrefix = "cupkg_"
	mainEntryPrefix    = "cu_main_"
	shortEntryPrefix   = "cu_short_"
)

var (
	// ErrPackageExpired は upload package が存在しない、または期限切れのことを表します。
	ErrPackageExpired = errors.New("creator upload package has expired")
	// ErrStorageFailure は S3 や package store などの dependency が失敗したことを表します。
	ErrStorageFailure = errors.New("creator upload storage failure")
	// ErrUploadFailure は expected object が揃っていないか package と不整合なことを表します。
	ErrUploadFailure = errors.New("creator upload package is incomplete")
)

// ValidationError は request payload の semantic validation failure を表します。
type ValidationError struct {
	message string
}

// NewValidationError は transport layer が返す validation error を構築します。
func NewValidationError(message string) *ValidationError {
	return &ValidationError{message: message}
}

func (e *ValidationError) Error() string {
	return e.message
}

// Message は transport layer がそのまま返せる validation message を返します。
func (e *ValidationError) Message() string {
	return e.message
}

// FileMetadata は creator upload の file metadata を表します。
type FileMetadata struct {
	FileName      string
	FileSizeBytes int64
	MimeType      string
}

// UploadEntryReference は completion request の 1 entry 参照を表します。
type UploadEntryReference struct {
	UploadEntryID string
}

// CompletePackageMain は completion request の main metadata を表します。
type CompletePackageMain struct {
	ConsentConfirmed   bool
	OwnershipConfirmed bool
	PriceJpy           int64
	UploadEntryID      string
}

// CompletePackageShort は completion request の short metadata を表します。
type CompletePackageShort struct {
	Caption       *string
	UploadEntryID string
}

// CreatePackageInput は initiation request を表します。
type CreatePackageInput struct {
	CreatorUserID uuid.UUID
	Main          *FileMetadata
	Shorts        []FileMetadata
}

// CompletePackageInput は completion request を表します。
type CompletePackageInput struct {
	CreatorUserID uuid.UUID
	PackageToken  string
	Main          *CompletePackageMain
	Shorts        []CompletePackageShort
}

// UploadTarget は client が raw bucket へ直接 upload する target を表します。
type UploadTarget struct {
	FileName      string
	MimeType      string
	Role          string
	Upload        DirectUpload
	UploadEntryID string
}

// DirectUpload は presigned PUT request を表します。
type DirectUpload struct {
	Headers map[string]string
	Method  string
	URL     string
}

// UploadTargetSet は main / shorts の upload target 一覧を表します。
type UploadTargetSet struct {
	Main   UploadTarget
	Shorts []UploadTarget
}

// CreatePackageResult は initiation response を表します。
type CreatePackageResult struct {
	ExpiresAt     time.Time
	PackageToken  string
	UploadTargets UploadTargetSet
}

// CreatedMediaAsset は completion 後に作成された media asset を表します。
type CreatedMediaAsset struct {
	ID              uuid.UUID
	MimeType        string
	ProcessingState string
}

// CreatedMain は completion 後に作成された draft main を表します。
type CreatedMain struct {
	ID         uuid.UUID
	MediaAsset CreatedMediaAsset
	State      string
}

// CreatedShort は completion 後に作成された draft short を表します。
type CreatedShort struct {
	CanonicalMainID uuid.UUID
	ID              uuid.UUID
	MediaAsset      CreatedMediaAsset
	State           string
}

// CompletePackageResult は completion response を表します。
type CompletePackageResult struct {
	Main   CreatedMain
	Shorts []CreatedShort
}

// ServiceConfig は creator upload service の実行設定を表します。
type ServiceConfig struct {
	PackageTTL         time.Duration
	RawBucketName      string
	ProcessingNotifier ProcessingNotifier
}

type packageStore interface {
	SavePackage(ctx context.Context, packageToken string, pkg storedPackage, ttl time.Duration) error
	GetPackage(ctx context.Context, packageToken string) (storedPackage, error)
	DeletePackage(ctx context.Context, packageToken string) error
}

type draftRepository interface {
	CreateDraftPackage(ctx context.Context, input createDraftPackageInput) (CompletePackageResult, error)
}

// ProcessingNotifier は completion 後の media processing wake-up を best-effort で通知します。
type ProcessingNotifier interface {
	NotifyProcessingQueued(ctx context.Context, mediaAssetIDs []uuid.UUID) error
}

// Storage は creator upload に必要な S3 操作を表します。
type Storage interface {
	PresignPutObject(ctx context.Context, bucket string, key string, contentType string, expires time.Duration) (medias3.PresignedUpload, error)
	HeadObject(ctx context.Context, bucket string, key string) (medias3.ObjectMetadata, error)
}

// Service は creator-private upload initiation / completion を扱います。
type Service struct {
	now           func() time.Time
	newToken      func(prefix string) (string, error)
	packageStore  packageStore
	packageTTL    time.Duration
	rawBucketName string
	repository    draftRepository
	storage       Storage
	notifier      ProcessingNotifier
}

// NewService は creator upload service を構築します。
func NewService(cfg ServiceConfig, storage Storage, packageStore packageStore, repository draftRepository) (*Service, error) {
	if storage == nil {
		return nil, fmt.Errorf("creator upload storage is required")
	}
	if packageStore == nil {
		return nil, fmt.Errorf("creator upload package store is required")
	}
	if repository == nil {
		return nil, fmt.Errorf("creator upload repository is required")
	}
	if strings.TrimSpace(cfg.RawBucketName) == "" {
		return nil, fmt.Errorf("creator upload raw bucket name is required")
	}
	if cfg.PackageTTL <= 0 {
		cfg.PackageTTL = defaultPackageTTL
	}

	return &Service{
		now:           time.Now,
		newToken:      generateOpaqueID,
		packageStore:  packageStore,
		packageTTL:    cfg.PackageTTL,
		rawBucketName: strings.TrimSpace(cfg.RawBucketName),
		repository:    repository,
		storage:       storage,
		notifier:      cfg.ProcessingNotifier,
	}, nil
}

// CreatePackage は upload initiation を処理します。
func (s *Service) CreatePackage(ctx context.Context, input CreatePackageInput) (CreatePackageResult, error) {
	if s == nil {
		return CreatePackageResult{}, fmt.Errorf("creator upload service is nil")
	}

	mainMetadata, err := validateMainMetadata(input.Main)
	if err != nil {
		return CreatePackageResult{}, err
	}
	shortMetadata, err := validateShortMetadata(input.Shorts)
	if err != nil {
		return CreatePackageResult{}, err
	}

	packageToken, err := s.newToken(packageTokenPrefix)
	if err != nil {
		return CreatePackageResult{}, fmt.Errorf("generate package token: %w", err)
	}

	expiresAt := s.now().UTC().Add(s.packageTTL)
	mainEntry, mainTarget, err := s.buildEntry(ctx, input.CreatorUserID, packageToken, roleMain, *mainMetadata, mainEntryPrefix, expiresAt)
	if err != nil {
		return CreatePackageResult{}, err
	}

	shortEntries := make([]storedEntry, 0, len(shortMetadata))
	shortTargets := make([]UploadTarget, 0, len(shortMetadata))
	for _, metadata := range shortMetadata {
		entry, target, buildErr := s.buildEntry(ctx, input.CreatorUserID, packageToken, roleShort, metadata, shortEntryPrefix, expiresAt)
		if buildErr != nil {
			return CreatePackageResult{}, buildErr
		}

		shortEntries = append(shortEntries, entry)
		shortTargets = append(shortTargets, target)
	}

	if err := s.packageStore.SavePackage(ctx, packageToken, storedPackage{
		CreatorUserID: input.CreatorUserID.String(),
		ExpiresAt:     expiresAt,
		Main:          mainEntry,
		Shorts:        shortEntries,
	}, s.packageTTL); err != nil {
		return CreatePackageResult{}, wrapPackageStoreFailure("save", packageToken, err)
	}

	return CreatePackageResult{
		ExpiresAt:    expiresAt,
		PackageToken: packageToken,
		UploadTargets: UploadTargetSet{
			Main:   mainTarget,
			Shorts: shortTargets,
		},
	}, nil
}

// CompletePackage は raw bucket object を検証し、draft main / shorts を作成します。
func (s *Service) CompletePackage(ctx context.Context, input CompletePackageInput) (CompletePackageResult, error) {
	if s == nil {
		return CompletePackageResult{}, fmt.Errorf("creator upload service is nil")
	}

	packageToken := strings.TrimSpace(input.PackageToken)
	if packageToken == "" {
		return CompletePackageResult{}, &ValidationError{message: "packageToken is required"}
	}

	mainSelection, shortSelections, err := validateCompletionSelections(input.Main, input.Shorts)
	if err != nil {
		return CompletePackageResult{}, err
	}

	pkg, err := s.packageStore.GetPackage(ctx, packageToken)
	if err != nil {
		if errors.Is(err, ErrPackageNotFound) {
			return CompletePackageResult{}, ErrPackageExpired
		}
		return CompletePackageResult{}, wrapPackageStoreFailure("load", packageToken, err)
	}

	now := s.now().UTC()
	if now.After(pkg.ExpiresAt) {
		_ = s.packageStore.DeletePackage(ctx, packageToken)
		return CompletePackageResult{}, ErrPackageExpired
	}
	if pkg.ConsumedAt != nil {
		return CompletePackageResult{}, ErrUploadFailure
	}
	if pkg.CreatorUserID != input.CreatorUserID.String() {
		return CompletePackageResult{}, ErrUploadFailure
	}

	mainEntry, orderedShorts, err := matchPackageEntries(pkg, mainSelection, shortSelections)
	if err != nil {
		return CompletePackageResult{}, err
	}

	if err := s.verifyObject(ctx, mainEntry); err != nil {
		return CompletePackageResult{}, err
	}
	for _, short := range orderedShorts {
		if err := s.verifyObject(ctx, short.Entry); err != nil {
			return CompletePackageResult{}, err
		}
	}

	result, err := s.repository.CreateDraftPackage(ctx, createDraftPackageInput{
		CreatorUserID: input.CreatorUserID,
		Main:          mainEntry,
		MainConsent:   mainSelection.ConsentConfirmed,
		MainOwnership: mainSelection.OwnershipConfirmed,
		MainPriceJpy:  mainSelection.PriceJpy,
		RawBucketName: s.rawBucketName,
		Shorts:        orderedShorts,
	})
	if err != nil {
		return CompletePackageResult{}, fmt.Errorf("persist upload package token=%s: %w", packageToken, err)
	}
	if s.notifier != nil {
		// DB の queued job を source of truth とし、wake-up 通知失敗では API を落とさない。
		_ = s.notifier.NotifyProcessingQueued(ctx, collectCreatedMediaAssetIDs(result))
	}

	if err := s.consumePackage(ctx, packageToken, pkg); err != nil {
		return CompletePackageResult{}, err
	}

	return result, nil
}

func (s *Service) consumePackage(ctx context.Context, packageToken string, pkg storedPackage) error {
	if err := s.packageStore.DeletePackage(ctx, packageToken); err == nil {
		return nil
	} else {
		consumedAt := s.now().UTC()
		pkg.ConsumedAt = &consumedAt

		remainingTTL := time.Until(pkg.ExpiresAt)
		if remainingTTL <= 0 {
			remainingTTL = time.Second
		}

		if saveErr := s.packageStore.SavePackage(ctx, packageToken, pkg, remainingTTL); saveErr != nil {
			return fmt.Errorf("%w: consume upload package token=%s: delete: %v; mark consumed: %v", ErrStorageFailure, packageToken, err, saveErr)
		}
	}

	return nil
}

func wrapPackageStoreFailure(operation string, packageToken string, err error) error {
	return fmt.Errorf("%w: %s upload package token=%s: %v", ErrStorageFailure, operation, packageToken, err)
}

func (s *Service) buildEntry(
	ctx context.Context,
	creatorUserID uuid.UUID,
	packageToken string,
	role string,
	metadata FileMetadata,
	entryPrefix string,
	expiresAt time.Time,
) (storedEntry, UploadTarget, error) {
	uploadEntryID, err := s.newToken(entryPrefix)
	if err != nil {
		return storedEntry{}, UploadTarget{}, fmt.Errorf("generate upload entry id role=%s: %w", role, err)
	}

	storageKey := buildStorageKey(creatorUserID, packageToken, role, uploadEntryID, metadata.FileName)
	entry := storedEntry{
		FileName:      metadata.FileName,
		FileSizeBytes: metadata.FileSizeBytes,
		MimeType:      metadata.MimeType,
		Role:          role,
		StorageKey:    storageKey,
		UploadEntryID: uploadEntryID,
	}

	expiresIn := expiresAt.Sub(s.now().UTC())
	presignedUpload, err := s.storage.PresignPutObject(ctx, s.rawBucketName, storageKey, metadata.MimeType, expiresIn)
	if err != nil {
		return storedEntry{}, UploadTarget{}, fmt.Errorf("%w: presign upload role=%s: %v", ErrStorageFailure, role, err)
	}

	return entry, UploadTarget{
		FileName: metadata.FileName,
		MimeType: metadata.MimeType,
		Role:     role,
		Upload: DirectUpload{
			Headers: presignedUpload.Headers,
			Method:  "PUT",
			URL:     presignedUpload.URL,
		},
		UploadEntryID: uploadEntryID,
	}, nil
}

func (s *Service) verifyObject(ctx context.Context, entry storedEntry) error {
	metadata, err := s.storage.HeadObject(ctx, s.rawBucketName, entry.StorageKey)
	if err != nil {
		if isObjectMissingError(err) {
			return ErrUploadFailure
		}
		return fmt.Errorf("%w: verify object entry=%s: %v", ErrStorageFailure, entry.UploadEntryID, err)
	}

	if metadata.ContentLength != entry.FileSizeBytes {
		return ErrUploadFailure
	}

	expectedMimeType, err := normalizeVideoMimeType(entry.MimeType)
	if err != nil {
		return fmt.Errorf("normalize expected mime type entry=%s: %w", entry.UploadEntryID, err)
	}
	actualMimeType, err := normalizeVideoMimeType(metadata.ContentType)
	if err != nil {
		return ErrUploadFailure
	}
	if actualMimeType != expectedMimeType {
		return ErrUploadFailure
	}

	return nil
}

func validateMainMetadata(main *FileMetadata) (*FileMetadata, error) {
	if main == nil {
		return nil, &ValidationError{message: "main is required"}
	}

	normalized, err := normalizeFileMetadata(*main, "main")
	if err != nil {
		return nil, err
	}

	return &normalized, nil
}

func validateShortMetadata(shorts []FileMetadata) ([]FileMetadata, error) {
	if len(shorts) == 0 {
		return nil, &ValidationError{message: "at least one short is required"}
	}

	normalized := make([]FileMetadata, 0, len(shorts))
	for _, short := range shorts {
		item, err := normalizeFileMetadata(short, "short")
		if err != nil {
			return nil, err
		}
		normalized = append(normalized, item)
	}

	return normalized, nil
}

func validateCompletionSelections(main *CompletePackageMain, shorts []CompletePackageShort) (CompletePackageMain, []CompletePackageShort, error) {
	if main == nil || strings.TrimSpace(main.UploadEntryID) == "" {
		return CompletePackageMain{}, nil, &ValidationError{message: "main uploadEntryId is required"}
	}
	if main.PriceJpy <= 0 {
		return CompletePackageMain{}, nil, &ValidationError{message: "main priceJpy must be greater than zero"}
	}
	if !main.OwnershipConfirmed {
		return CompletePackageMain{}, nil, &ValidationError{message: "main ownershipConfirmed must be true"}
	}
	if !main.ConsentConfirmed {
		return CompletePackageMain{}, nil, &ValidationError{message: "main consentConfirmed must be true"}
	}
	if len(shorts) == 0 {
		return CompletePackageMain{}, nil, &ValidationError{message: "at least one short is required"}
	}

	normalizedShorts := make([]CompletePackageShort, 0, len(shorts))
	for _, short := range shorts {
		uploadEntryID := strings.TrimSpace(short.UploadEntryID)
		if uploadEntryID == "" {
			return CompletePackageMain{}, nil, &ValidationError{message: "short uploadEntryId is required"}
		}

		normalizedShorts = append(normalizedShorts, CompletePackageShort{
			Caption:       normalizeOptionalCaption(short.Caption),
			UploadEntryID: uploadEntryID,
		})
	}

	return CompletePackageMain{
		ConsentConfirmed:   main.ConsentConfirmed,
		OwnershipConfirmed: main.OwnershipConfirmed,
		PriceJpy:           main.PriceJpy,
		UploadEntryID:      strings.TrimSpace(main.UploadEntryID),
	}, normalizedShorts, nil
}

func normalizeFileMetadata(metadata FileMetadata, label string) (FileMetadata, error) {
	fileName := strings.TrimSpace(metadata.FileName)
	if fileName == "" {
		return FileMetadata{}, &ValidationError{message: fmt.Sprintf("%s fileName is required", label)}
	}

	mimeType, err := normalizeVideoMimeType(metadata.MimeType)
	if err != nil {
		return FileMetadata{}, &ValidationError{message: fmt.Sprintf("%s mimeType must be video/*", label)}
	}
	if metadata.FileSizeBytes <= 0 {
		return FileMetadata{}, &ValidationError{message: fmt.Sprintf("%s fileSizeBytes must be greater than zero", label)}
	}

	return FileMetadata{
		FileName:      fileName,
		FileSizeBytes: metadata.FileSizeBytes,
		MimeType:      mimeType,
	}, nil
}

func collectCreatedMediaAssetIDs(result CompletePackageResult) []uuid.UUID {
	ids := make([]uuid.UUID, 0, 1+len(result.Shorts))
	if result.Main.MediaAsset.ID != uuid.Nil {
		ids = append(ids, result.Main.MediaAsset.ID)
	}
	for _, short := range result.Shorts {
		if short.MediaAsset.ID == uuid.Nil {
			continue
		}
		ids = append(ids, short.MediaAsset.ID)
	}

	return ids
}

func normalizeVideoMimeType(value string) (string, error) {
	mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(value))
	if err != nil {
		return "", err
	}

	normalized := strings.ToLower(strings.TrimSpace(mediaType))
	if !strings.HasPrefix(normalized, "video/") {
		return "", fmt.Errorf("mime type must be video/*")
	}

	return normalized, nil
}

func matchPackageEntries(pkg storedPackage, main CompletePackageMain, shorts []CompletePackageShort) (storedEntry, []createDraftShortInput, error) {
	if pkg.Main.Role != roleMain {
		return storedEntry{}, nil, ErrUploadFailure
	}
	if pkg.Main.UploadEntryID != main.UploadEntryID {
		return storedEntry{}, nil, ErrUploadFailure
	}
	if len(shorts) != len(pkg.Shorts) {
		return storedEntry{}, nil, ErrUploadFailure
	}

	indexByID := make(map[string]storedEntry, len(pkg.Shorts))
	for _, entry := range pkg.Shorts {
		if entry.Role != roleShort {
			return storedEntry{}, nil, ErrUploadFailure
		}
		indexByID[entry.UploadEntryID] = entry
	}

	ordered := make([]createDraftShortInput, 0, len(shorts))
	seen := make(map[string]struct{}, len(shorts))
	for _, selection := range shorts {
		entry, ok := indexByID[selection.UploadEntryID]
		if !ok {
			return storedEntry{}, nil, ErrUploadFailure
		}
		if _, duplicated := seen[selection.UploadEntryID]; duplicated {
			return storedEntry{}, nil, ErrUploadFailure
		}
		seen[selection.UploadEntryID] = struct{}{}
		ordered = append(ordered, createDraftShortInput{
			Caption: selection.Caption,
			Entry:   entry,
		})
	}

	return pkg.Main, ordered, nil
}

func normalizeOptionalCaption(value *string) *string {
	if value == nil {
		return nil
	}

	normalized := strings.TrimSpace(*value)
	if normalized == "" {
		return nil
	}

	return &normalized
}

func buildStorageKey(creatorUserID uuid.UUID, packageToken string, role string, uploadEntryID string, fileName string) string {
	return strings.Join([]string{
		"creator-upload",
		creatorUserID.String(),
		packageToken,
		role,
		uploadEntryID,
		sanitizeStorageSegment(fileName),
	}, "/")
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
