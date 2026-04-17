package creatorregistration

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	StateDraft     = "draft"
	StateSubmitted = "submitted"
	StateApproved  = "approved"
	StateRejected  = "rejected"
	StateSuspended = "suspended"

	PayoutRecipientTypeSelf     = "self"
	PayoutRecipientTypeBusiness = "business"

	EvidenceKindGovernmentID = "government_id"
	EvidenceKindPayoutProof  = "payout_proof"

	creatorProfilesHandleUniqueConstraint = "creator_profiles_handle_unique_idx"
)

var (
	ErrHandleAlreadyTaken        = errors.New("creator registration handle は既に使われています")
	ErrInvalidBirthDate          = errors.New("creator registration birth date が不正です")
	ErrInvalidDisplayName        = errors.New("creator registration display name が不正です")
	ErrInvalidHandle             = errors.New("creator registration handle が不正です")
	ErrInvalidLegalName          = errors.New("creator registration legal name が不正です")
	ErrInvalidPayoutRecipient    = errors.New("creator registration payout recipient が不正です")
	ErrInvalidPayoutRecipientTyp = errors.New("creator registration payout recipient type が不正です")
	ErrRegistrationIncomplete    = errors.New("creator registration intake が不足しています")
	ErrRegistrationStateConflict = errors.New("creator registration state conflict")
	ErrSharedProfileNotFound     = errors.New("shared viewer profile が見つかりません")
)

type queries interface {
	CreateCreatorCapability(ctx context.Context, arg sqlc.CreateCreatorCapabilityParams) (sqlc.AppCreatorCapability, error)
	CreateCreatorProfile(ctx context.Context, arg sqlc.CreateCreatorProfileParams) (sqlc.AppCreatorProfile, error)
	GetCreatorCapabilityByUserID(ctx context.Context, userID pgtype.UUID) (sqlc.AppCreatorCapability, error)
	GetCreatorCapabilityByUserIDForUpdate(ctx context.Context, userID pgtype.UUID) (sqlc.AppCreatorCapability, error)
	GetCreatorProfileByUserID(ctx context.Context, userID pgtype.UUID) (sqlc.AppCreatorProfile, error)
	GetCreatorRegistrationIntakeByUserID(ctx context.Context, userID pgtype.UUID) (sqlc.AppCreatorRegistrationIntake, error)
	GetUserProfileByUserID(ctx context.Context, userID pgtype.UUID) (sqlc.AppUserProfile, error)
	ListCreatorRegistrationEvidencesByUserID(ctx context.Context, userID pgtype.UUID) ([]sqlc.AppCreatorRegistrationEvidence, error)
	UpdateCreatorCapabilityState(ctx context.Context, arg sqlc.UpdateCreatorCapabilityStateParams) (sqlc.AppCreatorCapability, error)
	UpdateCreatorProfile(ctx context.Context, arg sqlc.UpdateCreatorProfileParams) (sqlc.AppCreatorProfile, error)
	UpsertCreatorRegistrationEvidence(ctx context.Context, arg sqlc.UpsertCreatorRegistrationEvidenceParams) (sqlc.AppCreatorRegistrationEvidence, error)
	UpsertCreatorRegistrationIntake(ctx context.Context, arg sqlc.UpsertCreatorRegistrationIntakeParams) (sqlc.AppCreatorRegistrationIntake, error)
}

// Repository は creator registration onboarding の永続化境界です。
type Repository struct {
	txBeginner postgres.TxBeginner
	queries    queries
	newQueries func(sqlc.DBTX) queries
}

type SharedProfilePreview struct {
	AvatarURL   *string
	DisplayName string
	Handle      string
	UserID      uuid.UUID
}

type CreatorDraft struct {
	Bio string
}

type ReviewTimeline struct {
	ApprovedAt  *time.Time
	RejectedAt  *time.Time
	SubmittedAt *time.Time
	SuspendedAt *time.Time
}

type Rejection struct {
	IsResubmitEligible      bool
	IsSupportReviewRequired bool
	ReasonCode              *string
	SelfServeResubmitCount  int32
	SelfServeResubmitRemain int32
}

