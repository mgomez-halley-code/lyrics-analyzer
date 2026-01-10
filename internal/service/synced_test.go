package service

import (
	"testing"

	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestParser_ParseSyncedLyrics(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		shouldError   bool
		expectedLines int
		validate      func(t *testing.T, lines []model.LyricLine)
	}{
		{
			name: "basic synced lyrics",
			input: `
				[00:10.00] First line
				[00:15.50] Second line
				[00:20.00] Third line
			`,
			shouldError:   false,
			expectedLines: 3,
			validate: func(t *testing.T, lines []model.LyricLine) {
				// Check first line
				assert.Equal(t, 1, lines[0].LineNumber)
				assert.Equal(t, "First line", lines[0].Text)
				assert.Equal(t, 10.0, *lines[0].Seconds)
				assert.Equal(t, 2, lines[0].WordCount)

				// Check second line
				assert.Equal(t, 2, lines[1].LineNumber)
				assert.Equal(t, "Second line", lines[1].Text)
				assert.Equal(t, 15.5, *lines[1].Seconds)
				assert.Equal(t, 2, lines[1].WordCount)

				// Check third line
				assert.Equal(t, 3, lines[2].LineNumber)
				assert.Equal(t, "Third line", lines[2].Text)
				assert.Equal(t, 20.0, *lines[2].Seconds)
				assert.Equal(t, 2, lines[2].WordCount)
			},
		},
		{
			name:          "empty synced lyrics",
			input:         "",
			shouldError:   true,
			expectedLines: 0,
		},
		{
			name: "malformed lines skipped",
			input: `
				[00:10.00] Valid line
				This line has no timestamp
				[00:20.00] Another valid line
				[invalid] Bad timestamp
			`,
			shouldError:   false,
			expectedLines: 2,
			validate: func(t *testing.T, lines []model.LyricLine) {
				assert.Equal(t, "Valid line", lines[0].Text)
				assert.Equal(t, "Another valid line", lines[1].Text)
			},
		},
		{
			name: "empty text with timestamp - skipped",
			input: `
				[00:10.00]
				[00:15.00] Valid line
			`,
			shouldError:   false,
			expectedLines: 1,
			validate: func(t *testing.T, lines []model.LyricLine) {
				// Empty text lines should be skipped by regex
				assert.Equal(t, "Valid line", lines[0].Text)
			},
		},
		{
			name: "extra whitespace around text",
			input: `
				[00:10.00]   Trimmed text
				[00:15.00]	Tab trimmed
			`,
			shouldError:   false,
			expectedLines: 2,
			validate: func(t *testing.T, lines []model.LyricLine) {
				assert.Equal(t, "Trimmed text", lines[0].Text)
				assert.Equal(t, "Tab trimmed", lines[1].Text)
			},
		},
		{
			name: "single word vs multi-word count",
			input: `
				[00:10.00] Word
				[00:15.00] Multiple words here
			`,
			shouldError:   false,
			expectedLines: 2,
			validate: func(t *testing.T, lines []model.LyricLine) {
				assert.Equal(t, 1, lines[0].WordCount)
				assert.Equal(t, 3, lines[1].WordCount)
			},
		},
		{
			name: "line numbers sequential",
			input: `
				[00:10.00] Line one
				[00:15.00] Line two
				[00:20.00] Line three
				[00:25.00] Line four
			`,
			shouldError:   false,
			expectedLines: 4,
			validate: func(t *testing.T, lines []model.LyricLine) {
				for i, line := range lines {
					assert.Equal(t, i+1, line.LineNumber)
				}
			},
		},
		{
			name: "timestamps not in order",
			input: `
				[00:20.00] Second
				[00:10.00] First
				[00:15.00] Third
			`,
			shouldError:   false,
			expectedLines: 3,
			validate: func(t *testing.T, lines []model.LyricLine) {
				// Parser should preserve order, not sort
				assert.Equal(t, 20.0, *lines[0].Seconds)
				assert.Equal(t, 10.0, *lines[1].Seconds)
				assert.Equal(t, 15.0, *lines[2].Seconds)
			},
		},
		{
			name: "large timestamp values",
			input: `
				[99:59.99] End of song
				[100:00.00] After 100 minutes
			`,
			shouldError:   false,
			expectedLines: 2,
			validate: func(t *testing.T, lines []model.LyricLine) {
				assert.Equal(t, 5999.99, *lines[0].Seconds)
				assert.Equal(t, 6000.0, *lines[1].Seconds)
			},
		},
		{
			name: "special characters in text",
			input: `
				[00:10.00] Hello! How's it going?
				[00:15.00] C'est la vie & rock 'n' roll
				[00:20.00] 100% awesome!!!
			`,
			shouldError:   false,
			expectedLines: 3,
			validate: func(t *testing.T, lines []model.LyricLine) {
				assert.Equal(t, "Hello! How's it going?", lines[0].Text)
				assert.Equal(t, "C'est la vie & rock 'n' roll", lines[1].Text)
				assert.Equal(t, "100% awesome!!!", lines[2].Text)
			},
		},
		{
			name: "blank lines between lyrics",
			input: `
				[00:10.00] First line

				[00:15.00] Second line

				[00:20.00] Third line
			`,
			shouldError:   false,
			expectedLines: 3,
			validate: func(t *testing.T, lines []model.LyricLine) {
				assert.Equal(t, "First line", lines[0].Text)
				assert.Equal(t, "Second line", lines[1].Text)
				assert.Equal(t, "Third line", lines[2].Text)
			},
		},
		{
			name: "3-digit milliseconds format",
			input: `
				[00:10.000] First line
				[00:15.500] Second line
				[00:20.123] Third line
			`,
			shouldError:   false,
			expectedLines: 3,
			validate: func(t *testing.T, lines []model.LyricLine) {
				assert.Equal(t, 10.0, *lines[0].Seconds)
				assert.Equal(t, 15.5, *lines[1].Seconds)
				assert.Equal(t, 20.123, *lines[2].Seconds)
			},
		},
		{
			name: "mixed 2-digit and 3-digit milliseconds",
			input: `
				[00:10.00] Two digits
				[00:15.500] Three digits
				[00:20.99] Two digits again
			`,
			shouldError:   false,
			expectedLines: 3,
			validate: func(t *testing.T, lines []model.LyricLine) {
				assert.Equal(t, 10.0, *lines[0].Seconds)
				assert.Equal(t, 15.5, *lines[1].Seconds)
				assert.Equal(t, 20.99, *lines[2].Seconds)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			lines, err := parser.ParseSyncedLyrics(tt.input)

			if tt.shouldError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, lines, tt.expectedLines)

			if tt.validate != nil {
				tt.validate(t, lines)
			}
		})
	}
}
