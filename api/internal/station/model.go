package station

import "time"

type Station struct {
	ID                     string     `json:"id"                              db:"id"`
	Name                   string     `json:"name"                            db:"name"`
	Slug                   string     `json:"slug"                            db:"slug"`
	Timezone               string     `json:"timezone"                        db:"timezone"`
	SeasonEndDate          *time.Time `json:"seasonEndDate,omitempty"         db:"season_end_date"`
	DefaultShowHorizonDays int        `json:"defaultShowHorizonDays"          db:"default_show_horizon_days"`
	LogoURL                *string    `json:"logoUrl,omitempty"               db:"logo_url"`
	CreatedAt              time.Time  `json:"createdAt"                       db:"created_at"`
	UpdatedAt              time.Time  `json:"updatedAt"                       db:"updated_at"`
}
