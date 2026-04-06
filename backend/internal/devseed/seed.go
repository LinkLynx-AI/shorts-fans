package devseed

import (
	"context"
	"fmt"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/auth"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	creatorUserID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	fanUserID     = uuid.MustParse("22222222-2222-2222-2222-222222222222")

	mainID        = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	shortAID      = uuid.MustParse("44444444-4444-4444-4444-444444444444")
	shortBID      = uuid.MustParse("55555555-5555-5555-5555-555555555555")
	mainAssetID   = uuid.MustParse("66666666-6666-6666-6666-666666666666")
	shortAAssetID = uuid.MustParse("77777777-7777-7777-7777-777777777777")
	shortBAssetID = uuid.MustParse("88888888-8888-8888-8888-888888888888")

	creatorApprovedAt       = time.Date(2026, 1, 2, 9, 0, 0, 0, time.UTC)
	creatorPublishedAt      = time.Date(2026, 1, 2, 9, 30, 0, 0, time.UTC)
	mainApprovedAt          = time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)
	shortAApprovedAt        = time.Date(2026, 1, 2, 10, 30, 0, 0, time.UTC)
	shortAPublishedAt       = time.Date(2026, 1, 2, 11, 0, 0, 0, time.UTC)
	shortBApprovedAt        = time.Date(2026, 1, 2, 11, 30, 0, 0, time.UTC)
	shortBPublishedAt       = time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
	fanFollowedAt           = time.Date(2026, 1, 2, 12, 30, 0, 0, time.UTC)
	fanUnlockedAt           = time.Date(2026, 1, 2, 13, 0, 0, 0, time.UTC)
	fanPinnedShortAt        = time.Date(2026, 1, 2, 13, 30, 0, 0, time.UTC)
	fanSessionExpiresAt     = time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	creatorSessionExpiresAt = time.Date(2026, 12, 31, 0, 5, 0, 0, time.UTC)
)

const (
	creatorDisplayName = "Mika Aoi"
	creatorHandle      = "mikaaoi"
	creatorAvatarURL   = "https://cdn.example.com/mock/creator/avatar-mika-aoi.jpg"
	creatorBio         = "Public shorts から paid main へつながる creator mock profile."

	mainPriceMinor = int64(1200)
	mainCurrency   = "JPY"

	mainPurchaseRef     = "mock-purchase-main-001"
	fanSessionToken     = "dev-fan-session-token"
	creatorSessionToken = "dev-creator-session-token"
)

type mediaAssetSeed struct {
	id          uuid.UUID
	storageKey  string
	playbackURL string
	durationMS  int64
	externalRef *string
}

type shortSeed struct {
	id                   uuid.UUID
	mediaAssetID         uuid.UUID
	approvedForPublishAt time.Time
	publishedAt          time.Time
}

var mediaAssets = []mediaAssetSeed{
	{
		id:          mainAssetID,
		storageKey:  "mock/mains/mika-aoi-main.mp4",
		playbackURL: "https://cdn.example.com/mock/mains/mika-aoi-main.m3u8",
		durationMS:  182000,
	},
	{
		id:          shortAAssetID,
		storageKey:  "mock/shorts/mika-aoi-short-a.mp4",
		playbackURL: "https://cdn.example.com/mock/shorts/mika-aoi-short-a.m3u8",
		durationMS:  18000,
	},
	{
		id:          shortBAssetID,
		storageKey:  "mock/shorts/mika-aoi-short-b.mp4",
		playbackURL: "https://cdn.example.com/mock/shorts/mika-aoi-short-b.m3u8",
		durationMS:  21000,
	},
}

var publicShorts = []shortSeed{
	{
		id:                   shortAID,
		mediaAssetID:         shortAAssetID,
		approvedForPublishAt: shortAApprovedAt,
		publishedAt:          shortAPublishedAt,
	},
	{
		id:                   shortBID,
		mediaAssetID:         shortBAssetID,
		approvedForPublishAt: shortBApprovedAt,
		publishedAt:          shortBPublishedAt,
	},
}

