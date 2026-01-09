package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser_ParsePlainLyrics(t *testing.T) {
	parser := NewParser()

	plainLyrics := `First line of lyrics
Second line here
Third line too`

	lines, err := parser.ParsePlainLyrics(plainLyrics)
	assert.NoError(t, err)
	assert.Len(t, lines, 3)

	// Check first line
	assert.Equal(t, 1, lines[0].LineNumber)
	assert.Equal(t, "First line of lyrics", lines[0].Text)

	// Plain lyrics should have no timestamps
	assert.Nil(t, lines[0].Timestamp)
	assert.Nil(t, lines[0].Seconds)

	// Check word count
	assert.Equal(t, 4, lines[0].WordCount)
}

func TestParser_EmptyPlainLyrics(t *testing.T) {
	parser := NewParser()

	_, err := parser.ParsePlainLyrics("")
	assert.Error(t, err)
}
