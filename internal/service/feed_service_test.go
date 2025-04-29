package service

import (
	"testing"
	"time"
)

func TestParseDate(t *testing.T) {
	// Define test cases
	tests := []struct {
		name     string
		input    string
		expected time.Time
		wantErr  bool
	}{
		{
			name:     "valid date",
			input:    "Mon, 01 Jan 2024 00:00:00 +0000",
			expected: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "valid date with GMT",
			input:    "Mon, 28 Apr 2025 13:03:04 GMT",
			expected: time.Date(2025, 4, 28, 13, 3, 4, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "invalid date format",
			input:    "2024-01-01",
			expected: time.Time{},
			wantErr:  true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: time.Time{},
			wantErr:  true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDate(tt.input)

			if tt.wantErr {
				if !got.IsZero() {
					t.Errorf("parseDate() = %v, want empty time.Time for error case", got)
				}
				return
			}

			if !got.Equal(tt.expected) {
				t.Errorf("parseDate() = %v, want %v", got, tt.expected)
			}
		})
	}
}
