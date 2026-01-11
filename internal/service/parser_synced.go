package service

import (
	"fmt"
	"strings"

	"github.com/mgomez-halley-code/lyrics-analyzer.git/internal/model"
)

// ParseSyncedLyrics parses synced lyrics with timestamps into structured format
func (p *Parser) ParseSyncedLyrics(syncedLyrics string) ([]model.LyricLine, error) {
	if syncedLyrics == "" {
		return nil, fmt.Errorf("synced lyrics are empty")
	}

	lines := strings.Split(syncedLyrics, "\n")
	var lyricLines []model.LyricLine
	lineNumber := 1

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Match timestamp pattern [mm:ss.xx] text
		matches := p.timestampRegex.FindStringSubmatch(line)
		if len(matches) != 4 {
			// Skip lines that don't match pattern
			continue
		}

		minutes := matches[1]
		seconds := matches[2]
		text := strings.TrimSpace(matches[3])

		// Parse timestamp
		timestamp := fmt.Sprintf("%s:%s", minutes, seconds)
		_, err := p.ParseTimestamp(timestamp)
		if err != nil {
			continue
		}

		// Count words
		wordCount := len(strings.Fields(text))

		lyricLines = append(lyricLines, model.LyricLine{
			LineNumber: lineNumber,
			Timestamp:  &timestamp,
			Text:       text,
			WordCount:  wordCount,
		})

		lineNumber++
	}

	return lyricLines, nil
}