// Summary は dev seed 適用後に利用しやすい主要 ID を返します。
type Summary struct {
	CreatorUserID       uuid.UUID
	FanUserID           uuid.UUID
	MainID              uuid.UUID
	ShortIDs            []uuid.UUID
	FanSessionToken     string
	CreatorSessionToken string
}

// Run はローカル開発用の固定 mock data を idempotent に投入します。
func Run(ctx context.Context, beginner postgres.TxBeginner) (Summary, error) {
	if beginner == nil {
		return Summary{}, fmt.Errorf("tx beginner が nil です")
	}

	if err := postgres.RunInTx(ctx, beginner, func(tx pgx.Tx) error {
		if err := upsertUser(ctx, tx, creatorUserID); err != nil {
			return err
		}
		if err := upsertUser(ctx, tx, fanUserID); err != nil {
			return err
		}
		if err := upsertCreatorCapability(ctx, tx); err != nil {
			return err
		}
		if err := upsertCreatorProfile(ctx, tx); err != nil {
			return err
		}
		for _, asset := range mediaAssets {
			if err := upsertMediaAsset(ctx, tx, asset); err != nil {
				return err
			}
		}
		if err := upsertMain(ctx, tx); err != nil {
			return err
		}
		for _, short := range publicShorts {
			if err := upsertShort(ctx, tx, short); err != nil {
				return err
			}
		}
		if err := upsertMainUnlock(ctx, tx); err != nil {
			return err
		}
		if err := upsertCreatorFollow(ctx, tx); err != nil {
			return err
		}
		if err := upsertPinnedShort(ctx, tx); err != nil {
			return err
		}
		if err := upsertAuthSession(ctx, tx, fanUserID, "fan", fanSessionToken, fanSessionExpiresAt); err != nil {
			return err
		}
		if err := upsertAuthSession(ctx, tx, creatorUserID, "creator", creatorSessionToken, creatorSessionExpiresAt); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return Summary{}, fmt.Errorf("dev seed 適用: %w", err)
	}

	shortIDs := make([]uuid.UUID, 0, len(publicShorts))
	for _, short := range publicShorts {
		shortIDs = append(shortIDs, short.id)
	}

	return Summary{
		CreatorUserID:       creatorUserID,
		FanUserID:           fanUserID,
		MainID:              mainID,
		ShortIDs:            shortIDs,
		FanSessionToken:     fanSessionToken,
		CreatorSessionToken: creatorSessionToken,
	}, nil
}

func upsertUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `
		INSERT INTO app.users (id)
		VALUES ($1)
		ON CONFLICT (id) DO NOTHING
	`, userID); err != nil {
		return fmt.Errorf("users upsert user_id=%s: %w", userID, err)
	}

	return nil
}

func upsertCreatorCapability(ctx context.Context, tx pgx.Tx) error {
	if _, err := tx.Exec(ctx, `
		INSERT INTO app.creator_capabilities (
			user_id,
			state,
			rejection_reason_code,
			is_resubmit_eligible,
			is_support_review_required,
			self_serve_resubmit_count,
			kyc_provider_case_ref,
			payout_provider_account_ref,
			submitted_at,
			approved_at,
			rejected_at,
			suspended_at
		) VALUES (
			$1,
			'approved',
			NULL,
			FALSE,
			FALSE,
			0,
			$2,
			$3,
			NULL,
			$4,
			NULL,
			NULL
		)
		ON CONFLICT (user_id) DO UPDATE
		SET
			state = EXCLUDED.state,
			rejection_reason_code = EXCLUDED.rejection_reason_code,
			is_resubmit_eligible = EXCLUDED.is_resubmit_eligible,
			is_support_review_required = EXCLUDED.is_support_review_required,
			self_serve_resubmit_count = EXCLUDED.self_serve_resubmit_count,
			kyc_provider_case_ref = EXCLUDED.kyc_provider_case_ref,
			payout_provider_account_ref = EXCLUDED.payout_provider_account_ref,
			submitted_at = EXCLUDED.submitted_at,
			approved_at = EXCLUDED.approved_at,
			rejected_at = EXCLUDED.rejected_at,
			suspended_at = EXCLUDED.suspended_at,
			updated_at = CURRENT_TIMESTAMP
	`, creatorUserID, "mock-kyc-case-mika-aoi", "mock-payout-account-mika-aoi", creatorApprovedAt); err != nil {
		return fmt.Errorf("creator_capabilities upsert user_id=%s: %w", creatorUserID, err)
	}

	return nil
}