type Surface struct {
	Kind             string
	WorkspacePreview *string
}

type Actions struct {
	CanEnterCreatorMode bool
	CanResubmit         bool
	CanSubmit           bool
}

type Registration struct {
	Actions       Actions
	CreatorDraft  CreatorDraft
	Rejection     *Rejection
	Review        ReviewTimeline
	SharedProfile SharedProfilePreview
	State         string
	Surface       Surface
}

type Evidence struct {
	FileName      string
	FileSizeBytes int64
	Kind          string
	MimeType      string
	UploadedAt    time.Time
}

type Intake struct {
	AcceptsConsentResponsibility bool
	BirthDate                    string
	CanSubmit                    bool
	CreatorBio                   string
	DeclaresNoProhibitedCategory bool
	Evidences                    []Evidence
	IsReadOnly                   bool
	LegalName                    string
	PayoutRecipientName          string
	PayoutRecipientType          string
	RegistrationState            *string
	SharedProfile                SharedProfilePreview
}

type SaveEvidenceInput struct {
	FileName      string
	FileSizeBytes int64
	Kind          string
	MimeType      string
	StorageBucket string
	StorageKey    string
	UploadedAt    time.Time
	UserID        uuid.UUID
}

type EvidenceStorageObject struct {
	Bucket string
	Key    string
}

type SaveEvidenceResult struct {
	Evidence       Evidence
	ReplacedObject *EvidenceStorageObject
}

type SaveIntakeInput struct {
	AcceptsConsentResponsibility bool
	BirthDate                    string
	CreatorBio                   string
	DeclaresNoProhibitedCategory bool
	LegalName                    string
	PayoutRecipientName          string
	PayoutRecipientType          string
	UserID                       uuid.UUID
}

type normalizedSaveIntakeInput struct {
	acceptsConsentResponsibility bool
	birthDate                    *time.Time
	creatorBio                   string
	declaresNoProhibitedCategory bool
	legalName                    string
	payoutRecipientName          string
	payoutRecipientType          *string
	userID                       uuid.UUID
}

type registrationSnapshot struct {
	capability     *sqlc.AppCreatorCapability
	creatorProfile *sqlc.AppCreatorProfile
	evidences      []sqlc.AppCreatorRegistrationEvidence
	intake         *sqlc.AppCreatorRegistrationIntake
	userProfile    sqlc.AppUserProfile
}

// NewRepository は pgxpool ベースの creator registration repository を構築します。
func NewRepository(pool *pgxpool.Pool) *Repository {
	if pool == nil {
		return &Repository{}
	}

	return &Repository{
		txBeginner: pool,
		queries:    sqlc.New(pool),
		newQueries: func(db sqlc.DBTX) queries {
			return sqlc.New(db)
		},
	}
}

func newRepository(q queries) *Repository {
	return &Repository{
		queries: q,
		newQueries: func(db sqlc.DBTX) queries {
			return sqlc.New(db)
		},
	}
}

// GetRegistration は current viewer の creator registration status を返します。
func (r *Repository) GetRegistration(ctx context.Context, userID uuid.UUID) (*Registration, error) {
	if r == nil || r.queries == nil {
		return nil, fmt.Errorf("creator registration repository が初期化されていません")
	}

	snapshot, err := r.loadSnapshot(ctx, r.queries, userID, false, false)
	if err != nil {
		return nil, err
	}
	if snapshot.capability == nil {
		return nil, nil
	}

	registration, err := buildRegistration(snapshot)
	if err != nil {
		return nil, err
	}

	return &registration, nil
}

// GetIntake は current viewer の editable onboarding intake を返します。
func (r *Repository) GetIntake(ctx context.Context, userID uuid.UUID) (Intake, error) {
	if r == nil || r.queries == nil {
		return Intake{}, fmt.Errorf("creator registration repository が初期化されていません")
	}

	snapshot, err := r.loadSnapshot(ctx, r.queries, userID, true, false)
	if err != nil {
		return Intake{}, err
	}

	return buildIntake(snapshot), nil
}

