package support

import "time"

const (
	APIDateTimeFormat = "2006-01-02T15:04:05Z"
	APIDateFormat     = "2006-01-02"

	DatabaseDateTimeFormat = "2006-01-02 15:04:05"
)

func APIDateTime(value time.Time) string {
	return value.UTC().Format(APIDateTimeFormat)
}

func APIDateTimePtr(value *time.Time) *string {
	if value == nil {
		return nil
	}

	formatted := APIDateTime(*value)

	return &formatted
}

func APIDate(value time.Time) string {
	return value.UTC().Format(APIDateFormat)
}

func DatabaseDateTime(value time.Time) string {
	return value.UTC().Format(DatabaseDateTimeFormat)
}
