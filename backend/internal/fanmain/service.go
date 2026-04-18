package fanmain

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/feed"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/payment"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/shorts"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/unlock"
	"github.com/google/uuid"
)

const (
	entryTokenKind     = "entry"
	playbackTokenKind  = "playback"
	defaultTokenTTL    = 15 * time.Minute
	defaultGrantTTL    = 15 * time.Minute
	mainAccessLocked   = "locked"
	mainAccessOwner    = "owner"
	mainAccessUnlocked = "unlocked"
)

// ErrAccessEntryNotFound は access entry 発行対象の short/main が見つからないことを表します。
var ErrAccessEntryNotFound = errors.New("main or short が見つかりません")

// ErrPlaybackNotFound は playback 対象の main が見つからないことを表します。
var ErrPlaybackNotFound = errors.New("main が見つかりません")

// ErrShortUnlockNotFound は unlock surface 対象の short が見つからないことを表します。
var ErrShortUnlockNotFound = errors.New("short が見つかりません")

// ErrMainLocked は main access を発行できないことを表します。
var ErrMainLocked = errors.New("main はまだ再生できません")

type feedReader interface {
	GetDetail(ctx context.Context, shortID uuid.UUID, viewerUserID *uuid.UUID) (feed.Detail, error)
}

type mainReader interface {
	GetUnlockableMain(ctx context.Context, id uuid.UUID) (shorts.Main, error)
}

type unlockRecorder interface {
	RecordMainUnlock(ctx context.Context, input unlock.RecordMainUnlockInput) (unlock.MainUnlock, error)
}

type unlockConversionRecorder interface {
	RecordUnlockConversion(ctx context.Context, viewerID uuid.UUID, detail feed.Detail, idempotencyKey string) error
}

type unlockConversionRetryStore interface {
	ClearPendingUnlockConversion(ctx context.Context, viewerID uuid.UUID, mainID uuid.UUID, shortID uuid.UUID) error
	HasPendingUnlockConversion(ctx context.Context, viewerID uuid.UUID, mainID uuid.UUID, shortID uuid.UUID) (bool, error)
	MarkPendingUnlockConversion(ctx context.Context, viewerID uuid.UUID, mainID uuid.UUID, shortID uuid.UUID) error
}

// Service は fan unlock / main playback / purchase 導線を扱います。
type Service struct {
	feedReader        feedReader
	mainReader        mainReader
	unlockRecorder    unlockRecorder
	paymentRepository paymentRepository
	purchaseGateway   purchaseGateway

	unlockConversionRecorder unlockConversionRecorder
	unlockConversionRetry    unlockConversionRetryStore

	now      func() time.Time
	tokenTTL time.Duration
	grantTTL time.Duration
}

// CreatorSummary は unlock / playback surface で使う creator 表示情報です。
type CreatorSummary struct {
	AvatarURL   *string
	Bio         string
	DisplayName string
	Handle      string
	ID          uuid.UUID
}

// MainAccessState は main access 状態を表します。
type MainAccessState struct {
	MainID uuid.UUID
	Reason string
	Status string
}

// MainPlaybackGrantKind は temporary grant の種別です。
type MainPlaybackGrantKind string

const (
	// MainPlaybackGrantKindOwner は owner preview 向け grant です。
	MainPlaybackGrantKindOwner MainPlaybackGrantKind = "owner"
	// MainPlaybackGrantKindPurchased は purchased playback 向け grant です。
	MainPlaybackGrantKindPurchased MainPlaybackGrantKind = "purchased"
)

// MainSummary は unlock / playback surface で使う main 表示情報です。
type MainSummary struct {
	DurationSeconds int64
	ID              uuid.UUID
	MediaAssetID    uuid.UUID
	PriceJPY        int64
}

// ShortSummary は unlock / playback surface で使う short 表示情報です。
type ShortSummary struct {
	Caption                string
	CanonicalMainID        uuid.UUID
	CreatorUserID          uuid.UUID
	ID                     uuid.UUID
	MediaAssetID           uuid.UUID
	PreviewDurationSeconds int64
}