// PrepareEvidenceUpload は evidence upload 作成前に draft state を保証します。
func (r *Repository) PrepareEvidenceUpload(ctx context.Context, userID uuid.UUID) error {
	if r == nil || r.txBeginner == nil || r.newQueries == nil {
		return fmt.Errorf("creator registration repository が初期化されていません")
	}

	return postgres.RunInTx(ctx, r.txBeginner, func(tx pgx.Tx) error {
		q := r.newQueries(tx)
		snapshot, err := r.loadSnapshot(ctx, q, userID, false, true)
		if err != nil {
			return err
		}
		if err := ensureCapabilityEditable(snapshot.capability); err != nil {
			return err
		}
		_, err = ensureDraftCapability(ctx, q, userID, snapshot.capability)
		return err
	})
}

// SaveEvidence は completed evidence metadata を upsert します。
func (r *Repository) SaveEvidence(ctx context.Context, input SaveEvidenceInput) (SaveEvidenceResult, error) {
	if r == nil || r.txBeginner == nil || r.newQueries == nil {
		return SaveEvidenceResult{}, fmt.Errorf("creator registration repository が初期化されていません")
	}

	var result SaveEvidenceResult
	err := postgres.RunInTx(ctx, r.txBeginner, func(tx pgx.Tx) error {
		q := r.newQueries(tx)
		snapshot, loadErr := r.loadSnapshot(ctx, q, input.UserID, true, true)
		if loadErr != nil {
			return loadErr
		}
		if err := ensureCapabilityEditable(snapshot.capability); err != nil {
			return err
		}
		if _, err := ensureDraftCapability(ctx, q, input.UserID, snapshot.capability); err != nil {
			return err
		}
		result.ReplacedObject = findEvidenceStorageObject(snapshot.evidences, input.Kind, input.StorageBucket, input.StorageKey)

		row, err := q.UpsertCreatorRegistrationEvidence(ctx, sqlc.UpsertCreatorRegistrationEvidenceParams{
			UserID:        postgres.UUIDToPG(input.UserID),
			Kind:          input.Kind,
			FileName:      strings.TrimSpace(input.FileName),
			MimeType:      strings.TrimSpace(input.MimeType),
			FileSizeBytes: input.FileSizeBytes,
			StorageBucket: strings.TrimSpace(input.StorageBucket),
			StorageKey:    strings.TrimSpace(input.StorageKey),
			UploadedAt:    postgres.TimeToPG(&input.UploadedAt),
		})
		if err != nil {
			return fmt.Errorf("creator registration evidence 保存 user=%s kind=%s: %w", input.UserID, input.Kind, err)
		}

		mapped, err := mapEvidence(row)
		if err != nil {
			return err
		}
		result.Evidence = mapped
		return nil
	})
	if err != nil {
		return SaveEvidenceResult{}, err
	}

	return result, nil
}

// SaveIntake は draft creator registration intake を保存します。
func (r *Repository) SaveIntake(ctx context.Context, input SaveIntakeInput) (Intake, error) {
	if r == nil || r.txBeginner == nil || r.newQueries == nil {
		return Intake{}, fmt.Errorf("creator registration repository が初期化されていません")
	}

	normalized, err := normalizeSaveIntakeInput(input)
	if err != nil {
		return Intake{}, err
	}

	var intake Intake
	err = postgres.RunInTx(ctx, r.txBeginner, func(tx pgx.Tx) error {
		q := r.newQueries(tx)
		snapshot, loadErr := r.loadSnapshot(ctx, q, normalized.userID, true, true)
		if loadErr != nil {
			return loadErr
		}
		if err := ensureCapabilityEditable(snapshot.capability); err != nil {
			return err
		}
		if _, err := ensureDraftCapability(ctx, q, normalized.userID, snapshot.capability); err != nil {
			return err
		}
		if _, err := upsertDraftProfile(ctx, q, snapshot.userProfile, normalized.creatorBio, snapshot.creatorProfile); err != nil {
			return err
		}
		if _, err := q.UpsertCreatorRegistrationIntake(ctx, sqlc.UpsertCreatorRegistrationIntakeParams{
			UserID:                       postgres.UUIDToPG(normalized.userID),
			LegalName:                    normalized.legalName,
			BirthDate:                    dateToPG(normalized.birthDate),
			PayoutRecipientType:          postgres.TextToPG(normalized.payoutRecipientType),
			PayoutRecipientName:          normalized.payoutRecipientName,
			DeclaresNoProhibitedCategory: normalized.declaresNoProhibitedCategory,
			AcceptsConsentResponsibility: normalized.acceptsConsentResponsibility,
		}); err != nil {
			return fmt.Errorf("creator registration intake 保存 user=%s: %w", normalized.userID, err)
		}

		updatedSnapshot, err := r.loadSnapshot(ctx, q, normalized.userID, true, false)
		if err != nil {
			return err
		}
		intake = buildIntake(updatedSnapshot)
		return nil
	})
	if err != nil {
		return Intake{}, err
	}

	return intake, nil
}

