package analytics

import "time"

type ListenerStat struct {
	ID          string    `json:"id"          db:"id"`
	ChannelID   string    `json:"channelId"   db:"channel_id"`
	EpisodeID   *string   `json:"episodeId"   db:"episode_id"`
	Hour        time.Time `json:"hour"        db:"hour"`
	CountryCode string    `json:"countryCode" db:"country_code"`
	Listeners   int       `json:"listeners"   db:"listeners"`
	Peak        int       `json:"peak"        db:"peak"`
	Samples     int       `json:"samples"     db:"samples"`
	CreatedAt   time.Time `json:"createdAt"   db:"created_at"`
}

// ⊹ ࣪ ˖ summary returned by the analytics API
type ChannelStats struct {
	CurrentListeners int              `json:"currentListeners"`
	PeakListeners    int              `json:"peakListeners"`
	Countries        []CountryCount   `json:"countries"`
}

type CountryCount struct {
	CountryCode string `json:"countryCode" db:"country_code"`
	Listeners   int    `json:"listeners"   db:"listeners"`
}

type HourlyCount struct {
	Hour      time.Time `json:"hour"      db:"hour"`
	Listeners int       `json:"listeners" db:"listeners"`
	Peak      int       `json:"peak"      db:"peak"`
}
