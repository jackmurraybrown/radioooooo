package episode

import "time"

const (
	TypeLive     = "live"
	TypeRecorded = "recorded"
	TypeExternal = "external"
	TypePlaylist = "playlist"
)

type Source struct {
	Adapter string `json:"adapter"`
	Ref     string `json:"ref"`
}

type Episode struct {
	ID          string    `json:"id"`
	ChannelID   string    `json:"channelId"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	StartTime   time.Time `json:"startTime"`
	EndTime     time.Time `json:"endTime"`
	Type        string    `json:"type"`
	Source      Source    `json:"source"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// dbEpisode is a flat struct that maps 1:1 to the episodes table columns.
// it is used for scanning db rows and converted to Episode before returning to callers.
type dbEpisode struct {
	ID            string    `db:"id"`
	ChannelID     string    `db:"channel_id"`
	Title         string    `db:"title"`
	Description   string    `db:"description"`
	StartTime     time.Time `db:"start_time"`
	EndTime       time.Time `db:"end_time"`
	Type          string    `db:"type"`
	SourceAdapter string    `db:"source_adapter"`
	SourceRef     string    `db:"source_ref"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

func (d dbEpisode) toEpisode() Episode {
	return Episode{
		ID:          d.ID,
		ChannelID:   d.ChannelID,
		Title:       d.Title,
		Description: d.Description,
		StartTime:   d.StartTime,
		EndTime:     d.EndTime,
		Type:        d.Type,
		Source:      Source{Adapter: d.SourceAdapter, Ref: d.SourceRef},
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}