// Submit は draft creator registration を submitted へ遷移します。
func (r *Repository) Submit(ctx context.Context, userID uuid.UUID) (Registration, error) {
	if r == nil || r.txBeginner == nil || r.newQueries == nil {
		return Registration{}, fmt.Errorf("creator registration repository が初期化されていません")
	}

	var registration Registration
	err := postgres.RunInTx(ctx, r.txBeginner, func(tx pgx.Tx) error {
		q := r.newQueries(tx)
		snapshot, err := r.loadSnapshot(ctx, q, userID, true, true)
		if err != nil {
			return err
		}
		if snapshot.capability == nil {
			return ErrRegistrationIncomplete
		}
		if !canCapabilitySubmit(snapshot.capability) {
			return ErrRegistrationStateConflict
		}
		if !isSnapshotComplete(snapshot) {
			return ErrRegistrationIncomplete
		}

		now := time.Now().UTC()
		nextResubmitCount := snapshot.capability.SelfServeResubmitCount
		if snapshot.capability.State == StateRejected {
			nextResubmitCount++
		}
		updatedCapability, err := q.UpdateCreatorCapabilityState(ctx, sqlc.UpdateCreatorCapabilityStateParams{
			State:                    StateSubmitted,
			RejectionReasonCode:      pgtype.Text{},
			IsResubmitEligible:       false,
			IsSupportReviewRequired:  false,
			SelfServeResubmitCount:   nextResubmitCount,
			KycProviderCaseRef:       snapshot.capability.KycProviderCaseRef,
			PayoutProviderAccountRef: snapshot.capability.PayoutProviderAccountRef,
			SubmittedAt:              postgres.TimeToPG(&now),
			ApprovedAt:               pgtype.Timestamptz{},
			RejectedAt:               pgtype.Timestamptz{},
			SuspendedAt:              pgtype.Timestamptz{},
			UserID:                   postgres.UUIDToPG(userID),
		})
		if err != nil {
			return fmt.Errorf("creator registration submit user=%s: %w", userID, err)
		}

		snapshot.capability = &updatedCapability
		registration, err = buildRegistration(snapshot)
		return err
	})
	if err != nil {
		return Registration{}, err
	}

	return registration, nil
}

func buildIntake(snapshot registrationSnapshot) Intake {
	var birthDate string
	if snapshot.intake != nil {
		birthDate = dateStringFromPG(snapshot.intake.BirthDate)
	}

	intake := Intake{
		AcceptsConsentResponsibility: snapshot.intake != nil && snapshot.intake.AcceptsConsentResponsibility,
		BirthDate:                    birthDate,
		CreatorBio:                   creatorBioFromSnapshot(snapshot),
		DeclaresNoProhibitedCategory: snapshot.intake != nil && snapshot.intake.DeclaresNoProhibitedCategory,
		Evidences:                    mapEvidenceList(snapshot.evidences),
		LegalName:                    stringOrEmpty(snapshot.intake, func(row sqlc.AppCreatorRegistrationIntake) string { return row.LegalName }),
		PayoutRecipientName:          stringOrEmpty(snapshot.intake, func(row sqlc.AppCreatorRegistrationIntake) string { return row.PayoutRecipientName }),
		PayoutRecipientType:          optionalTextOrEmpty(snapshot.intake, func(row sqlc.AppCreatorRegistrationIntake) pgtype.Text { return row.PayoutRecipientType }),
		SharedProfile:                buildSharedProfile(snapshot.userProfile),
	}
	if snapshot.capability != nil {
		intake.RegistrationState = &snapshot.capability.State
		intake.IsReadOnly = !isCapabilityEditable(snapshot.capability)
	}
	intake.CanSubmit = isCapabilityEditable(snapshot.capability) && isSnapshotComplete(snapshot)

	return intake
}