// UnlockCtaState は unlock CTA の表示状態です。
type UnlockCtaState struct {
	MainDurationSeconds   *int64
	PriceJPY              *int64
	ResumePositionSeconds *int64
	State                 string
}

// UnlockSurface は short ごとの unlock surface です。
type UnlockSurface struct {
	Access          MainAccessState
	Creator         CreatorSummary
	Main            MainSummary
	MainAccessToken string
	Purchase        UnlockPurchaseState
	Short           ShortSummary
	UnlockCta       UnlockCtaState
}

// AccessEntryInput は main access entry 発行時の入力です。
type AccessEntryInput struct {
	EntryToken  string
	FromShortID uuid.UUID
	MainID      uuid.UUID
	ViewerID    uuid.UUID
}

// AccessEntryResult は発行済み playback grant を表します。
type AccessEntryResult struct {
	GrantKind  MainPlaybackGrantKind
	GrantToken string
}

// PlaybackSurface は main playback surface です。
type PlaybackSurface struct {
	Access                MainAccessState
	Creator               CreatorSummary
	EntryShort            ShortSummary
	Main                  MainSummary
	ResumePositionSeconds *int64
}

// NewService は fan unlock / main playback service を構築します。
func NewService(
	feedReader feedReader,
	mainReader mainReader,
	unlockRecorder unlockRecorder,
	paymentRepository paymentRepository,
	purchaseGateway purchaseGateway,
) *Service {
	return &Service{
		feedReader:        feedReader,
		mainReader:        mainReader,
		unlockRecorder:    unlockRecorder,
		paymentRepository: paymentRepository,
		purchaseGateway:   purchaseGateway,
		now:               time.Now,
		tokenTTL:          defaultTokenTTL,
		grantTTL:          defaultGrantTTL,
	}
}

// WithRecommendationRecorder は unlock conversion signal recorder を注入します。
func (s *Service) WithRecommendationRecorder(recorder unlockConversionRecorder) *Service {
	if s == nil {
		return nil
	}

	s.unlockConversionRecorder = recorder

	return s
}

// WithUnlockConversionRetryStore は unlock conversion retry state store を注入します。
func (s *Service) WithUnlockConversionRetryStore(store unlockConversionRetryStore) *Service {
	if s == nil {
		return nil
	}

	s.unlockConversionRetry = store

	return s
}

// GetUnlockSurface は short 起点の unlock surface を返します。
func (s *Service) GetUnlockSurface(ctx context.Context, viewerID uuid.UUID, sessionBinding string, shortID uuid.UUID) (UnlockSurface, error) {
	detail, main, err := s.loadLinkedSurface(ctx, viewerID, shortID)
	if err != nil {
		switch {
		case errors.Is(err, feed.ErrPublicShortNotFound):
			return UnlockSurface{}, ErrShortUnlockNotFound
		case errors.Is(err, shorts.ErrUnlockableMainNotFound):
			return UnlockSurface{}, ErrMainLocked
		default:
			return UnlockSurface{}, err
		}
	}

	savedMethods, err := s.listSavedPaymentMethods(ctx, viewerID)
	if err != nil {
		return UnlockSurface{}, err
	}

	var inflightAttempt *payment.MainPurchaseAttempt
	attempt, err := s.getInflightAttempt(ctx, viewerID, main.ID)
	switch {
	case err == nil:
		inflightAttempt = &attempt
	case errors.Is(err, payment.ErrMainPurchaseAttemptNotFound):
		inflightAttempt = nil
	default:
		return UnlockSurface{}, err
	}

	entryToken, err := issueSignedToken(sessionBinding, s.now().UTC(), s.tokenTTL, signedTokenPayload{
		Kind:        entryTokenKind,
		MainID:      main.ID,
		FromShortID: detail.Item.Short.ID,
		ViewerID:    viewerID,
	})
	if err != nil {
		return UnlockSurface{}, err
	}

	purchaseState := buildUnlockPurchaseState(detail.Item.Unlock, savedMethods, inflightAttempt)

	return UnlockSurface{
		Access:          buildMainAccessState(detail.Item.Unlock, main.ID),
		Creator:         buildCreatorSummary(detail.Item.Creator),
		Main:            buildMainSummary(main, detail.Item.Unlock.MainDurationSeconds),
		MainAccessToken: entryToken,
		Purchase:        purchaseState,
		Short:           buildShortSummary(detail.Item.Short),
		UnlockCta:       buildUnlockCtaState(detail.Item.Unlock, purchaseState),
	}, nil
}

