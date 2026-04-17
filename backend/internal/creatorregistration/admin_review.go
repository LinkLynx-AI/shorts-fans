package creatorregistration

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	medias3 "github.com/LinkLynx-AI/shorts-fans/backend/internal/s3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const defaultReviewEvidenceAccessTTL = 15 * time.Minute

// ReviewQueueItem は admin review queue に表示する creator registration summary です。
type ReviewQueueItem struct {
	CreatorBio    string
	LegalName     string
	Review        ReviewTimeline
	SharedProfile SharedProfilePreview
	State         string
	UserID        uuid.UUID
}

// ReviewCaseIntake は admin detail に表示する intake 情報です。
type ReviewCaseIntake struct {
	AcceptsConsentResponsibility bool
	BirthDate                    string
	DeclaresNoProhibitedCategory bool
	LegalName                    string
	PayoutRecipientName          string
	PayoutRecipientType          string
}

// ReviewEvidence は admin detail で参照できる signed URL 付き evidence です。
type ReviewEvidence struct {
	Evidence
	AccessURL string
}

// ReviewCase は admin review detail に表示する creator registration case 全体です。
type ReviewCase struct {
	CreatorBio    string
	Evidences     []ReviewEvidence
	Intake        ReviewCaseIntake
	Rejection     *Rejection
	Review        ReviewTimeline
	SharedProfile SharedProfilePreview
	State         string
	UserID        uuid.UUID
}

// ReviewServiceConfig は admin review service の設定です。
type ReviewServiceConfig struct {
	EvidenceAccessTTL time.Duration
}

// ReviewService は admin creator review queue / detail / decision を扱います。
type ReviewService struct {
	evidenceAccessTTL time.Duration
	repository        *Repository
	signEvidenceURL   func(ctx context.Context, bucket string, key string, expires time.Duration) (string, error)
}

// NewReviewService は admin creator review service を構築します。
func NewReviewService(cfg ReviewServiceConfig, signer *medias3.Client, repository *Repository) (*ReviewService, error) {
	if signer == nil {
		return nil, fmt.Errorf("creator registration review signer is required")
	}
	if repository == nil || repository.queries == nil {
		return nil, fmt.Errorf("creator registration repository が初期化されていません")
	}
	if cfg.EvidenceAccessTTL <= 0 {
		cfg.EvidenceAccessTTL = defaultReviewEvidenceAccessTTL
	}

	return &ReviewService{
		evidenceAccessTTL: cfg.EvidenceAccessTTL,
		repository:        repository,
		signEvidenceURL:   signer.PresignGetObject,
	}, nil
}

// ListCases は state filter に一致する review queue を返します。
func (s *ReviewService) ListCases(ctx context.Context, state string) ([]ReviewQueueItem, error) {
	if s == nil || s.repository == nil || s.repository.queries == nil {
		return nil, fmt.Errorf("creator registration review service が初期化されていません")
	}

	normalizedState, err := normalizeReviewListState(state)
	if err != nil {
		return nil, err
	}

	rows, err := s.repository.queries.ListCreatorRegistrationReviewCasesByState(ctx, normalizedState)
	if err != nil {
		return nil, fmt.Errorf("creator registration review queue 取得 state=%s: %w", normalizedState, err)
	}

	result := make([]ReviewQueueItem, 0, len(rows))
	for _, row := range rows {
		item, mapErr := mapReviewQueueItem(row)
		if mapErr != nil {
			return nil, mapErr
		}
		result = append(result, item)
	}

	return result, nil
}

// GetCase は user ごとの review case detail を返します。
func (s *ReviewService) GetCase(ctx context.Context, userID uuid.UUID) (ReviewCase, error) {
	snapshot, err := s.loadReviewCaseSnapshot(ctx, userID, true)
	if err != nil {
		return ReviewCase{}, err
	}

	evidenceAccessURLs, err := s.signEvidenceAccessURLs(ctx, snapshot.evidences)
	if err != nil {
		return ReviewCase{}, err
	}

	return buildReviewCase(snapshot, evidenceAccessURLs)
}

// ApplyDecision は admin review decision を反映し、更新後の review case を返します。
func (s *ReviewService) ApplyDecision(ctx context.Context, input ReviewDecisionInput) (ReviewCase, error) {
	if s == nil || s.repository == nil {
		return ReviewCase{}, fmt.Errorf("creator registration review service が初期化されていません")
	}
	if _, err := s.loadReviewCaseSnapshot(ctx, input.UserID, false); err != nil {
		return ReviewCase{}, err
	}

	if _, err := s.repository.ApplyReviewDecision(ctx, input); err != nil {
		if errors.Is(err, ErrRegistrationIncomplete) {
			return ReviewCase{}, ErrReviewCaseNotFound
		}
		return ReviewCase{}, err
	}

	return s.GetCase(ctx, input.UserID)
}

func (s *ReviewService) loadReviewCaseSnapshot(
	ctx context.Context,
	userID uuid.UUID,
	includeEvidences bool,
) (registrationSnapshot, error) {
	if s == nil || s.repository == nil || s.repository.queries == nil {
		return registrationSnapshot{}, fmt.Errorf("creator registration review service が初期化されていません")
	}

	snapshot, err := s.repository.loadSnapshot(ctx, s.repository.queries, userID, includeEvidences)
	if err != nil {
		if errors.Is(err, ErrSharedProfileNotFound) {
			return registrationSnapshot{}, ErrReviewCaseNotFound
		}
		return registrationSnapshot{}, err
	}
	if err := validateReviewCaseState(snapshot.capability); err != nil {
		return registrationSnapshot{}, err
	}

	return snapshot, nil
}