func buildRegistration(snapshot registrationSnapshot) (Registration, error) {
	if snapshot.capability == nil {
		return Registration{}, fmt.Errorf("registration snapshot に capability がありません")
	}

	registration := Registration{
		Actions: Actions{
			CanEnterCreatorMode: snapshot.capability.State == StateApproved,
			CanResubmit:         canCapabilitySelfServeResubmit(*snapshot.capability),
			CanSubmit:           snapshot.capability.State == StateDraft,
		},
		CreatorDraft: CreatorDraft{
			Bio: creatorBioFromSnapshot(snapshot),
		},
		Review: ReviewTimeline{
			ApprovedAt:  postgres.OptionalTimeFromPG(snapshot.capability.ApprovedAt),
			RejectedAt:  postgres.OptionalTimeFromPG(snapshot.capability.RejectedAt),
			SubmittedAt: postgres.OptionalTimeFromPG(snapshot.capability.SubmittedAt),
			SuspendedAt: postgres.OptionalTimeFromPG(snapshot.capability.SuspendedAt),
		},
		SharedProfile: buildSharedProfile(snapshot.userProfile),
		State:         snapshot.capability.State,
	}

	if snapshot.capability.State == StateApproved {
		registration.Surface = Surface{Kind: "creator_workspace"}
	} else {
		preview := "static_mock"
		registration.Surface = Surface{
			Kind:             "read_only_onboarding",
			WorkspacePreview: &preview,
		}
	}

	if snapshot.capability.State == StateRejected {
		registration.Rejection = &Rejection{
			IsResubmitEligible:      snapshot.capability.IsResubmitEligible,
			IsSupportReviewRequired: snapshot.capability.IsSupportReviewRequired,
			ReasonCode:              postgres.OptionalTextFromPG(snapshot.capability.RejectionReasonCode),
			SelfServeResubmitCount:  snapshot.capability.SelfServeResubmitCount,
			SelfServeResubmitRemain: maxInt32(0, 2-snapshot.capability.SelfServeResubmitCount),
		}
	}

	return registration, nil
}

func buildSharedProfile(profile sqlc.AppUserProfile) SharedProfilePreview {
	userID, err := postgres.UUIDFromPG(profile.UserID)
	if err != nil {
		userID = uuid.Nil
	}

	return SharedProfilePreview{
		AvatarURL:   postgres.OptionalTextFromPG(profile.AvatarUrl),
		DisplayName: profile.DisplayName,
		Handle:      profile.Handle,
		UserID:      userID,
	}
}

func creatorBioFromSnapshot(snapshot registrationSnapshot) string {
	if snapshot.creatorProfile == nil {
		return ""
	}

	return snapshot.creatorProfile.Bio
}

func ensureDraftCapability(ctx context.Context, q queries, userID uuid.UUID, capability *sqlc.AppCreatorCapability) (sqlc.AppCreatorCapability, error) {
	if capability != nil {
		return *capability, nil
	}

	row, err := q.CreateCreatorCapability(ctx, sqlc.CreateCreatorCapabilityParams{
		UserID:                   postgres.UUIDToPG(userID),
		State:                    StateDraft,
		RejectionReasonCode:      pgtype.Text{},
		IsResubmitEligible:       false,
		IsSupportReviewRequired:  false,
		SelfServeResubmitCount:   0,
		KycProviderCaseRef:       pgtype.Text{},
		PayoutProviderAccountRef: pgtype.Text{},
		SubmittedAt:              pgtype.Timestamptz{},
		ApprovedAt:               pgtype.Timestamptz{},
		RejectedAt:               pgtype.Timestamptz{},
		SuspendedAt:              pgtype.Timestamptz{},
	})
	if err != nil {
		return sqlc.AppCreatorCapability{}, fmt.Errorf("creator registration draft capability 作成 user=%s: %w", userID, err)
	}

	return row, nil
}

