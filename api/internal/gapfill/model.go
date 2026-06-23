package gapfill

import "time"

type Rule struct {
	ID        string    `json:"id"                    db:"id"`
	ChannelID string    `json:"channelId"             db:"channel_id"`
	Priority  int       `json:"priority"              db:"priority"`
	TimeFrom  *string   `json:"timeFrom,omitempty"    db:"time_from"`
	TimeTo    *string   `json:"timeTo,omitempty"      db:"time_to"`
	Type      string    `json:"type"                  db:"type"`
	SourceRef string    `json:"sourceRef"             db:"source_ref"`
	CreatedAt time.Time `json:"createdAt"             db:"created_at"`
}

// ⋆˙⟡ a gap in the schedule
type Gap struct {
	Start time.Time
	End   time.Time
}

func (g Gap) DurationMinutes() int {
	return int(g.End.Sub(g.Start).Minutes())
}
