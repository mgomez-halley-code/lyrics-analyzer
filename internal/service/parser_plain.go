package service

import (
	"fmt"
	"strings"

	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/model"
)

// ParsePlainLyrics parses plain lyrics without timestamps
func (p *Parser) ParsePlainLyrics(plainLyrics string) ([]model.LyricLine, error) {
	if plainLyrics == "" {
		return nil, fmt.Errorf("plain lyrics are empty")
	}

	lines := strings.Split(plainLyrics, "\n")
	var lyricLines []model.LyricLine
	lineNumber := 1

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		wordCount := len(strings.Fields(line))

		lyricLines = append(lyricLines, model.LyricLine{
			LineNumber: lineNumber,
			Text:       line,
			WordCount:  wordCount,
		})

		lineNumber++
	}

	return lyricLines, nil
}
