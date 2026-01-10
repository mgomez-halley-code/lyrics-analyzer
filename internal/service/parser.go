package service

import "regexp"

// Parser handles lyrics parsing and implements the LyricsParser interface
type Parser struct {
	timestampRegex *regexp.Regexp
}

// NewParser creates a new parser instance
func NewParser() *Parser {
	return &Parser{
		// Matches: [mm:ss.xx] or [mm:ss.xxx] text
		// Supports both 2-digit (00:10.50) and 3-digit (00:10.500) milliseconds
		timestampRegex: regexp.MustCompile(`\[(\d+):(\d+\.\d{2,3})\]\s*(.+)`),
	}
}
