package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser_ParseTimestamp(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name      string
		timestamp string
		expected  float64
		wantError bool
	}{
		{
			name:      "Valid timestamp with decimals",
			timestamp: "1:23.45",
			expected:  83.45,
			wantError: false,
		},
		{
			name:      "Zero minutes",
			timestamp: "0:15.50",
			expected:  15.50,
			wantError: false,
		},
		{
			name:      "Multiple minutes",
			timestamp: "3:30.00",
			expected:  210.00,
			wantError: false,
		},
		{
			name:      "Invalid format - no colon",
			timestamp: "123.45",
			expected:  0,
			wantError: true,
		},
		{
			name:      "Invalid format - letters",
			timestamp: "1:abc",
			expected:  0,
			wantError: true,
		},
		{
			name:      "Empty string",
			timestamp: "",
			expected:  0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseTimestamp(tt.timestamp)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
