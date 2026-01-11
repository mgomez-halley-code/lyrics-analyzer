package model

import "errors"

// Lyrics type constants
const (
	LyricsTypeSynced = "synced"
	LyricsTypePlain  = "plain"
)

// Source constants
const (
	SourceLRCLib = "lrclib"
)

// Sentinel errors for type-safe error handling
var (
	ErrNotFound    = errors.New("lyrics not found")
	ErrRateLimited = errors.New("rate limit exceeded")
)
