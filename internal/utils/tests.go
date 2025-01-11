package utils

import "time"

func MockGetCurrentTime() time.Time {
	return time.Date(2025, time.January, 1, 1, 0, 0, 0, time.UTC)
}

func MockGetCurrentTimeISO() string {
	return MockGetCurrentTime().Format(time.RFC3339)
}
