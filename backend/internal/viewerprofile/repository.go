package viewerprofile

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

var (
	// ErrProfileNotFound は shared viewer profile が見つからないことを表します。
	ErrProfileNotFound = errors.New("viewer profile が見つかりません")
	// ErrCreatorProfileNotFound は creator mirror が見つからないことを表します。
	ErrCreatorProfileNotFound = errors.New("creator profile が見つかりません")
	// ErrInvalidDisplayName は shared display name が不正なことを表します。
	ErrInvalidDisplayName = errors.New("viewer display name が不正です")
	// ErrInvalidHandle は shared handle が不正なことを表します。
	ErrInvalidHandle = errors.New("viewer handle が不正です")
	// ErrHandleAlreadyTaken は handle が既に使われていることを表します。
	ErrHandleAlreadyTaken = errors.New("viewer handle は既に使われています")
)

const (
	userProfilesHandleUniqueConstraint    = "user_profiles_handle_unique_idx"
	creatorProfilesHandleUniqueConstraint = "creator_profiles_handle_unique_idx"
)

type queries interface {
	CreateUserProfile(ctx context.Context, arg sqlc.CreateUserProfileParams) (sqlc.AppUserProfile, error)
	GetCreatorProfileByUserID(ctx context.Context, userID pgtype.UUID) (sqlc.AppCreatorProfile, error)
	GetUserProfileByUserID(ctx context.Context, userID pgtype.UUID) (sqlc.AppUserProfile, error)
	UpdateCreatorProfile(ctx context.Context, arg sqlc.UpdateCreatorProfileParams) (sqlc.AppCreatorProfile, error)
	UpdateUserProfile(ctx context.Context, arg sqlc.UpdateUserProfileParams) (sqlc.AppUserProfile, error)
}

// HandleReservationReader は active sign-up の handle reservation を参照します。
type HandleReservationReader interface {
	IsHandleReserved(ctx context.Context, handle string) (bool, error)
}

// Repository は shared viewer profile 永続化をまとめます。
type Repository struct {
	txBeginner              postgres.TxBeginner
	queries                 queries
	newQueries              func(sqlc.DBTX) queries
	handleReservationReader HandleReservationReader
}

