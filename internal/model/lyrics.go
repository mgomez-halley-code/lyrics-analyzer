package model

// LyricLine represents a single line of lyrics
type LyricLine struct {
	LineNumber int     `json:"lineNumber"`
	Timestamp  *string `json:"timestamp,omitempty"`
	Text       string  `json:"text"`
	WordCount  int     `json:"wordCount"`
}

// LyricsData contains structured lyrics information
type LyricsData struct {
	Type          string      `json:"type"` // "synced" or "plain"
	HasTimestamps bool        `json:"hasTimestamps"`
	TotalLines    int         `json:"totalLines"`
	Lines         []LyricLine `json:"lines"`
}

// Chorus represents detected chorus information
type Chorus struct {
	Detected    bool   `json:"detected"`
	Text        string `json:"text,omitempty"`
	Occurrences int    `json:"occurrences"`
	LineNumbers []int  `json:"lineNumbers,omitempty"`
}

// Structure contains song structure analysis
type Structure struct {
	Chorus *Chorus `json:"chorus"`
}