// IssueAccessEntry は main playback grant を発行します。
func (s *Service) IssueAccessEntry(ctx context.Context, sessionBinding string, input AccessEntryInput) (AccessEntryResult, error) {
	detail, main, err := s.loadLinkedSurface(ctx, input.ViewerID, input.FromShortID)
	if err != nil {
		switch {
		case errors.Is(err, feed.ErrPublicShortNotFound):
			return AccessEntryResult{}, ErrAccessEntryNotFound
		case errors.Is(err, shorts.ErrUnlockableMainNotFound):
			return AccessEntryResult{}, ErrMainLocked
		default:
			return AccessEntryResult{}, err
		}
	}

	if detail.Item.Short.CanonicalMainID != input.MainID || main.ID != input.MainID {
		return AccessEntryResult{}, ErrAccessEntryNotFound
	}

	entryToken, err := readSignedToken(sessionBinding, s.now().UTC(), input.EntryToken)
	if err != nil {
		return AccessEntryResult{}, ErrMainLocked
	}

	if entryToken.Kind != entryTokenKind ||
		entryToken.MainID != input.MainID ||
		entryToken.FromShortID != input.FromShortID ||
		entryToken.ViewerID != input.ViewerID {
		return AccessEntryResult{}, ErrMainLocked
	}

	grantKind, err := s.resolveAccessEntryGrantKind(ctx, detail.Item.Unlock, input.ViewerID, input.MainID)
	if err != nil {
		return AccessEntryResult{}, err
	}
	if grantKind == "" {
		return AccessEntryResult{}, ErrMainLocked
	}

	unlockRequired := grantKind == MainPlaybackGrantKindPurchased && !detail.Item.Unlock.IsUnlocked
	s.markPendingUnlockConversionIfNeeded(ctx, unlockRequired, input.ViewerID, input.MainID, detail.Item.Short.ID)
	grantToken, err := issueSignedToken(sessionBinding, s.now().UTC(), s.grantTTL, signedTokenPayload{
		GrantKind:   grantKind,
		Kind:        playbackTokenKind,
		MainID:      input.MainID,
		FromShortID: input.FromShortID,
		ViewerID:    input.ViewerID,
	})
	if err != nil {
		return AccessEntryResult{}, err
	}

	s.recordUnlockConversionIfNeeded(ctx, unlockRequired, detail, input.ViewerID, input.MainID)

	return AccessEntryResult{
		GrantKind:  grantKind,
		GrantToken: grantToken,
	}, nil
}

func (s *Service) recordUnlockConversionIfNeeded(
	ctx context.Context,
	unlockRequired bool,
	detail feed.Detail,
	viewerID uuid.UUID,
	mainID uuid.UUID,
) {
	if s == nil || s.unlockConversionRecorder == nil {
		return
	}
	shouldRecord := unlockRequired
	if !shouldRecord {
		pending, err := s.hasPendingUnlockConversion(ctx, viewerID, mainID, detail.Item.Short.ID)
		if err != nil || !pending {
			return
		}

		shouldRecord = true
	}
	if !shouldRecord {
		return
	}

	idempotencyKey := fmt.Sprintf(
		"unlock-conversion:%s:%s:%s",
		strings.TrimSpace(viewerID.String()),
		strings.TrimSpace(mainID.String()),
		strings.TrimSpace(detail.Item.Short.ID.String()),
	)

	// Recommendation signal failure must not block the access-entry UX.
	if err := s.unlockConversionRecorder.RecordUnlockConversion(ctx, viewerID, detail, idempotencyKey); err != nil {
		return
	}

	s.clearPendingUnlockConversion(ctx, viewerID, mainID, detail.Item.Short.ID)
}

