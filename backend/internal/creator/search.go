package creator

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const defaultSearchPageSize = 20

// PublicProfileCursor は公開 creator profile 一覧の keyset cursor です。
type PublicProfileCursor struct {
	PublishedAt time.Time
	Handle      string
}

// GetPublicProfileByHandle は handle から公開中の creator profile を取得します。
func (r *Repository) GetPublicProfileByHandle(ctx context.Context, handle string) (Profile, error) {
	normalizedHandle, err := normalizeRequiredHandle(handle)
	if err != nil {
		return Profile{}, fmt.Errorf("公開 creator profile 取得 handle 正規化: %w", err)
	}

	row, err := r.queries.GetPublicCreatorProfileByHandle(ctx, normalizedHandle)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Profile{}, fmt.Errorf("公開 creator profile 取得 handle=%s: %w", normalizedHandle, ErrProfileNotFound)
		}

		return Profile{}, fmt.Errorf("公開 creator profile 取得 handle=%s: %w", normalizedHandle, err)
	}

	profile, err := mapPublicProfile(row)
	if err != nil {
		return Profile{}, fmt.Errorf("公開 creator profile 取得結果の変換 handle=%s: %w", normalizedHandle, err)
	}

	return profile, nil
}

// ListRecentPublicProfiles は公開 creator profile の recent 一覧を返します。
func (r *Repository) ListRecentPublicProfiles(ctx context.Context, cursor *PublicProfileCursor, limit int) ([]Profile, *PublicProfileCursor, error) {
	params, pageLimit := buildPublicProfilePageParams(cursor, limit)

	rows, err := r.queries.ListRecentPublicCreatorProfiles(ctx, params)
	if err != nil {
		return nil, nil, fmt.Errorf("公開 creator profile recent 一覧取得: %w", err)
	}

	return mapPublicProfilePage(rows, pageLimit, "公開 creator profile recent 一覧取得結果の変換")
}

// SearchPublicProfiles は display name / handle のみを対象に公開 creator profile を検索します。
func (r *Repository) SearchPublicProfiles(ctx context.Context, query string, cursor *PublicProfileCursor, limit int) ([]Profile, *PublicProfileCursor, error) {
	trimmedQuery := strings.TrimSpace(query)
	pageParams, pageLimit := buildPublicProfilePageParams(cursor, limit)

	rows, err := r.queries.SearchPublicCreatorProfiles(ctx, sqlc.SearchPublicCreatorProfilesParams{
		DisplayNameQuery: pgtype.Text{
			String: escapeLikePattern(trimmedQuery),
			Valid:  true,
		},
		HandlePrefixQuery: escapeLikePattern(normalizeSearchHandleQuery(trimmedQuery)),
		CursorPublishedAt: pageParams.CursorPublishedAt,
		CursorHandle:      pageParams.CursorHandle,
		LimitCount:        pageParams.LimitCount,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("公開 creator profile 検索 query=%q: %w", trimmedQuery, err)
	}

	return mapPublicProfilePage(rows, pageLimit, fmt.Sprintf("公開 creator profile 検索結果の変換 query=%q", trimmedQuery))
}

func buildPublicProfilePageParams(cursor *PublicProfileCursor, limit int) (sqlc.ListRecentPublicCreatorProfilesParams, int) {
	if limit <= 0 {
		limit = defaultSearchPageSize
	}

	params := sqlc.ListRecentPublicCreatorProfilesParams{
		LimitCount: int32(limit + 1),
	}
	if cursor == nil {
		return params, limit
	}

	params.CursorPublishedAt = postgres.TimeToPG(&cursor.PublishedAt)
	params.CursorHandle = postgres.TextToPG(&cursor.Handle)

	return params, limit
}

func mapPublicProfilePage(rows []sqlc.AppPublicCreatorProfile, limit int, label string) ([]Profile, *PublicProfileCursor, error) {
	profiles := make([]Profile, 0, min(limit, len(rows)))
	for index, row := range rows {
		if index >= limit {
			break
		}

		profile, err := mapPublicProfile(row)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", label, err)
		}

		profiles = append(profiles, profile)
	}

	if len(rows) <= limit {
		return profiles, nil, nil
	}

	lastProfile := profiles[len(profiles)-1]
	if lastProfile.PublishedAt == nil || lastProfile.Handle == nil {
		return nil, nil, fmt.Errorf("%s: cursor 生成に必要な published_at または handle がありません", label)
	}

	return profiles, &PublicProfileCursor{
		PublishedAt: *lastProfile.PublishedAt,
		Handle:      *lastProfile.Handle,
	}, nil
}

func normalizeRequiredHandle(handle string) (string, error) {
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

func normalizeSearchHandleQuery(query string) string {
	var builder strings.Builder
	for _, char := range strings.TrimSpace(strings.ToLower(query)) {
		if char == '@' {
			continue
		}
		if isAllowedHandleRune(char) {
			builder.WriteRune(char)
		}
	}

	return builder.String()
}

func escapeLikePattern(value string) string {
	var builder strings.Builder
	for _, char := range value {
		if char == '%' || char == '_' || char == '\\' {
			builder.WriteRune('\\')
		}
		builder.WriteRune(char)
	}

	return builder.String()
}

func isAllowedHandleRune(char rune) bool {
	return unicode.IsDigit(char) || (char >= 'a' && char <= 'z') || char == '.' || char == '_'
}
