package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser_ParseSyncedLyrics(t *testing.T) {
	parser := NewParser()

	syncedLyrics := `[00:10.00] First line
[00:15.50] Second line
[00:20.00] Third line`

	lines, err := parser.ParseSyncedLyrics(syncedLyrics)
	assert.NoError(t, err)
	assert.Len(t, lines, 3)

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
}

func TestParser_EmptySyncedLyrics(t *testing.T) {
	parser := NewParser()

	_, err := parser.ParseSyncedLyrics("")
	assert.Error(t, err)
}

func TestParser_MalformedSyncedLyrics(t *testing.T) {
	parser := NewParser()

	// Lyrics with some valid and some invalid lines
	syncedLyrics := `[00:10.00] Valid line
This line has no timestamp
[00:20.00] Another valid line
[invalid] Bad timestamp`

	lines, err := parser.ParseSyncedLyrics(syncedLyrics)
	assert.NoError(t, err)
	assert.Len(t, lines, 2)
	assert.Equal(t, "Valid line", lines[0].Text)
	assert.Equal(t, "Another valid line", lines[1].Text)
}
