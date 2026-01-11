package service

import (
	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/model"
)

// ChorusDetector detects chorus sections in lyrics
type ChorusDetector struct{}

// NewChorusDetector creates a new chorus detector
func NewChorusDetector() *ChorusDetector {
	return &ChorusDetector{}
}

// DetectChorus identifies repeated sections (chorus)
func (cd *ChorusDetector) DetectChorus(lines []model.LyricLine) *model.Chorus {
	if len(lines) == 0 {
		return &model.Chorus{
			Detected: false,
		}
	}

	// Count occurrences of each line
	lineCount := make(map[string][]int)

	for _, line := range lines {
		if line.Text == "" {
			continue
		}

		lineCount[line.Text] = append(lineCount[line.Text], line.LineNumber)
	}

	// Find the most repeated line (must appear at least 2 times)
	var chorusLineNumbers []int
	maxOccurrences := 0

	for _, lineNumbers := range lineCount {
		if len(lineNumbers) >= 2 && len(lineNumbers) > maxOccurrences {
			maxOccurrences = len(lineNumbers)
			chorusLineNumbers = lineNumbers
		}
	}

	if maxOccurrences == 0 {
		return &model.Chorus{
			Detected: false,
		}
	}

	// Find the text for the chorus (first occurrence)
	var chorusText string
	for text, lineNumbers := range lineCount {
		if len(lineNumbers) == maxOccurrences {
			chorusText = text
			break
		}
	}

	return &model.Chorus{
		Detected:    true,
		Text:        chorusText,
		Occurrences: maxOccurrences,
		LineNumbers: chorusLineNumbers,
	}
}
