package model

import "time"

// SongAnalysisResponse is the main API response
type SongAnalysisResponse struct {
	Track      Track       `json:"track"`
	Lyrics     *LyricsData `json:"lyrics,omitempty"`
	Structure  *Structure  `json:"structure,omitempty"`
	Statistics *Statistics `json:"statistics,omitempty"`
	Metadata   Metadata    `json:"metadata"`
}

// Metadata contains response metadata
type Metadata struct {
	Source           string    `json:"source"`
	Cached           bool      `json:"cached"`
	ProcessingTimeMs int64     `json:"processingTimeMs"`
	Timestamp        time.Time `json:"timestamp"`
	Warnings         []string  `json:"warnings,omitempty"`
	Message          string    `json:"message,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}