func (s *Service) resolveAccessEntryGrantKind(
	ctx context.Context,
	preview feed.UnlockPreview,
	viewerID uuid.UUID,
	mainID uuid.UUID,
) (MainPlaybackGrantKind, error) {
	grantKind := resolveGrantKind(preview)
	if grantKind != "" {
		return grantKind, nil
	}

	if _, err := s.getLatestSucceededAttempt(ctx, viewerID, mainID); err == nil {
		return MainPlaybackGrantKindPurchased, nil
	} else if !errors.Is(err, payment.ErrMainPurchaseAttemptNotFound) {
		return "", err
	}

	return "", nil
}

// GetPlaybackSurface は temporary grant を検証して main playback surface を返します。
func (s *Service) GetPlaybackSurface(ctx context.Context, viewerID uuid.UUID, sessionBinding string, mainID uuid.UUID, fromShortID uuid.UUID, grantToken string) (PlaybackSurface, error) {
	detail, main, err := s.loadLinkedSurface(ctx, viewerID, fromShortID)
	if err != nil {
		switch {
		case errors.Is(err, feed.ErrPublicShortNotFound):
			return PlaybackSurface{}, ErrPlaybackNotFound
		case errors.Is(err, shorts.ErrUnlockableMainNotFound):
			return PlaybackSurface{}, ErrMainLocked
		default:
			return PlaybackSurface{}, err
		}
	}

	if detail.Item.Short.CanonicalMainID != mainID || main.ID != mainID {
		return PlaybackSurface{}, ErrPlaybackNotFound
	}

	grant, err := readSignedToken(sessionBinding, s.now().UTC(), grantToken)
	if err != nil {
		return PlaybackSurface{}, ErrMainLocked
	}

	if grant.Kind != playbackTokenKind ||
		grant.MainID != mainID ||
		grant.FromShortID != fromShortID ||
		grant.ViewerID != viewerID ||
		grant.GrantKind == "" {
		return PlaybackSurface{}, ErrMainLocked
	}

	return PlaybackSurface{
		Access:                buildGrantedAccessState(mainID, grant.GrantKind),
		Creator:               buildCreatorSummary(detail.Item.Creator),
		EntryShort:            buildShortSummary(detail.Item.Short),
		Main:                  buildMainSummary(main, detail.Item.Unlock.MainDurationSeconds),
		ResumePositionSeconds: nil,
	}, nil
}

func (s *Service) loadLinkedSurface(ctx context.Context, viewerID uuid.UUID, shortID uuid.UUID) (feed.Detail, shorts.Main, error) {
	if s == nil || s.feedReader == nil || s.mainReader == nil {
		return feed.Detail{}, shorts.Main{}, fmt.Errorf("fan main service が初期化されていません")
	}

	detail, err := s.feedReader.GetDetail(ctx, shortID, &viewerID)
	if err != nil {
		return feed.Detail{}, shorts.Main{}, err
	}

	main, err := s.mainReader.GetUnlockableMain(ctx, detail.Item.Short.CanonicalMainID)
	if err != nil {
		return feed.Detail{}, shorts.Main{}, err
	}

	return detail, main, nil
}

func buildCreatorSummary(source feed.CreatorSummary) CreatorSummary {
	return CreatorSummary{
		AvatarURL:   source.AvatarURL,
		Bio:         source.Bio,
		DisplayName: source.DisplayName,
		Handle:      source.Handle,
		ID:          source.ID,
	}
}

func buildGrantedAccessState(mainID uuid.UUID, grantKind MainPlaybackGrantKind) MainAccessState {
	switch grantKind {
	case MainPlaybackGrantKindOwner:
		return buildOwnerAccessState(mainID)
	default:
		return buildPurchasedAccessState(mainID)
	}
}

func buildMainAccessState(source feed.UnlockPreview, mainID uuid.UUID) MainAccessState {
	switch {
	case source.IsOwner:
		return buildOwnerAccessState(mainID)
	case source.IsUnlocked:
		return buildPurchasedAccessState(mainID)
	default:
		return buildLockedAccessState(mainID)
	}
}

