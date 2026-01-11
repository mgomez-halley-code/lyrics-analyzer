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

// Statistics contains lyrics statistics
type Statistics struct {
	TotalLines          int     `json:"totalLines"`
	UniqueLines         int     `json:"uniqueLines"`
	TotalWords          int     `json:"totalWords"`
	UniqueWords         int     `json:"uniqueWords"`
	AverageWordsPerLine float64 `json:"averageWordsPerLine"`
	RepetitionRatio     float64 `json:"repetitionRatio"`
}
