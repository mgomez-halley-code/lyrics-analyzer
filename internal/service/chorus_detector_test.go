package service

import (
	"testing"

	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestChorusDetector_DetectChorus_ThreeOrMoreOccurrences(t *testing.T) {
	detector := NewChorusDetector()

	lines := []model.LyricLine{
		{LineNumber: 1, Text: "Verse line one"},
		{LineNumber: 2, Text: "Never gonna give you up"},
		{LineNumber: 3, Text: "Verse line two"},
		{LineNumber: 4, Text: "Never gonna give you up"},
		{LineNumber: 5, Text: "Verse line three"},
		{LineNumber: 6, Text: "Never gonna give you up"},
	}

	chorus := detector.DetectChorus(lines)

	assert.True(t, chorus.Detected)
	assert.Equal(t, "Never gonna give you up", chorus.Text)
	assert.Equal(t, 3, chorus.Occurrences)
	assert.Equal(t, []int{2, 4, 6}, chorus.LineNumbers)
}

func TestChorusDetector_DetectChorus_TwoOccurrences(t *testing.T) {
	detector := NewChorusDetector()

	lines := []model.LyricLine{
		{LineNumber: 1, Text: "Verse line one"},
		{LineNumber: 2, Text: "Repeated line here"},
		{LineNumber: 3, Text: "Verse line two"},
		{LineNumber: 4, Text: "Repeated line here"},
		{LineNumber: 5, Text: "Verse line three"},
	}

	chorus := detector.DetectChorus(lines)

	assert.True(t, chorus.Detected, "Expected chorus to be detected (2 occurrences)")
	assert.Equal(t, 2, chorus.Occurrences)
}


func TestChorusDetector_DetectChorus_NoRepetition(t *testing.T) {
	detector := NewChorusDetector()

	lines := []model.LyricLine{
		{LineNumber: 1, Text: "Line one"},
		{LineNumber: 2, Text: "Line two"},
		{LineNumber: 3, Text: "Line three"},
		{LineNumber: 4, Text: "Line four"},
	}

	chorus := detector.DetectChorus(lines)

	assert.False(t, chorus.Detected, "Expected no chorus detection (no repetition)")
	assert.Equal(t, 0, chorus.Occurrences)
}

func TestChorusDetector_DetectChorus_EmptyLines(t *testing.T) {
	detector := NewChorusDetector()

	lines := []model.LyricLine{}

	chorus := detector.DetectChorus(lines)

	assert.False(t, chorus.Detected, "Expected no chorus detection (empty lines)")
}

func TestChorusDetector_DetectChorus_PreferMoreOccurrences(t *testing.T) {
	detector := NewChorusDetector()

	lines := []model.LyricLine{
		{LineNumber: 1, Text: "Line A"},
		{LineNumber: 2, Text: "Line B"},
		{LineNumber: 3, Text: "Line A"},
		{LineNumber: 4, Text: "Line B"},
		{LineNumber: 5, Text: "Line B"},
		{LineNumber: 6, Text: "Line B"},
	}

	chorus := detector.DetectChorus(lines)

	assert.True(t, chorus.Detected)
	// Should prefer "Line B" (4 occurrences) over "Line A" (2 occurrences)
	assert.Equal(t, "Line B", chorus.Text, "Expected 'Line B' (most repeated)")
	assert.Equal(t, 4, chorus.Occurrences)
}

