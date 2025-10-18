package utils

import "time"

// UnixToISO converts a Unix timestamp to ISO 8601 string format
func UnixToISO(timestamp int64) string {
	return time.Unix(timestamp, 0).UTC().Format(time.RFC3339)
}
