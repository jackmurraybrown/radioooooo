package playlist

import "time"

// --- types ˚₊✧ ---

type Playlist struct {
	ID            string    `json:"id"                      db:"id"`
	StationID     string    `json:"stationId"               db:"station_id"`
	Name          string    `json:"name"                    db:"name"`
	Shuffle       bool      `json:"shuffle"                 db:"shuffle"`
	Loop          bool      `json:"loop"                    db:"loop"`
	SourceAdapter *string   `json:"sourceAdapter,omitempty" db:"source_adapter"`
	SourceRef     *string   `json:"sourceRef,omitempty"     db:"source_ref"`
	CreatedAt     time.Time `json:"createdAt"               db:"created_at"`
	UpdatedAt     time.Time `json:"updatedAt"               db:"updated_at"`
}

type PlaylistTrack struct {
	LocalRef     string   `db:"local_ref"`
	LoudnessLUFS *float64 `db:"loudness_lufs"`
}

// PlaylistItem joins playlist_items with media for display in the admin UI.
type PlaylistItem struct {
	ID         string    `json:"id"                db:"id"`
	PlaylistID string    `json:"playlistId"        db:"playlist_id"`
	MediaID    string    `json:"mediaId"           db:"media_id"`
	Position   int       `json:"position"          db:"position"`
	Title      string    `json:"title"             db:"title"`
	Artist     *string   `json:"artist,omitempty"  db:"artist"`
	Duration   *int      `json:"duration,omitempty" db:"duration"`
	CreatedAt  time.Time `json:"createdAt"         db:"created_at"`
}
