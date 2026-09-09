package support

import "time"

// The API renders times as UTC ISO-8601 with a "Z" suffix and no sub-second
// precision, and dates as plain calendar dates.
const (
	APIDateTimeFormat = "2006-01-02T15:04:05Z"
	APIDateFormat     = "2006-01-02"

	// DatabaseDateTimeFormat is how datetimes are stored in SQLite.
	DatabaseDateTimeFormat = "2006-01-02 15:04:05"
)

// APIDateTime formats a time the way the API renders it.
func APIDateTime(value time.Time) string {
	return value.UTC().Format(APIDateTimeFormat)
}

// APIDateTimePtr formats a nullable time the way the API renders it.
func APIDateTimePtr(value *time.Time) *string {
	if value == nil {
		return nil
	}

	formatted := APIDateTime(*value)

	return &formatted
}

// APIDate formats a date the way the API renders it.
func APIDate(value time.Time) string {
	return value.UTC().Format(APIDateFormat)
}

// DatabaseDateTime formats a time for storage.
func DatabaseDateTime(value time.Time) string {
	return value.UTC().Format(DatabaseDateTimeFormat)
}
