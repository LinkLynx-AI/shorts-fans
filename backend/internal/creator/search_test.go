package creator

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestNormalizeRequiredHandle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:  "trims leading at mark and lowercases",
			input: " @Mi.Na_1 ",
			want:  "mi.na_1",
		},
		{
			name:    "rejects empty handle",
			input:   "@",
			wantErr: ErrInvalidHandle,
		},
		{
			name:    "rejects invalid rune",
			input:   "alice-1",
			wantErr: ErrInvalidHandle,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := normalizeRequiredHandle(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("normalizeRequiredHandle(%q) error got %v want %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("normalizeRequiredHandle(%q) got %q want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetPublicProfileByHandle(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	userID := uuid.New()
	row := testPublicProfileRow(userID, now, stringPtr("Alice"), stringPtr("alice"), stringPtr("https://cdn.example.com/alice.jpg"), timePtr(now))
	var gotHandle string

	repo := newRepository(repositoryStubQueries{
		getPublicByHandle: func(_ context.Context, handle string) (sqlc.AppPublicCreatorProfile, error) {
			gotHandle = handle
			return row, nil
		},
	})

	profile, err := repo.GetPublicProfileByHandle(context.Background(), "@Alice")
	if err != nil {
		t.Fatalf("GetPublicProfileByHandle() error = %v, want nil", err)
	}
	if gotHandle != "alice" {
		t.Fatalf("GetPublicProfileByHandle() handle arg got %v want %v", gotHandle, "alice")
	}
	if profile.Handle == nil || *profile.Handle != "alice" {
		t.Fatalf("GetPublicProfileByHandle() handle got %#v want %q", profile.Handle, "alice")
	}
}

func TestGetPublicProfileByHandleNotFound(t *testing.T) {
	t.Parallel()

	repo := newRepository(repositoryStubQueries{
		getPublicByHandle: func(context.Context, string) (sqlc.AppPublicCreatorProfile, error) {
			return sqlc.AppPublicCreatorProfile{}, pgx.ErrNoRows
		},
	})

	if _, err := repo.GetPublicProfileByHandle(context.Background(), "missing"); !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("GetPublicProfileByHandle() error got %v want %v", err, ErrProfileNotFound)
	}
}

func TestListRecentPublicProfiles(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	userA := uuid.New()
	userB := uuid.New()
	userC := uuid.New()
	var gotParams sqlc.ListRecentPublicCreatorProfilesParams

	repo := newRepository(repositoryStubQueries{
		listRecentPublic: func(_ context.Context, arg sqlc.ListRecentPublicCreatorProfilesParams) ([]sqlc.AppPublicCreatorProfile, error) {
			gotParams = arg
			return []sqlc.AppPublicCreatorProfile{
				testPublicProfileRow(userA, now, stringPtr("Alice"), stringPtr("alice"), stringPtr("https://cdn.example.com/a.jpg"), timePtr(now)),
				testPublicProfileRow(userB, now.Add(-time.Hour), stringPtr("Bia"), stringPtr("bia"), stringPtr("https://cdn.example.com/b.jpg"), timePtr(now.Add(-time.Hour))),
				testPublicProfileRow(userC, now.Add(-2*time.Hour), stringPtr("Cora"), stringPtr("cora"), stringPtr("https://cdn.example.com/c.jpg"), timePtr(now.Add(-2*time.Hour))),
			}, nil
		},
	})

	profiles, nextCursor, err := repo.ListRecentPublicProfiles(context.Background(), nil, 2)
	if err != nil {
		t.Fatalf("ListRecentPublicProfiles() error = %v, want nil", err)
	}
	if gotParams.LimitCount != 3 {
		t.Fatalf("ListRecentPublicProfiles() limit got %d want %d", gotParams.LimitCount, 3)
	}
	if len(profiles) != 2 {
		t.Fatalf("ListRecentPublicProfiles() len got %d want %d", len(profiles), 2)
	}
	if nextCursor == nil {
		t.Fatal("ListRecentPublicProfiles() nextCursor = nil, want non-nil")
	}
	if nextCursor.Handle != "bia" {
		t.Fatalf("ListRecentPublicProfiles() next handle got %q want %q", nextCursor.Handle, "bia")
	}
}

func TestSearchPublicProfiles(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	userID := uuid.New()
	cursor := &PublicProfileCursor{
		PublishedAt: now,
		Handle:      "alice",
	}
	var gotParams sqlc.SearchPublicCreatorProfilesParams

	repo := newRepository(repositoryStubQueries{
		searchPublic: func(_ context.Context, arg sqlc.SearchPublicCreatorProfilesParams) ([]sqlc.AppPublicCreatorProfile, error) {
			gotParams = arg
			return []sqlc.AppPublicCreatorProfile{
				testPublicProfileRow(userID, now.Add(-time.Hour), stringPtr("Mina Rei"), stringPtr("minarei"), stringPtr("https://cdn.example.com/mina.jpg"), timePtr(now.Add(-time.Hour))),
			}, nil
		},
	})

	profiles, nextCursor, err := repo.SearchPublicProfiles(context.Background(), " @MiNa ", cursor, 2)
	if err != nil {
		t.Fatalf("SearchPublicProfiles() error = %v, want nil", err)
	}
	if gotParams.DisplayNameQuery.String != "@MiNa" {
		t.Fatalf("SearchPublicProfiles() display query got %q want %q", gotParams.DisplayNameQuery.String, "@MiNa")
	}
	if gotParams.HandlePrefixQuery != "mina" {
		t.Fatalf("SearchPublicProfiles() handle prefix got %#v want %q", gotParams.HandlePrefixQuery, "mina")
	}
	if gotParams.CursorHandle != pgText(stringPtr("alice")) {
		t.Fatalf("SearchPublicProfiles() cursor handle got %v want %v", gotParams.CursorHandle, pgText(stringPtr("alice")))
	}
	if len(profiles) != 1 {
		t.Fatalf("SearchPublicProfiles() len got %d want %d", len(profiles), 1)
	}
	if nextCursor != nil {
		t.Fatalf("SearchPublicProfiles() nextCursor got %#v want nil", nextCursor)
	}
}

func TestSearchPublicProfilesEscapesLikePattern(t *testing.T) {
	t.Parallel()

	var gotParams sqlc.SearchPublicCreatorProfilesParams

	repo := newRepository(repositoryStubQueries{
		searchPublic: func(_ context.Context, arg sqlc.SearchPublicCreatorProfilesParams) ([]sqlc.AppPublicCreatorProfile, error) {
			gotParams = arg
			return nil, nil
		},
	})

	if _, _, err := repo.SearchPublicProfiles(context.Background(), ` %a_b\c `, nil, 2); err != nil {
		t.Fatalf("SearchPublicProfiles() error = %v, want nil", err)
	}
	if gotParams.DisplayNameQuery.String != `\%a\_b\\c` {
		t.Fatalf("SearchPublicProfiles() escaped display query got %q want %q", gotParams.DisplayNameQuery.String, `\%a\_b\\c`)
	}
	if gotParams.HandlePrefixQuery != `a\_bc` {
		t.Fatalf("SearchPublicProfiles() escaped handle prefix got %#v want %q", gotParams.HandlePrefixQuery, `a\_bc`)
	}
}