func upsertCreatorProfile(ctx context.Context, tx pgx.Tx) error {
	if _, err := tx.Exec(ctx, `
		INSERT INTO app.creator_profiles (
			user_id,
			display_name,
			handle,
			avatar_url,
			bio,
			published_at
		) VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6
		)
		ON CONFLICT (user_id) DO UPDATE
		SET
			display_name = EXCLUDED.display_name,
			handle = EXCLUDED.handle,
			avatar_url = EXCLUDED.avatar_url,
			bio = EXCLUDED.bio,
			published_at = EXCLUDED.published_at,
			updated_at = CURRENT_TIMESTAMP
	`, creatorUserID, creatorDisplayName, creatorHandle, creatorAvatarURL, creatorBio, creatorPublishedAt); err != nil {
		return fmt.Errorf("creator_profiles upsert user_id=%s: %w", creatorUserID, err)
	}

	return nil
}

func upsertMediaAsset(ctx context.Context, tx pgx.Tx, asset mediaAssetSeed) error {
	if _, err := tx.Exec(ctx, `
		INSERT INTO app.media_assets (
			id,
			creator_user_id,
			processing_state,
			storage_provider,
			storage_bucket,
			storage_key,
			playback_url,
			mime_type,
			duration_ms,
			external_upload_ref
		) VALUES (
			$1,
			$2,
			'ready',
			's3',
			'mock-media-bucket',
			$3,
			$4,
			'video/mp4',
			$5,
			$6
		)
		ON CONFLICT (id) DO UPDATE
		SET
			processing_state = EXCLUDED.processing_state,
			storage_provider = EXCLUDED.storage_provider,
			storage_bucket = EXCLUDED.storage_bucket,
			storage_key = EXCLUDED.storage_key,
			playback_url = EXCLUDED.playback_url,
			mime_type = EXCLUDED.mime_type,
			duration_ms = EXCLUDED.duration_ms,
			external_upload_ref = EXCLUDED.external_upload_ref,
			updated_at = CURRENT_TIMESTAMP
	`, asset.id, creatorUserID, asset.storageKey, asset.playbackURL, asset.durationMS, asset.externalRef); err != nil {
		return fmt.Errorf("media_assets upsert asset_id=%s: %w", asset.id, err)
	}

	return nil
}

func upsertMain(ctx context.Context, tx pgx.Tx) error {
	if _, err := tx.Exec(ctx, `
		INSERT INTO app.mains (
			id,
			creator_user_id,
			media_asset_id,
			state,
			review_reason_code,
			post_report_state,
			price_minor,
			currency_code,
			ownership_confirmed,
			consent_confirmed,
			approved_for_unlock_at
		) VALUES (
			$1,
			$2,
			$3,
			'approved_for_unlock',
			NULL,
			NULL,
			$4,
			$5,
			TRUE,
			TRUE,
			$6
		)
		ON CONFLICT (id) DO UPDATE
		SET
			state = EXCLUDED.state,
			review_reason_code = EXCLUDED.review_reason_code,
			post_report_state = EXCLUDED.post_report_state,
			price_minor = EXCLUDED.price_minor,
			currency_code = EXCLUDED.currency_code,
			ownership_confirmed = EXCLUDED.ownership_confirmed,
			consent_confirmed = EXCLUDED.consent_confirmed,
			approved_for_unlock_at = EXCLUDED.approved_for_unlock_at,
			updated_at = CURRENT_TIMESTAMP
	`, mainID, creatorUserID, mainAssetID, mainPriceMinor, mainCurrency, mainApprovedAt); err != nil {
		return fmt.Errorf("mains upsert main_id=%s: %w", mainID, err)
	}

	return nil
}

