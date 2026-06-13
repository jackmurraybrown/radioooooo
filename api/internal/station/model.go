package station

import "time"

type Station struct {
	ID        string    `json:"id"                db:"id"`
	Name      string    `json:"name"              db:"name"`
	Slug      string    `json:"slug"              db:"slug"`
	LogoURL   *string   `json:"logoUrl,omitempty" db:"logo_url"`
	CreatedAt time.Time `json:"createdAt"         db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt"         db:"updated_at"`
}
