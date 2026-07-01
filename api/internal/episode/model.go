package episode

import "time"

const (
	TypeLive     = "live"
	TypeRecorded = "recorded"
	TypeExternal = "external"
	TypePlaylist = "playlist"
)

type Episode struct {
	ID            string     `json:"id"                       db:"id"`
	ChannelID     string     `json:"channelId"                db:"channel_id"`
	ShowID        *string    `json:"showId,omitempty"         db:"show_id"`
	Title         string     `json:"title"                    db:"title"`
	Description   string     `json:"description,omitempty"    db:"description"`
	ImageRef      *string    `json:"imageRef,omitempty"       db:"image_ref"`
	Color         *string    `json:"color,omitempty"          db:"color"`
	StartTime     time.Time  `json:"startTime"                db:"start_time"`
	EndTime       time.Time  `json:"endTime"                  db:"end_time"`
	Type          string     `json:"type"                     db:"type"     enum:"live,recorded,external,playlist"`
	SourceAdapter string     `json:"sourceAdapter"            db:"source_adapter"`
	SourceRef     string     `json:"sourceRef"                db:"source_ref"`
	OriginalStart *time.Time `json:"originalStart,omitempty"  db:"original_start"`
	IcalUID       *string    `json:"icalUid,omitempty"        db:"ical_uid"`
	IcalFeedID    *string    `json:"icalFeedId,omitempty"     db:"ical_feed_id"`
	AutoFilled    bool       `json:"autoFilled"               db:"auto_filled"`
	RepeatOf      *string    `json:"repeatOf,omitempty"       db:"repeat_of"`
	ContactEmail  *string    `json:"contactEmail,omitempty"   db:"contact_email"`
	CreatedAt     time.Time  `json:"createdAt"                db:"created_at"`
	UpdatedAt     time.Time  `json:"updatedAt"                db:"updated_at"`
}
