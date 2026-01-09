package service

import "regexp"

// Parser handles lyrics parsing and implements the LyricsParser interface
type Parser struct {
	timestampRegex *regexp.Regexp
}

// NewParser creates a new parser instance
func NewParser() *Parser {
	return &Parser{
		// Matches: [mm:ss.xx] text
		timestampRegex: regexp.MustCompile(`\[(\d+):(\d+\.\d+)\]\s*(.+)`),
	}
}
