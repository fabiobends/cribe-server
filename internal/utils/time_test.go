package utils

import (
	"testing"
)

func TestUnixToISO(t *testing.T) {
	tests := []struct {
		name      string
		timestamp int64
		expected  string
	}{
		{
			name:      "Standard timestamp",
			timestamp: 1234567890,
			expected:  "2009-02-13T23:31:30Z",
		},
		{
			name:      "Zero timestamp (Unix epoch)",
			timestamp: 0,
			expected:  "1970-01-01T00:00:00Z",
		},
		{
			name:      "Recent timestamp",
			timestamp: 1701169200,
			expected:  "2023-11-28T11:00:00Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UnixToISO(tt.timestamp)
			if result != tt.expected {
				t.Errorf("UnixToISO(%d) = %s; expected %s", tt.timestamp, result, tt.expected)
			}
		})
	}
}