func ensureCapabilityEditable(capability *sqlc.AppCreatorCapability) error {
	if !isCapabilityEditable(capability) {
		return ErrRegistrationStateConflict
	}

	return nil
}

func (r *Repository) loadSnapshot(
	ctx context.Context,
	q queries,
	userID uuid.UUID,
	includeEvidences bool,
	lockCapability bool,
) (registrationSnapshot, error) {
	userProfile, err := q.GetUserProfileByUserID(ctx, postgres.UUIDToPG(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return registrationSnapshot{}, ErrSharedProfileNotFound
		}
		return registrationSnapshot{}, fmt.Errorf("creator registration shared profile 取得 user=%s: %w", userID, err)
	}

	snapshot := registrationSnapshot{userProfile: userProfile}

	capability, err := loadCapability(ctx, q, userID, lockCapability)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return registrationSnapshot{}, fmt.Errorf("creator registration capability 取得 user=%s: %w", userID, err)
		}
		return snapshot, nil
	}
	snapshot.capability = &capability

	creatorProfile, err := q.GetCreatorProfileByUserID(ctx, postgres.UUIDToPG(userID))
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return registrationSnapshot{}, fmt.Errorf("creator registration draft profile 取得 user=%s: %w", userID, err)
		}
	} else {
		snapshot.creatorProfile = &creatorProfile
	}

	intake, err := q.GetCreatorRegistrationIntakeByUserID(ctx, postgres.UUIDToPG(userID))
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return registrationSnapshot{}, fmt.Errorf("creator registration intake 取得 user=%s: %w", userID, err)
		}
	} else {
		snapshot.intake = &intake
	}

	if includeEvidences {
		evidences, err := q.ListCreatorRegistrationEvidencesByUserID(ctx, postgres.UUIDToPG(userID))
		if err != nil {
			return registrationSnapshot{}, fmt.Errorf("creator registration evidence 取得 user=%s: %w", userID, err)
		}
		snapshot.evidences = evidences
	}

	return snapshot, nil
}

func loadCapability(ctx context.Context, q queries, userID uuid.UUID, lockCapability bool) (sqlc.AppCreatorCapability, error) {
	userPGID := postgres.UUIDToPG(userID)
	if lockCapability {
		return q.GetCreatorCapabilityByUserIDForUpdate(ctx, userPGID)
	}

	return q.GetCreatorCapabilityByUserID(ctx, userPGID)
}

func isSnapshotComplete(snapshot registrationSnapshot) bool {
	if snapshot.capability == nil {
		return false
	}
	if strings.TrimSpace(snapshot.userProfile.DisplayName) == "" || strings.TrimSpace(snapshot.userProfile.Handle) == "" {
		return false
	}
	if strings.TrimSpace(creatorBioFromSnapshot(snapshot)) == "" {
		return false
	}
	if snapshot.intake == nil {
		return false
	}
	if strings.TrimSpace(snapshot.intake.LegalName) == "" {
		return false
	}
	if !snapshot.intake.BirthDate.Valid {
		return false
	}
	if !snapshot.intake.PayoutRecipientType.Valid || strings.TrimSpace(snapshot.intake.PayoutRecipientType.String) == "" {
		return false
	}
	if strings.TrimSpace(snapshot.intake.PayoutRecipientName) == "" {
		return false
	}
	if !snapshot.intake.DeclaresNoProhibitedCategory || !snapshot.intake.AcceptsConsentResponsibility {
		return false
	}

	kinds := make(map[string]struct{}, len(snapshot.evidences))
	for _, evidence := range snapshot.evidences {
		kinds[evidence.Kind] = struct{}{}
	}

	_, hasGovernmentID := kinds[EvidenceKindGovernmentID]
	_, hasPayoutProof := kinds[EvidenceKindPayoutProof]
	return hasGovernmentID && hasPayoutProof
}