func buildPurchasedAccessState(mainID uuid.UUID) MainAccessState {
	return MainAccessState{
		MainID: mainID,
		Reason: "purchased",
		Status: mainAccessUnlocked,
	}
}

func buildOwnerAccessState(mainID uuid.UUID) MainAccessState {
	return MainAccessState{
		MainID: mainID,
		Reason: "owner_preview",
		Status: mainAccessOwner,
	}
}

func buildLockedAccessState(mainID uuid.UUID) MainAccessState {
	return MainAccessState{
		MainID: mainID,
		Reason: "unlock_required",
		Status: mainAccessLocked,
	}
}

func buildMainSummary(source shorts.Main, durationSeconds int64) MainSummary {
	return MainSummary{
		DurationSeconds: durationSeconds,
		ID:              source.ID,
		MediaAssetID:    source.MediaAssetID,
		PriceJPY:        source.PriceMinor,
	}
}

func buildShortSummary(source feed.ShortSummary) ShortSummary {
	return ShortSummary{
		Caption:                source.Caption,
		CanonicalMainID:        source.CanonicalMainID,
		CreatorUserID:          source.CreatorUserID,
		ID:                     source.ID,
		MediaAssetID:           source.MediaAssetID,
		PreviewDurationSeconds: source.PreviewDurationSeconds,
	}
}

func buildUnlockCtaState(source feed.UnlockPreview, purchase UnlockPurchaseState) UnlockCtaState {
	switch purchase.State {
	case "owner_preview":
		return UnlockCtaState{State: "owner_preview"}
	case "already_purchased":
		return UnlockCtaState{State: "continue_main"}
	case "setup_required":
		return UnlockCtaState{
			MainDurationSeconds: int64Ptr(source.MainDurationSeconds),
			PriceJPY:            int64Ptr(source.PriceJPY),
			State:               "setup_required",
		}
	case "purchase_pending", "purchase_ready":
		return UnlockCtaState{
			MainDurationSeconds: int64Ptr(source.MainDurationSeconds),
			PriceJPY:            int64Ptr(source.PriceJPY),
			State:               "unlock_available",
		}
	default:
		return UnlockCtaState{
			MainDurationSeconds: int64Ptr(source.MainDurationSeconds),
			PriceJPY:            int64Ptr(source.PriceJPY),
			State:               "unlock_available",
		}
	}
}

func resolveGrantKind(source feed.UnlockPreview) MainPlaybackGrantKind {
	switch {
	case source.IsOwner:
		return MainPlaybackGrantKindOwner
	case source.IsUnlocked:
		return MainPlaybackGrantKindPurchased
	default:
		return ""
	}
}

func (s *Service) markPendingUnlockConversionIfNeeded(ctx context.Context, unlockRequired bool, viewerID uuid.UUID, mainID uuid.UUID, shortID uuid.UUID) {
	if !unlockRequired || s == nil || s.unlockConversionRetry == nil {
		return
	}

	_ = s.unlockConversionRetry.MarkPendingUnlockConversion(ctx, viewerID, mainID, shortID)
}

func (s *Service) hasPendingUnlockConversion(ctx context.Context, viewerID uuid.UUID, mainID uuid.UUID, shortID uuid.UUID) (bool, error) {
	if s == nil || s.unlockConversionRetry == nil {
		return false, nil
	}

	return s.unlockConversionRetry.HasPendingUnlockConversion(ctx, viewerID, mainID, shortID)
}

func (s *Service) clearPendingUnlockConversion(ctx context.Context, viewerID uuid.UUID, mainID uuid.UUID, shortID uuid.UUID) {
	if s == nil || s.unlockConversionRetry == nil {
		return
	}

	_ = s.unlockConversionRetry.ClearPendingUnlockConversion(ctx, viewerID, mainID, shortID)
}

func int64Ptr(value int64) *int64 {
	return &value
}

func stringPtr(value string) *string {
	return &value
}
