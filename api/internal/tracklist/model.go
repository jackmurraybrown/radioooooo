package tracklist

import "time"

type Track struct {
	ID        string    `json:"id"                   db:"id"`
	EpisodeID string    `json:"episodeId"            db:"episode_id"`
	Position  int       `json:"position"             db:"position"`
	Title     string    `json:"title"                db:"title"`
	Artist    *string   `json:"artist,omitempty"     db:"artist"`
	Album     *string   `json:"album,omitempty"      db:"album"`
	StartedAt *int      `json:"startedAt,omitempty"  db:"started_at"`
	EndedAt   *int      `json:"endedAt,omitempty"    db:"ended_at"`
	CreatedAt time.Time `json:"createdAt"            db:"created_at"`
}

// ⊹ ˖ input from both the DJ form and the wails webhook
type TrackInput struct {
	Title     string  `json:"title"                minLength:"1"`
	Artist    *string `json:"artist,omitempty"`
	Album     *string `json:"album,omitempty"`
	StartedAt *int    `json:"startedAt,omitempty"  doc:"seconds offset from episode start"`
	EndedAt   *int    `json:"endedAt,omitempty"    doc:"seconds offset from episode start"`
}

// ⋆˙⟡ episode info returned alongside tracks
type EpisodeInfo struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	StationID string    `json:"-"`
}
