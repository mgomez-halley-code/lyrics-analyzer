package model

// LyricsSourceData represents raw lyrics data returned by external providers.
// This is the contract between the client layer and service layer.
// It contains unprocessed data that will be parsed and enriched by the service.
//
// This model lives in the model package (not service) to avoid circular dependencies:
// client → service would create a cycle if service owned this type.
type LyricsSourceData struct {
	TrackID      int
	TrackName    string
	ArtistName   string
	AlbumName    string
	Duration     int
	Instrumental bool
	SyncedLyrics string
	PlainLyrics  string
}
