package user

import (
	"errors"
	"time"
)

var errWrongPassword = errors.New("current password is incorrect")

// --- types ˚₊✧ ---

type User struct {
	ID        string    `json:"id"        db:"id"`
	StationID string    `json:"stationId" db:"station_id"`
	Email     string    `json:"email"     db:"email"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}