func canCapabilitySelfServeResubmit(capability sqlc.AppCreatorCapability) bool {
	if capability.State != StateRejected {
		return false
	}
	if !capability.IsResubmitEligible || capability.IsSupportReviewRequired {
		return false
	}

	return maxInt32(0, 2-capability.SelfServeResubmitCount) > 0
}

func canCapabilitySubmit(capability *sqlc.AppCreatorCapability) bool {
	if capability == nil {
		return false
	}
	if capability.State == StateDraft {
		return true
	}

	return canCapabilitySelfServeResubmit(*capability)
}

func isCapabilityEditable(capability *sqlc.AppCreatorCapability) bool {
	if capability == nil {
		return true
	}
	if capability.State == StateDraft {
		return true
	}

	return canCapabilitySelfServeResubmit(*capability)
}

func mapCreatorProfileWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == creatorProfilesHandleUniqueConstraint {
		return ErrHandleAlreadyTaken
	}

	return err
}

func mapEvidence(row sqlc.AppCreatorRegistrationEvidence) (Evidence, error) {
	uploadedAt, err := postgres.RequiredTimeFromPG(row.UploadedAt)
	if err != nil {
		return Evidence{}, fmt.Errorf("creator registration evidence uploaded_at 変換に失敗しました: %w", err)
	}

	return Evidence{
		FileName:      row.FileName,
		FileSizeBytes: row.FileSizeBytes,
		Kind:          row.Kind,
		MimeType:      row.MimeType,
		UploadedAt:    uploadedAt,
	}, nil
}

func findEvidenceStorageObject(
	evidences []sqlc.AppCreatorRegistrationEvidence,
	kind string,
	currentBucket string,
	currentKey string,
) *EvidenceStorageObject {
	for _, evidence := range evidences {
		if evidence.Kind != kind {
			continue
		}

		bucket := strings.TrimSpace(evidence.StorageBucket)
		key := strings.TrimSpace(evidence.StorageKey)
		if bucket == "" || key == "" {
			return nil
		}
		if bucket == strings.TrimSpace(currentBucket) && key == strings.TrimSpace(currentKey) {
			return nil
		}

		return &EvidenceStorageObject{
			Bucket: bucket,
			Key:    key,
		}
	}

	return nil
}

func mapEvidenceList(rows []sqlc.AppCreatorRegistrationEvidence) []Evidence {
	if len(rows) == 0 {
		return []Evidence{}
	}

	result := make([]Evidence, 0, len(rows))
	for _, row := range rows {
		evidence, err := mapEvidence(row)
		if err != nil {
			continue
		}
		result = append(result, evidence)
	}

	return result
}

func normalizeHandle(handle string) (string, error) {
	normalized := strings.TrimSpace(handle)
	normalized = strings.TrimPrefix(normalized, "@")
	normalized = strings.ToLower(normalized)
	if normalized == "" {
		return "", ErrInvalidHandle
	}
	for _, char := range normalized {
		if !isAllowedHandleRune(char) {
			return "", ErrInvalidHandle
		}
	}
	return normalized, nil
}

