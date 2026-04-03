package postgres

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// UUIDFromPG は null でない pgtype.UUID を uuid.UUID に変換します。
func UUIDFromPG(value pgtype.UUID) (uuid.UUID, error) {
	if !value.Valid {
		return uuid.Nil, fmt.Errorf("uuid が null です")
	}

	return uuid.UUID(value.Bytes), nil
}

// UUIDToPG は uuid.UUID を pgtype.UUID に変換します。
func UUIDToPG(value uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: [16]byte(value),
		Valid: true,
	}
}

// RequiredTimeFromPG は null でない pgtype.Timestamptz を time.Time に変換します。
func RequiredTimeFromPG(value pgtype.Timestamptz) (time.Time, error) {
	if !value.Valid {
		return time.Time{}, fmt.Errorf("timestamp が null です")
	}

	return value.Time, nil
}

// OptionalTimeFromPG は nullable な pgtype.Timestamptz を *time.Time に変換します。
func OptionalTimeFromPG(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}

	result := value.Time
	return &result
}

// TimeToPG は *time.Time を pgtype.Timestamptz に変換します。
func TimeToPG(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}

	return pgtype.Timestamptz{
		Time:  *value,
		Valid: true,
	}
}

// OptionalTextFromPG は nullable な pgtype.Text を *string に変換します。
func OptionalTextFromPG(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}

	result := value.String
	return &result
}

// TextToPG は *string を pgtype.Text に変換します。
func TextToPG(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}

	return pgtype.Text{
		String: *value,
		Valid:  true,
	}
}

// OptionalInt64FromPG は nullable な pgtype.Int8 を *int64 に変換します。
func OptionalInt64FromPG(value pgtype.Int8) *int64 {
	if !value.Valid {
		return nil
	}

	result := value.Int64
	return &result
}

// Int64ToPG は *int64 を pgtype.Int8 に変換します。
func Int64ToPG(value *int64) pgtype.Int8 {
	if value == nil {
		return pgtype.Int8{}
	}

	return pgtype.Int8{
		Int64: *value,
		Valid: true,
	}
}
