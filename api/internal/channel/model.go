package channel

import "time"

type Channel struct {
	ID        string    `json:"id"        db:"id"`
	StationID string    `json:"stationId" db:"station_id"`
	Name      string    `json:"name"      db:"name"`
	Slug      string    `json:"slug"      db:"slug"`
	Mount     string    `json:"mount"     db:"mount"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}
