package show

import "time"

type Show struct {
	ID              string    `json:"id"              db:"id"`
	ChannelID       string    `json:"channelId"       db:"channel_id"`
	Title           string    `json:"title"           db:"title"`
	Description     string    `json:"description"     db:"description"`
	ImageRef        *string   `json:"imageRef"        db:"image_ref"`
	RecurrenceRule  string    `json:"recurrenceRule"  db:"recurrence_rule"`
	DurationMinutes int       `json:"durationMinutes" db:"duration_minutes"`
	Type            string    `json:"type"            db:"type"`
	SourceAdapter   string    `json:"sourceAdapter"   db:"source_adapter"`
	SourceRef       string    `json:"sourceRef"       db:"source_ref"`
	AllowRepeat     bool      `json:"allowRepeat"     db:"allow_repeat"`
	CreatedAt       time.Time `json:"createdAt"       db:"created_at"`
	UpdatedAt       time.Time `json:"updatedAt"       db:"updated_at"`
}
