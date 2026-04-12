package fanprofile

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
)

// DefaultLibraryPageSize は fan profile library 一覧の既定 page size です。
const DefaultLibraryPageSize = DefaultPinnedShortsPageSize

// LibraryItem は private hub 上の unlocked main row を表します。
type LibraryItem struct {
	CreatorAvatarURL                 *string
	CreatorBio                       string
	CreatorDisplayName               string
	CreatorHandle                    string
	CreatorUserID                    uuid.UUID
	EntryShortCaption                string
	EntryShortCanonicalMainID        uuid.UUID
	EntryShortID                     uuid.UUID
	EntryShortMediaAssetID           uuid.UUID
	EntryShortPreviewDurationSeconds int64
	MainDurationSeconds              int64
	MainID                           uuid.UUID
	PurchasedAt                      time.Time
	UnlockCreatedAt                  time.Time
}

// LibraryCursor は library 一覧の keyset cursor です。
type LibraryCursor struct {
	MainID          uuid.UUID
	PurchasedAt     time.Time
	UnlockCreatedAt time.Time
}

// ListLibrary は authenticated viewer の library 一覧 1 page を返します。
func (r *Repository) ListLibrary(
	ctx context.Context,
	viewerID uuid.UUID,
	cursor *LibraryCursor,
	limit int,
) ([]LibraryItem, *LibraryCursor, error) {
	params, pageLimit := buildLibraryPageParams(viewerID, cursor, limit)

	rows, err := r.queries.ListFanProfileLibraryItems(ctx, params)
	if err != nil {
		return nil, nil, fmt.Errorf("fan profile library 取得 viewer=%s: %w", viewerID, err)
	}

	return mapLibraryPage(rows, pageLimit, fmt.Sprintf("fan profile library 取得結果の変換 viewer=%s", viewerID))
}

func buildLibraryPageParams(
	viewerID uuid.UUID,
	cursor *LibraryCursor,
	limit int,
) (sqlc.ListFanProfileLibraryItemsParams, int) {
	if limit <= 0 {
		limit = DefaultLibraryPageSize
	}

	params := sqlc.ListFanProfileLibraryItemsParams{
		UserID:     postgres.UUIDToPG(viewerID),
		LimitCount: int32(limit + 1),
	}
	if cursor == nil {
		return params, limit
	}

	params.CursorMainID = postgres.UUIDToPG(cursor.MainID)
	params.CursorPurchasedAt = postgres.TimeToPG(&cursor.PurchasedAt)
	params.CursorUnlockCreatedAt = postgres.TimeToPG(&cursor.UnlockCreatedAt)

	return params, limit
}

func mapLibraryPage(
	rows []sqlc.ListFanProfileLibraryItemsRow,
	limit int,
	label string,
) ([]LibraryItem, *LibraryCursor, error) {
	items := make([]LibraryItem, 0, min(limit, len(rows)))
	for index, row := range rows {
		if index >= limit {
			break
		}

		item, err := mapLibraryItem(row)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", label, err)
		}

		items = append(items, item)
	}

	if len(rows) <= limit {
		if len(items) == 0 {
			return []LibraryItem{}, nil, nil
		}

		return items, nil, nil
	}

	lastItem := items[len(items)-1]

	return items, &LibraryCursor{
		MainID:          lastItem.MainID,
		PurchasedAt:     lastItem.PurchasedAt,
		UnlockCreatedAt: lastItem.UnlockCreatedAt,
	}, nil
}

func mapLibraryItem(row sqlc.ListFanProfileLibraryItemsRow) (LibraryItem, error) {
	mainID, err := postgres.UUIDFromPG(row.MainID)
	if err != nil {
		return LibraryItem{}, fmt.Errorf("library main id 変換: %w", err)
	}
	creatorUserID, err := postgres.UUIDFromPG(row.CreatorUserID)
	if err != nil {
		return LibraryItem{}, fmt.Errorf("library creator user id 変換: %w", err)
	}
	entryShortID, err := postgres.UUIDFromPG(row.EntryShortID)
	if err != nil {
		return LibraryItem{}, fmt.Errorf("library entry short id 変換: %w", err)
	}
	entryShortCanonicalMainID, err := postgres.UUIDFromPG(row.EntryShortCanonicalMainID)
	if err != nil {
		return LibraryItem{}, fmt.Errorf("library entry short canonical main id 変換: %w", err)
	}
	entryShortMediaAssetID, err := postgres.UUIDFromPG(row.EntryShortMediaAssetID)
	if err != nil {
		return LibraryItem{}, fmt.Errorf("library entry short media asset id 変換: %w", err)
	}
	purchasedAt, err := postgres.RequiredTimeFromPG(row.PurchasedAt)
	if err != nil {
		return LibraryItem{}, fmt.Errorf("library purchased_at 変換: %w", err)
	}
	unlockCreatedAt, err := postgres.RequiredTimeFromPG(row.UnlockCreatedAt)
	if err != nil {
		return LibraryItem{}, fmt.Errorf("library unlock created_at 変換: %w", err)
	}

	if mainID != entryShortCanonicalMainID {
		return LibraryItem{}, fmt.Errorf("library entry short canonical main id が一致しません main=%s entry_short=%s", mainID, entryShortCanonicalMainID)
	}
	if !row.DisplayName.Valid || strings.TrimSpace(row.DisplayName.String) == "" {
		return LibraryItem{}, fmt.Errorf("library creator display_name がありません")
	}
	if strings.TrimSpace(row.Handle) == "" {
		return LibraryItem{}, fmt.Errorf("library creator handle がありません")
	}
	if !row.MainDurationMs.Valid || row.MainDurationMs.Int64 <= 0 {
		return LibraryItem{}, fmt.Errorf("library main duration_ms がありません")
	}
	if !row.EntryShortDurationMs.Valid || row.EntryShortDurationMs.Int64 <= 0 {
		return LibraryItem{}, fmt.Errorf("library entry short duration_ms がありません")
	}

	entryShortCaption := ""
	if row.EntryShortCaption.Valid {
		entryShortCaption = strings.TrimSpace(row.EntryShortCaption.String)
	}

	return LibraryItem{
		CreatorAvatarURL:                 postgres.OptionalTextFromPG(row.AvatarUrl),
		CreatorBio:                       row.Bio,
		CreatorDisplayName:               strings.TrimSpace(row.DisplayName.String),
		CreatorHandle:                    strings.TrimSpace(row.Handle),
		CreatorUserID:                    creatorUserID,
		EntryShortCaption:                entryShortCaption,
		EntryShortCanonicalMainID:        entryShortCanonicalMainID,
		EntryShortID:                     entryShortID,
		EntryShortMediaAssetID:           entryShortMediaAssetID,
		EntryShortPreviewDurationSeconds: (row.EntryShortDurationMs.Int64 + 999) / 1000,
		MainDurationSeconds:              (row.MainDurationMs.Int64 + 999) / 1000,
		MainID:                           mainID,
		PurchasedAt:                      purchasedAt,
		UnlockCreatedAt:                  unlockCreatedAt,
	}, nil
}
