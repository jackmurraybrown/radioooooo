package notify

// ⊹ ࣪ ˖ ၊၊||၊ email template store — custom per-station overrides with defaults

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EmailTemplate struct {
	ID        string    `json:"id"        db:"id"`
	StationID string    `json:"stationId" db:"station_id"`
	Type      string    `json:"type"      db:"type"`
	Subject   string    `json:"subject"   db:"subject"`
	Body      string    `json:"body"      db:"body"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

var defaults = map[string]EmailTemplate{
	"show_reminder": {
		Type:    "show_reminder",
		Subject: `your show "{{.Title}}" is in {{.DaysUntil}} days`,
		Body:    `your show **{{.Title}}** is scheduled in {{.DaysUntil}} days ({{.StartTime}}).`,
	},
}

type TemplateStore struct {
	db *pgxpool.Pool
}

func NewTemplateStore(db *pgxpool.Pool) *TemplateStore {
	return &TemplateStore{db: db}
}

// Get returns the custom template for a station, or the hardcoded default ⋆˙⟡
func (s *TemplateStore) Get(ctx context.Context, stationID, templateType string) (EmailTemplate, error) {
	rows, err := s.db.Query(ctx, `
		select id::text, station_id::text, type, subject, body, created_at, updated_at
		from email_templates
		where station_id = $1::uuid and type = $2
	`, stationID, templateType)
	if err != nil {
		return EmailTemplate{}, err
	}
	tmpl, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[EmailTemplate])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			if def, ok := defaults[templateType]; ok {
				def.StationID = stationID
				return def, nil
			}
			return EmailTemplate{}, pgx.ErrNoRows
		}
		return EmailTemplate{}, err
	}
	return tmpl, nil
}

// Upsert creates or updates a station's custom template ✮⋆‧°
func (s *TemplateStore) Upsert(ctx context.Context, stationID, templateType, subject, body string) (EmailTemplate, error) {
	rows, err := s.db.Query(ctx, `
		insert into email_templates (station_id, type, subject, body)
		values ($1::uuid, $2, $3, $4)
		on conflict (station_id, type)
		do update set subject = excluded.subject, body = excluded.body, updated_at = now()
		returning id::text, station_id::text, type, subject, body, created_at, updated_at
	`, stationID, templateType, subject, body)
	if err != nil {
		return EmailTemplate{}, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[EmailTemplate])
}

// DefaultTemplate returns the hardcoded default for a given type
func DefaultTemplate(templateType string) (EmailTemplate, bool) {
	def, ok := defaults[templateType]
	return def, ok
}