func normalizeSaveIntakeInput(input SaveIntakeInput) (normalizedSaveIntakeInput, error) {
	legalName := strings.TrimSpace(input.LegalName)
	if input.LegalName != "" && legalName == "" {
		return normalizedSaveIntakeInput{}, ErrInvalidLegalName
	}

	creatorBio := strings.TrimSpace(input.CreatorBio)
	var birthDate *time.Time
	if strings.TrimSpace(input.BirthDate) != "" {
		parsed, err := parseBirthDate(input.BirthDate)
		if err != nil {
			return normalizedSaveIntakeInput{}, err
		}
		birthDate = &parsed
	}

	var payoutRecipientType *string
	if strings.TrimSpace(input.PayoutRecipientType) != "" {
		normalizedType := strings.TrimSpace(input.PayoutRecipientType)
		switch normalizedType {
		case PayoutRecipientTypeSelf, PayoutRecipientTypeBusiness:
			payoutRecipientType = &normalizedType
		default:
			return normalizedSaveIntakeInput{}, ErrInvalidPayoutRecipientTyp
		}
	}

	payoutRecipientName := strings.TrimSpace(input.PayoutRecipientName)
	if input.PayoutRecipientName != "" && payoutRecipientName == "" {
		return normalizedSaveIntakeInput{}, ErrInvalidPayoutRecipient
	}

	return normalizedSaveIntakeInput{
		acceptsConsentResponsibility: input.AcceptsConsentResponsibility,
		birthDate:                    birthDate,
		creatorBio:                   creatorBio,
		declaresNoProhibitedCategory: input.DeclaresNoProhibitedCategory,
		legalName:                    legalName,
		payoutRecipientName:          payoutRecipientName,
		payoutRecipientType:          payoutRecipientType,
		userID:                       input.UserID,
	}, nil
}

func parseBirthDate(value string) (time.Time, error) {
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, ErrInvalidBirthDate
	}

	return parsed.UTC(), nil
}

func upsertDraftProfile(
	ctx context.Context,
	q queries,
	userProfile sqlc.AppUserProfile,
	creatorBio string,
	existing *sqlc.AppCreatorProfile,
) (sqlc.AppCreatorProfile, error) {
	displayName := strings.TrimSpace(userProfile.DisplayName)
	if displayName == "" {
		return sqlc.AppCreatorProfile{}, ErrInvalidDisplayName
	}
	handle, err := normalizeHandle(userProfile.Handle)
	if err != nil {
		return sqlc.AppCreatorProfile{}, err
	}
	avatarURL := postgres.OptionalTextFromPG(userProfile.AvatarUrl)

	if existing == nil {
		row, err := q.CreateCreatorProfile(ctx, sqlc.CreateCreatorProfileParams{
			UserID:      postgres.UUIDToPG(uuid.UUID(userProfile.UserID.Bytes)),
			DisplayName: postgres.TextToPG(&displayName),
			Handle:      handle,
			AvatarUrl:   postgres.TextToPG(avatarURL),
			Bio:         creatorBio,
			PublishedAt: pgtype.Timestamptz{},
		})
		if err != nil {
			return sqlc.AppCreatorProfile{}, mapCreatorProfileWriteError(err)
		}
		return row, nil
	}

	row, err := q.UpdateCreatorProfile(ctx, sqlc.UpdateCreatorProfileParams{
		DisplayName: postgres.TextToPG(&displayName),
		Handle:      handle,
		AvatarUrl:   postgres.TextToPG(avatarURL),
		Bio:         creatorBio,
		UserID:      postgres.UUIDToPG(uuid.UUID(userProfile.UserID.Bytes)),
	})
	if err != nil {
		return sqlc.AppCreatorProfile{}, mapCreatorProfileWriteError(err)
	}

	return row, nil
}

func dateStringFromPG(value pgtype.Date) string {
	if !value.Valid {
		return ""
	}

	return value.Time.Format("2006-01-02")
}

func dateToPG(value *time.Time) pgtype.Date {
	if value == nil {
		return pgtype.Date{}
	}

	return pgtype.Date{
		Time:  *value,
		Valid: true,
	}
}

func isAllowedHandleRune(char rune) bool {
	return unicode.IsDigit(char) || (char >= 'a' && char <= 'z') || char == '.' || char == '_'
}

func maxInt32(a int32, b int32) int32 {
	if a > b {
		return a
	}

	return b
}

func optionalTextOrEmpty[T any](value *T, extractor func(T) pgtype.Text) string {
	if value == nil {
		return ""
	}

	resolved := extractor(*value)
	if !resolved.Valid {
		return ""
	}

	return resolved.String
}

func stringOrEmpty[T any](value *T, extractor func(T) string) string {
	if value == nil {
		return ""
	}

	return extractor(*value)
}
