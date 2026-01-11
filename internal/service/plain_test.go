package service

import (
	"testing"

	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestParser_ParsePlainLyrics(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		shouldError   bool
		expectedLines int
		validate      func(t *testing.T, lines []model.LyricLine)
	}{
		{
			name: "basic plain lyrics",
			input: `
				First line of lyrics
				Second line here
				Third line too
			`,
			shouldError:   false,
			expectedLines: 3,
			validate: func(t *testing.T, lines []model.LyricLine) {
				// Check first line
				assert.Equal(t, 1, lines[0].LineNumber)
				assert.Equal(t, "First line of lyrics", lines[0].Text)
				assert.Equal(t, 4, lines[0].WordCount)

				// Plain lyrics should have no timestamps
				assert.Nil(t, lines[0].Timestamp)

				// Check second line
				assert.Equal(t, 2, lines[1].LineNumber)
				assert.Equal(t, "Second line here", lines[1].Text)
				assert.Equal(t, 3, lines[1].WordCount)

				// Check third line
				assert.Equal(t, 3, lines[2].LineNumber)
				assert.Equal(t, "Third line too", lines[2].Text)
				assert.Equal(t, 3, lines[2].WordCount)
			},
		},
		{
			name:          "empty plain lyrics",
			input:         "",
			shouldError:   true,
			expectedLines: 0,
		},
		{
			name: "blank lines ignored",
			input: `
				First line

				Second line

				Third line
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
			name: "extra whitespace trimmed",
			input: `
				  Trimmed text
				Tab trimmed
			`,
			shouldError:   false,
			expectedLines: 2,
			validate: func(t *testing.T, lines []model.LyricLine) {
				assert.Equal(t, "Trimmed text", lines[0].Text)
				assert.Equal(t, "Tab trimmed", lines[1].Text)
			},
		},
		{
			name: "single word vs multi-word",
			input: `
				Word
				Multiple words here
			`,
			shouldError:   false,
			expectedLines: 2,
			validate: func(t *testing.T, lines []model.LyricLine) {
				assert.Equal(t, 1, lines[0].WordCount)
				assert.Equal(t, 3, lines[1].WordCount)
			},
		},
		{
			name: "special characters preserved",
			input: `
				Hello! How's it going?
				C'est la vie & rock 'n' roll
				100% awesome!!!
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
			name: "line numbers sequential",
			input: `
				Line one
				Line two
				Line three
				Line four
			`,
			shouldError:   false,
			expectedLines: 4,
			validate: func(t *testing.T, lines []model.LyricLine) {
				for i, line := range lines {
					assert.Equal(t, i+1, line.LineNumber)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			lines, err := parser.ParsePlainLyrics(tt.input)

			if tt.shouldError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, lines, tt.expectedLines)

			// All plain lyrics should have nil timestamps
			for _, line := range lines {
				assert.Nil(t, line.Timestamp)
			}

			if tt.validate != nil {
				tt.validate(t, lines)
			}
		})
	}
}