// Profile は shared viewer profile の domain model です。
type Profile struct {
	UserID      uuid.UUID
	DisplayName string
	Handle      string
	AvatarURL   *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CreateProfileInput は shared viewer profile 作成入力です。
type CreateProfileInput struct {
	AvatarURL   *string
	DisplayName string
	Handle      string
	UserID      uuid.UUID
}

// UpdateProfileInput は shared viewer profile 更新入力です。
type UpdateProfileInput struct {
	AvatarURL   *string
	DisplayName string
	Handle      string
	UserID      uuid.UUID
}

// NewRepository は pgxpool ベースの viewer profile repository を構築します。
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

// WithHandleReservationReader は sign-up 中の handle reservation checker を設定します。
func (r *Repository) WithHandleReservationReader(reader HandleReservationReader) *Repository {
	if r == nil {
		return nil
	}

	r.handleReservationReader = reader
	return r
}

// CreateProfile は shared viewer profile を作成します。
func (r *Repository) CreateProfile(ctx context.Context, input CreateProfileInput) (Profile, error) {
	if r == nil || r.queries == nil {
		return Profile{}, fmt.Errorf("viewer profile repository が初期化されていません")
	}

	displayName, handle, err := normalizeProfileInput(input.DisplayName, input.Handle)
	if err != nil {
		return Profile{}, err
	}
	if err := r.ensureHandleAvailable(ctx, handle); err != nil {
		return Profile{}, err
	}

	row, err := r.queries.CreateUserProfile(ctx, sqlc.CreateUserProfileParams{
		UserID:      postgres.UUIDToPG(input.UserID),
		DisplayName: displayName,
		Handle:      handle,
		AvatarUrl:   postgres.TextToPG(input.AvatarURL),
	})
	if err != nil {
		return Profile{}, fmt.Errorf("viewer profile 作成 user=%s: %w", input.UserID, mapProfileWriteError(err))
	}

	profile, err := mapProfile(row)
	if err != nil {
		return Profile{}, fmt.Errorf("viewer profile 作成結果の変換 user=%s: %w", input.UserID, err)
	}

	return profile, nil
}

// GetProfile は shared viewer profile を取得します。
func (r *Repository) GetProfile(ctx context.Context, userID uuid.UUID) (Profile, error) {
	if r == nil || r.queries == nil {
		return Profile{}, fmt.Errorf("viewer profile repository が初期化されていません")
	}

	row, err := r.queries.GetUserProfileByUserID(ctx, postgres.UUIDToPG(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Profile{}, fmt.Errorf("viewer profile 取得 user=%s: %w", userID, ErrProfileNotFound)
		}

		return Profile{}, fmt.Errorf("viewer profile 取得 user=%s: %w", userID, err)
	}

	profile, err := mapProfile(row)
	if err != nil {
		return Profile{}, fmt.Errorf("viewer profile 取得結果の変換 user=%s: %w", userID, err)
	}

	return profile, nil
}

// UpdateProfile は shared viewer profile を更新し、creator mirror が存在すれば同期します。
func (r *Repository) UpdateProfile(ctx context.Context, input UpdateProfileInput) (Profile, error) {
	return r.updateProfile(ctx, input, nil)
}

// UpdateCreatorProfileSync は shared viewer profile と creator mirror を同時更新します。
func (r *Repository) UpdateCreatorProfileSync(ctx context.Context, input UpdateProfileInput, creatorBio string) (Profile, error) {
	return r.updateProfile(ctx, input, &creatorBio)
}

func (r *Repository) updateProfile(ctx context.Context, input UpdateProfileInput, creatorBio *string) (Profile, error) {
	if r == nil || r.txBeginner == nil || r.newQueries == nil {
		return Profile{}, fmt.Errorf("viewer profile repository が初期化されていません")
	}

	displayName, handle, err := normalizeProfileInput(input.DisplayName, input.Handle)
	if err != nil {
		return Profile{}, err
	}

	var result Profile
	err = postgres.RunInTx(ctx, r.txBeginner, func(tx pgx.Tx) error {
		q := r.newQueries(tx)
		if err := r.ensureHandleAvailableForUpdate(ctx, q, input.UserID, handle); err != nil {
			return err
		}

		row, updateErr := q.UpdateUserProfile(ctx, sqlc.UpdateUserProfileParams{
			UserID:      postgres.UUIDToPG(input.UserID),
			DisplayName: displayName,
			Handle:      handle,
			AvatarUrl:   postgres.TextToPG(input.AvatarURL),
		})
		if updateErr != nil {
			if errors.Is(updateErr, pgx.ErrNoRows) {
				return fmt.Errorf("viewer profile 更新 user=%s: %w", input.UserID, ErrProfileNotFound)
			}

			return fmt.Errorf("viewer profile 更新 user=%s: %w", input.UserID, mapProfileWriteError(updateErr))
		}

		profile, mapErr := mapProfile(row)
		if mapErr != nil {
			return fmt.Errorf("viewer profile 更新結果の変換 user=%s: %w", input.UserID, mapErr)
		}

		existingCreatorProfile, creatorErr := q.GetCreatorProfileByUserID(ctx, postgres.UUIDToPG(input.UserID))
		if creatorErr != nil {
			if errors.Is(creatorErr, pgx.ErrNoRows) {
				if creatorBio != nil {
					return fmt.Errorf("creator mirror 更新 user=%s: %w", input.UserID, ErrCreatorProfileNotFound)
				}

				result = profile
				return nil
			}

			return fmt.Errorf("creator mirror 取得 user=%s: %w", input.UserID, creatorErr)
		}

		nextBio := existingCreatorProfile.Bio
		if creatorBio != nil {
			nextBio = *creatorBio
		}

		if _, creatorErr = q.UpdateCreatorProfile(ctx, sqlc.UpdateCreatorProfileParams{
			DisplayName: postgres.TextToPG(&displayName),
			Handle:      handle,
			AvatarUrl:   postgres.TextToPG(input.AvatarURL),
			Bio:         nextBio,
			UserID:      postgres.UUIDToPG(input.UserID),
		}); creatorErr != nil {
			if errors.Is(creatorErr, pgx.ErrNoRows) {
				return fmt.Errorf("creator mirror 更新 user=%s: %w", input.UserID, ErrCreatorProfileNotFound)
			}

			return fmt.Errorf("creator mirror 更新 user=%s: %w", input.UserID, mapProfileWriteError(creatorErr))
		}

		result = profile
		return nil
	})
	if err != nil {
		return Profile{}, err
	}

	return result, nil
}

func (r *Repository) ensureHandleAvailable(ctx context.Context, handle string) error {
	if r == nil || r.handleReservationReader == nil {
		return nil
	}

	reserved, err := r.handleReservationReader.IsHandleReserved(ctx, handle)
	if err != nil {
		return fmt.Errorf("viewer handle reservation 取得 handle=%s: %w", handle, err)
	}
	if reserved {
		return ErrHandleAlreadyTaken
	}

	return nil
}

func (r *Repository) ensureHandleAvailableForUpdate(ctx context.Context, q queries, userID uuid.UUID, handle string) error {
	if err := r.ensureHandleAvailable(ctx, handle); err == nil {
		return nil
	} else if !errors.Is(err, ErrHandleAlreadyTaken) {
		return err
	}

	if q == nil {
		return ErrHandleAlreadyTaken
	}

	currentProfile, getErr := q.GetUserProfileByUserID(ctx, postgres.UUIDToPG(userID))
	if getErr != nil {
		if errors.Is(getErr, pgx.ErrNoRows) {
			return ErrHandleAlreadyTaken
		}
		return fmt.Errorf("viewer profile 取得 user=%s: %w", userID, getErr)
	}

	if strings.TrimSpace(currentProfile.Handle) == handle {
		return nil
	}

	return ErrHandleAlreadyTaken
}

func normalizeProfileInput(displayName string, handle string) (string, string, error) {
	normalizedDisplayName := strings.TrimSpace(displayName)
	if normalizedDisplayName == "" {
		return "", "", ErrInvalidDisplayName
	}

	normalizedHandle := strings.TrimSpace(handle)
	normalizedHandle = strings.TrimPrefix(normalizedHandle, "@")
	normalizedHandle = strings.ToLower(normalizedHandle)
	if normalizedHandle == "" {
		return "", "", ErrInvalidHandle
	}

	for _, char := range normalizedHandle {
		if !isAllowedHandleRune(char) {
			return "", "", ErrInvalidHandle
		}
	}

	return normalizedDisplayName, normalizedHandle, nil
}

func isAllowedHandleRune(char rune) bool {
	return unicode.IsDigit(char) || (char >= 'a' && char <= 'z') || char == '.' || char == '_'
}

func mapProfileWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && (pgErr.ConstraintName == userProfilesHandleUniqueConstraint || pgErr.ConstraintName == creatorProfilesHandleUniqueConstraint) {
		return ErrHandleAlreadyTaken
	}

	return err
}

func mapProfile(row sqlc.AppUserProfile) (Profile, error) {
	userID, err := postgres.UUIDFromPG(row.UserID)
	if err != nil {
		return Profile{}, fmt.Errorf("viewer profile の user id 変換: %w", err)
	}
	createdAt, err := postgres.RequiredTimeFromPG(row.CreatedAt)
	if err != nil {
		return Profile{}, fmt.Errorf("viewer profile の created_at 変換: %w", err)
	}
	updatedAt, err := postgres.RequiredTimeFromPG(row.UpdatedAt)
	if err != nil {
		return Profile{}, fmt.Errorf("viewer profile の updated_at 変換: %w", err)
	}

	return Profile{
		UserID:      userID,
		DisplayName: row.DisplayName,
		Handle:      row.Handle,
		AvatarURL:   postgres.OptionalTextFromPG(row.AvatarUrl),
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}
