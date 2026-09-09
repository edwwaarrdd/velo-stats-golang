package database

import (
	"database/sql"
	"fmt"
	"time"

	"velostats/internal/support"
)

// ParseDateTime reads a datetime column, which SQLite hands back either as the
// stored string or, for expressions, as a time.
func ParseDateTime(value any) (time.Time, error) {
	switch typed := value.(type) {
	case time.Time:
		return typed.UTC(), nil
	case string:
		for _, layout := range []string{support.DatabaseDateTimeFormat, time.RFC3339, "2006-01-02T15:04:05Z07:00", "2006-01-02"} {
			if parsed, err := time.ParseInLocation(layout, typed, time.UTC); err == nil {
				return parsed, nil
			}
		}

		return time.Time{}, fmt.Errorf("unrecognised datetime %q", typed)
	default:
		return time.Time{}, fmt.Errorf("unrecognised datetime of type %T", value)
	}
}

// NullTime converts a nullable datetime column into a time pointer.
func NullTime(value sql.NullString) (*time.Time, error) {
	if !value.Valid {
		return nil, nil
	}

	parsed, err := ParseDateTime(value.String)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}

// NullTimeValue stores a time pointer, writing NULL when it is absent.
func NullTimeValue(value *time.Time) any {
	if value == nil {
		return nil
	}

	return support.DatabaseDateTime(*value)
}

// FloatPointer converts a nullable float column into a float pointer.
func FloatPointer(value sql.NullFloat64) *float64 {
	if !value.Valid {
		return nil
	}

	return &value.Float64
}

// IntPointer converts a nullable integer column into an int pointer.
func IntPointer(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}

	return &value.Int64
}