func upsertShort(ctx context.Context, tx pgx.Tx, short shortSeed) error {
	if _, err := tx.Exec(ctx, `
		INSERT INTO app.shorts (
			id,
			creator_user_id,
			canonical_main_id,
			media_asset_id,
			state,
			review_reason_code,
			post_report_state,
			approved_for_publish_at,
			published_at
		) VALUES (
			$1,
			$2,
			$3,
			$4,
			'approved_for_publish',
			NULL,
			NULL,
			$5,
			$6
		)
		ON CONFLICT (id) DO UPDATE
		SET
			state = EXCLUDED.state,
			review_reason_code = EXCLUDED.review_reason_code,
			post_report_state = EXCLUDED.post_report_state,
			approved_for_publish_at = EXCLUDED.approved_for_publish_at,
			published_at = EXCLUDED.published_at,
			updated_at = CURRENT_TIMESTAMP
	`, short.id, creatorUserID, mainID, short.mediaAssetID, short.approvedForPublishAt, short.publishedAt); err != nil {
		return fmt.Errorf("shorts upsert short_id=%s: %w", short.id, err)
	}

	return nil
}

func upsertMainUnlock(ctx context.Context, tx pgx.Tx) error {
	if _, err := tx.Exec(ctx, `
		INSERT INTO app.main_unlocks (
			user_id,
			main_id,
			payment_provider_purchase_ref,
			purchased_at
		) VALUES (
			$1,
			$2,
			$3,
			$4
		)
		ON CONFLICT (user_id, main_id) DO UPDATE
		SET
			payment_provider_purchase_ref = EXCLUDED.payment_provider_purchase_ref,
			purchased_at = EXCLUDED.purchased_at
	`, fanUserID, mainID, mainPurchaseRef, fanUnlockedAt); err != nil {
		return fmt.Errorf("main_unlocks upsert user_id=%s main_id=%s: %w", fanUserID, mainID, err)
	}

	return nil
}

func upsertCreatorFollow(ctx context.Context, tx pgx.Tx) error {
	if _, err := tx.Exec(ctx, `
		INSERT INTO app.creator_follows (
			user_id,
			creator_user_id,
			followed_at
		) VALUES (
			$1,
			$2,
			$3
		)
		ON CONFLICT (user_id, creator_user_id) DO UPDATE
		SET
			followed_at = EXCLUDED.followed_at
	`, fanUserID, creatorUserID, fanFollowedAt); err != nil {
		return fmt.Errorf("creator_follows upsert user_id=%s creator_user_id=%s: %w", fanUserID, creatorUserID, err)
	}

	return nil
}

func upsertPinnedShort(ctx context.Context, tx pgx.Tx) error {
	if _, err := tx.Exec(ctx, `
		INSERT INTO app.pinned_shorts (
			user_id,
			short_id,
			pinned_at
		) VALUES (
			$1,
			$2,
			$3
		)
		ON CONFLICT (user_id, short_id) DO UPDATE
		SET
			pinned_at = EXCLUDED.pinned_at
	`, fanUserID, shortAID, fanPinnedShortAt); err != nil {
		return fmt.Errorf("pinned_shorts upsert user_id=%s short_id=%s: %w", fanUserID, shortAID, err)
	}

	return nil
}

func upsertAuthSession(
	ctx context.Context,
	tx pgx.Tx,
	userID uuid.UUID,
	activeMode string,
	rawSessionToken string,
	expiresAt time.Time,
) error {
	if _, err := tx.Exec(ctx, `
		INSERT INTO app.auth_sessions (
			user_id,
			active_mode,
			session_token_hash,
			expires_at
		) VALUES (
			$1,
			$2,
			$3,
			$4
		)
		ON CONFLICT (session_token_hash) DO UPDATE
		SET
			user_id = EXCLUDED.user_id,
			active_mode = EXCLUDED.active_mode,
			expires_at = EXCLUDED.expires_at,
			revoked_at = NULL,
			updated_at = CURRENT_TIMESTAMP
	`, userID, activeMode, auth.HashSessionToken(rawSessionToken), expiresAt); err != nil {
		return fmt.Errorf("auth_sessions upsert user_id=%s active_mode=%s: %w", userID, activeMode, err)
	}

	return nil
}
