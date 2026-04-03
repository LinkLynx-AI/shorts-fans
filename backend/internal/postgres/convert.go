package postgres

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// UUIDFromPG converts a non-null pgtype.UUID into uuid.UUID.
func UUIDFromPG(value pgtype.UUID) (uuid.UUID, error) {
	if !value.Valid {
		return uuid.Nil, fmt.Errorf("uuid is null")
	}

	return uuid.UUID(value.Bytes), nil
}

// UUIDToPG converts uuid.UUID into pgtype.UUID.
func UUIDToPG(value uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: [16]byte(value),
		Valid: true,
	}
}

// RequiredTimeFromPG converts a non-null pgtype.Timestamptz into time.Time.
func RequiredTimeFromPG(value pgtype.Timestamptz) (time.Time, error) {
	if !value.Valid {
		return time.Time{}, fmt.Errorf("timestamp is null")
	}

	return value.Time, nil
}

// OptionalTimeFromPG converts a nullable pgtype.Timestamptz into *time.Time.
func OptionalTimeFromPG(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}

	result := value.Time
	return &result
}

// TimeToPG converts *time.Time into pgtype.Timestamptz.
func TimeToPG(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}

	return pgtype.Timestamptz{
		Time:  *value,
		Valid: true,
	}
}

// OptionalTextFromPG converts a nullable pgtype.Text into *string.
func OptionalTextFromPG(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}

	result := value.String
	return &result
}

// TextToPG converts *string into pgtype.Text.
func TextToPG(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}

	return pgtype.Text{
		String: *value,
		Valid:  true,
	}
}

// OptionalInt64FromPG converts a nullable pgtype.Int8 into *int64.
func OptionalInt64FromPG(value pgtype.Int8) *int64 {
	if !value.Valid {
		return nil
	}

	result := value.Int64
	return &result
}

// Int64ToPG converts *int64 into pgtype.Int8.
func Int64ToPG(value *int64) pgtype.Int8 {
	if value == nil {
		return pgtype.Int8{}
	}

	return pgtype.Int8{
		Int64: *value,
		Valid: true,
	}
}
