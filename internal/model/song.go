package model

// Track represents song metadata
type Track struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Artist       string `json:"artist"`
	Album        string `json:"album,omitempty"`
	Duration     int    `json:"duration"`
	Instrumental bool   `json:"instrumental"`
}
