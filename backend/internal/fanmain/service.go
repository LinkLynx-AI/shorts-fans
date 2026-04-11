package fanmain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/feed"
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

// Service は fan unlock / main playback 導線を扱います。
type Service struct {
	feedReader     feedReader
	mainReader     mainReader
	unlockRecorder unlockRecorder
	now            func() time.Time
	tokenTTL       time.Duration
	grantTTL       time.Duration
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
	// MainPlaybackGrantKindUnlocked は unlocked playback 向け grant です。
	MainPlaybackGrantKindUnlocked MainPlaybackGrantKind = "unlocked"
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

// UnlockSetupState は unlock setup 状態です。
type UnlockSetupState struct {
	Required                bool
	RequiresAgeConfirmation bool
	RequiresTermsAcceptance bool
}

// UnlockSurface は short ごとの unlock surface です。
type UnlockSurface struct {
	Access          MainAccessState
	Creator         CreatorSummary
	Main            MainSummary
	MainAccessToken string
	Setup           UnlockSetupState
	Short           ShortSummary
	UnlockCta       UnlockCtaState
}

// AccessEntryInput は main access entry 発行時の入力です。
type AccessEntryInput struct {
	AcceptedAge   bool
	AcceptedTerms bool
	EntryToken    string
	FromShortID   uuid.UUID
	MainID        uuid.UUID
	ViewerID      uuid.UUID
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
func NewService(feedReader feedReader, mainReader mainReader, unlockRecorder unlockRecorder) *Service {
	return &Service{
		feedReader:     feedReader,
		mainReader:     mainReader,
		unlockRecorder: unlockRecorder,
		now:            time.Now,
		tokenTTL:       defaultTokenTTL,
		grantTTL:       defaultGrantTTL,
	}
}

// GetUnlockSurface は short 起点の unlock surface を返します。
func (s *Service) GetUnlockSurface(ctx context.Context, viewerID uuid.UUID, sessionBinding string, shortID uuid.UUID) (UnlockSurface, error) {
	detail, main, err := s.loadLinkedSurface(ctx, viewerID, shortID)
	if err != nil {
		if errors.Is(err, feed.ErrPublicShortNotFound) || errors.Is(err, shorts.ErrUnlockableMainNotFound) {
			return UnlockSurface{}, ErrShortUnlockNotFound
		}

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

	return UnlockSurface{
		Access:          buildMainAccessState(detail.Item.Unlock, main.ID),
		Creator:         buildCreatorSummary(detail.Item.Creator),
		Main:            buildMainSummary(main, detail.Item.Unlock.MainDurationSeconds),
		MainAccessToken: entryToken,
		Setup: UnlockSetupState{
			Required:                false,
			RequiresAgeConfirmation: false,
			RequiresTermsAcceptance: false,
		},
		Short:     buildShortSummary(detail.Item.Short),
		UnlockCta: buildUnlockCtaState(detail.Item.Unlock),
	}, nil
}

// IssueAccessEntry は main playback grant を発行します。
func (s *Service) IssueAccessEntry(ctx context.Context, sessionBinding string, input AccessEntryInput) (AccessEntryResult, error) {
	detail, main, err := s.loadLinkedSurface(ctx, input.ViewerID, input.FromShortID)
	if err != nil {
		if errors.Is(err, feed.ErrPublicShortNotFound) || errors.Is(err, shorts.ErrUnlockableMainNotFound) {
			return AccessEntryResult{}, ErrAccessEntryNotFound
		}

		return AccessEntryResult{}, err
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

	grantKind := resolveGrantKind(detail.Item.Unlock)
	if grantKind == "" {
		return AccessEntryResult{}, ErrMainLocked
	}

	if err := s.recordMainUnlockIfNeeded(ctx, grantKind, input.ViewerID, input.MainID); err != nil {
		return AccessEntryResult{}, err
	}

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

	return AccessEntryResult{
		GrantKind:  grantKind,
		GrantToken: grantToken,
	}, nil
}

func (s *Service) recordMainUnlockIfNeeded(
	ctx context.Context,
	grantKind MainPlaybackGrantKind,
	viewerID uuid.UUID,
	mainID uuid.UUID,
) error {
	if grantKind != MainPlaybackGrantKindUnlocked {
		return nil
	}

	if s == nil || s.unlockRecorder == nil {
		return fmt.Errorf("fan main unlock recorder が初期化されていません")
	}

	_, err := s.unlockRecorder.RecordMainUnlock(ctx, unlock.RecordMainUnlockInput{
		UserID: viewerID,
		MainID: mainID,
	})
	if err == nil || errors.Is(err, unlock.ErrAlreadyUnlocked) {
		return nil
	}

	return fmt.Errorf("main unlock 記録 viewer=%s main=%s: %w", viewerID, mainID, err)
}

// GetPlaybackSurface は temporary grant を検証して main playback surface を返します。
func (s *Service) GetPlaybackSurface(ctx context.Context, viewerID uuid.UUID, sessionBinding string, mainID uuid.UUID, fromShortID uuid.UUID, grantToken string) (PlaybackSurface, error) {
	detail, main, err := s.loadLinkedSurface(ctx, viewerID, fromShortID)
	if err != nil {
		if errors.Is(err, feed.ErrPublicShortNotFound) || errors.Is(err, shorts.ErrUnlockableMainNotFound) {
			return PlaybackSurface{}, ErrPlaybackNotFound
		}

		return PlaybackSurface{}, err
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
		grant.ViewerID != viewerID {
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
		return MainAccessState{
			MainID: mainID,
			Reason: "owner_preview",
			Status: mainAccessOwner,
		}
	default:
		return MainAccessState{
			MainID: mainID,
			Reason: "session_unlocked",
			Status: mainAccessUnlocked,
		}
	}
}

func buildMainAccessState(source feed.UnlockPreview, mainID uuid.UUID) MainAccessState {
	switch {
	case source.IsOwner:
		return MainAccessState{
			MainID: mainID,
			Reason: "owner_preview",
			Status: mainAccessOwner,
		}
	case source.IsUnlocked:
		return MainAccessState{
			MainID: mainID,
			Reason: "session_unlocked",
			Status: mainAccessUnlocked,
		}
	default:
		return MainAccessState{
			MainID: mainID,
			Reason: "unlock_required",
			Status: mainAccessLocked,
		}
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

func buildUnlockCtaState(source feed.UnlockPreview) UnlockCtaState {
	switch {
	case source.IsOwner:
		return UnlockCtaState{
			State: "owner_preview",
		}
	case source.IsUnlocked:
		return UnlockCtaState{
			State: "continue_main",
		}
	default:
		mainDurationSeconds := source.MainDurationSeconds
		priceJPY := source.PriceJPY

		return UnlockCtaState{
			MainDurationSeconds: &mainDurationSeconds,
			PriceJPY:            &priceJPY,
			State:               "unlock_available",
		}
	}
}

func resolveGrantKind(source feed.UnlockPreview) MainPlaybackGrantKind {
	switch {
	case source.IsOwner:
		return MainPlaybackGrantKindOwner
	case source.IsUnlocked:
		return MainPlaybackGrantKindUnlocked
	default:
		return MainPlaybackGrantKindUnlocked
	}
}