func validateReviewCaseState(capability *sqlc.AppCreatorCapability) error {
	if capability == nil {
		return ErrReviewCaseNotFound
	}
	if _, err := normalizeReviewListState(capability.State); err != nil {
		return ErrReviewCaseNotFound
	}

	return nil
}

func normalizeReviewListState(value string) (string, error) {
	switch strings.TrimSpace(value) {
	case StateSubmitted:
		return StateSubmitted, nil
	case StateApproved:
		return StateApproved, nil
	case StateRejected:
		return StateRejected, nil
	case StateSuspended:
		return StateSuspended, nil
	default:
		return "", ErrInvalidReviewState
	}
}

func mapReviewQueueItem(row sqlc.ListCreatorRegistrationReviewCasesByStateRow) (ReviewQueueItem, error) {
	userID, err := postgres.UUIDFromPG(row.UserID)
	if err != nil {
		return ReviewQueueItem{}, fmt.Errorf("creator registration review queue user_id 変換に失敗しました: %w", err)
	}

	return ReviewQueueItem{
		CreatorBio: row.CreatorBio,
		LegalName:  row.LegalName,
		Review: ReviewTimeline{
			ApprovedAt:  postgres.OptionalTimeFromPG(row.ApprovedAt),
			RejectedAt:  postgres.OptionalTimeFromPG(row.RejectedAt),
			SubmittedAt: postgres.OptionalTimeFromPG(row.SubmittedAt),
			SuspendedAt: postgres.OptionalTimeFromPG(row.SuspendedAt),
		},
		SharedProfile: SharedProfilePreview{
			AvatarURL:   postgres.OptionalTextFromPG(row.AvatarUrl),
			DisplayName: row.DisplayName,
			Handle:      row.Handle,
			UserID:      userID,
		},
		State:  row.State,
		UserID: userID,
	}, nil
}

func buildReviewCase(snapshot registrationSnapshot, evidenceAccessURLs map[string]string) (ReviewCase, error) {
	if err := validateReviewCaseState(snapshot.capability); err != nil {
		return ReviewCase{}, err
	}

	evidences := make([]ReviewEvidence, 0, len(snapshot.evidences))
	for _, row := range snapshot.evidences {
		evidence, err := mapEvidence(row)
		if err != nil {
			return ReviewCase{}, err
		}

		accessURL, ok := evidenceAccessURLs[row.Kind]
		if !ok || strings.TrimSpace(accessURL) == "" {
			return ReviewCase{}, fmt.Errorf("creator registration review evidence access url が不足しています: kind=%s", row.Kind)
		}

		evidences = append(evidences, ReviewEvidence{
			Evidence:  evidence,
			AccessURL: accessURL,
		})
	}

	registration, err := buildRegistration(snapshot)
	if err != nil {
		return ReviewCase{}, err
	}

	intake := ReviewCaseIntake{
		AcceptsConsentResponsibility: snapshot.intake != nil && snapshot.intake.AcceptsConsentResponsibility,
		BirthDate:                    dateStringFromPG(optionalDate(snapshot.intake)),
		DeclaresNoProhibitedCategory: snapshot.intake != nil && snapshot.intake.DeclaresNoProhibitedCategory,
		LegalName:                    stringOrEmpty(snapshot.intake, func(row sqlc.AppCreatorRegistrationIntake) string { return row.LegalName }),
		PayoutRecipientName:          stringOrEmpty(snapshot.intake, func(row sqlc.AppCreatorRegistrationIntake) string { return row.PayoutRecipientName }),
		PayoutRecipientType:          optionalTextOrEmpty(snapshot.intake, func(row sqlc.AppCreatorRegistrationIntake) pgtype.Text { return row.PayoutRecipientType }),
	}

	return ReviewCase{
		CreatorBio:    creatorBioFromSnapshot(snapshot),
		Evidences:     evidences,
		Intake:        intake,
		Rejection:     registration.Rejection,
		Review:        registration.Review,
		SharedProfile: registration.SharedProfile,
		State:         registration.State,
		UserID:        registration.SharedProfile.UserID,
	}, nil
}

func (s *ReviewService) signEvidenceAccessURLs(
	ctx context.Context,
	evidences []sqlc.AppCreatorRegistrationEvidence,
) (map[string]string, error) {
	result := make(map[string]string, len(evidences))
	for _, evidence := range evidences {
		bucket := strings.TrimSpace(evidence.StorageBucket)
		key := strings.TrimSpace(evidence.StorageKey)
		if bucket == "" || key == "" {
			return nil, fmt.Errorf("creator registration review evidence storage location が不足しています: kind=%s", evidence.Kind)
		}

		accessURL, err := s.signEvidenceURL(ctx, bucket, key, s.evidenceAccessTTL)
		if err != nil {
			return nil, fmt.Errorf("creator registration review evidence access url 生成 kind=%s: %w", evidence.Kind, err)
		}
		result[evidence.Kind] = accessURL
	}

	return result, nil
}

func optionalDate(row *sqlc.AppCreatorRegistrationIntake) pgtype.Date {
	if row == nil {
		return pgtype.Date{}
	}

	return row.BirthDate
}
