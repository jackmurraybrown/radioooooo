package ical

import "time"

type Feed struct {
	ID           string     `json:"id"                       db:"id"`
	StationID    string     `json:"stationId"                db:"station_id"`
	ChannelID    string     `json:"channelId"                db:"channel_id"`
	Type         string     `json:"type"                     db:"type"`
	URL          string     `json:"url"                      db:"url"`
	Username     *string    `json:"username,omitempty"        db:"username"`
	Password     *string    `json:"-"                         db:"password"`
	CalendarPath *string    `json:"calendarPath,omitempty"    db:"calendar_path"`
	LastSyncedAt *time.Time `json:"lastSyncedAt"             db:"last_synced_at"`
	CreatedAt    time.Time  `json:"createdAt"                db:"created_at"`
	UpdatedAt    time.Time  `json:"updatedAt"                db:"updated_at"`
}
