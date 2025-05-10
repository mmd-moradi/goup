package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// UUIDToPgUUID converts a uuid.UUID to a pgtype.UUID
func UUIDToPgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: id,
		Valid: true,
	}
}

// PgUUIDToUUID converts a pgtype.UUID to a uuid.UUID
func PgUUIDToUUID(id pgtype.UUID) uuid.UUID {
	if !id.Valid {
		return uuid.Nil
	}
	return uuid.UUID(id.Bytes)
}

// TimeToTimestamptz converts a time.Time to pgtype.Timestamptz
func TimeToTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  t,
		Valid: !t.IsZero(),
	}
}

// TimestamptzToTime converts a pgtype.Timestamptz to time.Time
func TimestamptzToTime(ts pgtype.Timestamptz) time.Time {
	if !ts.Valid {
		return time.Time{} // Zero time
	}
	return ts.Time
}
