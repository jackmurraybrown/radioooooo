package notify

// . ݁₊ ✶. ˖ˎˊ˗ post-show tracklist email — sends submission link after episode ends

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"text/template"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/matcornic/hermes/v2"
	"github.com/riverqueue/river"
	"radioooooo/internal/tracklist"
)

type TracklistEmailArgs struct{}

func (TracklistEmailArgs) Kind() string { return "tracklist_email" }

type TracklistEmailWorker struct {
	river.WorkerDefaults[TracklistEmailArgs]
	db        *pgxpool.Pool
	mailer    Mailer
	templates *TemplateStore
	tracks    *tracklist.Store
	frontURL  string
}

func NewTracklistEmailWorker(db *pgxpool.Pool, mailer Mailer, templates *TemplateStore, tracks *tracklist.Store, frontURL string) *TracklistEmailWorker {
	return &TracklistEmailWorker{db: db, mailer: mailer, templates: templates, tracks: tracks, frontURL: frontURL}
}

type tracklistEmailVars struct {
	Title       string
	EndTime     string
	EditURL     string
	StationName string
}

type endedEpisode struct {
	ID           string    `db:"id"`
	Title        string    `db:"title"`
	EndTime      time.Time `db:"end_time"`
	ContactEmail string    `db:"contact_email"`
	StationID    string    `db:"station_id"`
	StationName  string    `db:"station_name"`
	LogoURL      *string   `db:"logo_url"`
}

func (w *TracklistEmailWorker) Work(ctx context.Context, job *river.Job[TracklistEmailArgs]) error {
	// ⋆˙⟡ find episodes that ended in the last 30 days without a tracklist email
	rows, err := w.db.Query(ctx, `
		select e.id::text, e.title, e.end_time, e.contact_email,
		       st.id::text as station_id, st.name as station_name, st.logo_url
		from episodes e
		join channels c on c.id = e.channel_id
		join stations st on st.id = c.station_id
		where e.end_time < now()
		  and e.end_time > now() - interval '30 days'
		  and e.contact_email is not null
		  and e.id not in (select episode_id from sent_tracklist_emails)
	`)
	if err != nil {
		return fmt.Errorf("tracklist email: query: %w", err)
	}

	episodes, err := pgx.CollectRows(rows, pgx.RowToStructByName[endedEpisode])
	if err != nil {
		return fmt.Errorf("tracklist email: collect: %w", err)
	}

	for _, ep := range episodes {
		token, err := w.tracks.CreateToken(ctx, ep.ID)
		if err != nil {
			slog.Error("tracklist email: token failed", "episode", ep.ID, "error", err)
			continue
		}

		editURL := fmt.Sprintf("%s/tracklist/%s", w.frontURL, token)
		vars := tracklistEmailVars{
			Title:       ep.Title,
			EndTime:     ep.EndTime.Format("Mon 2 Jan at 15:04 UTC"),
			EditURL:     editURL,
			StationName: ep.StationName,
		}

		tmpl, _ := w.templates.Get(ctx, ep.StationID, "tracklist_email")
		subject, err := renderTracklistTemplate(tmpl.Subject, vars)
		if err != nil {
			slog.Error("tracklist email: render subject", "episode", ep.ID, "error", err)
			subject = fmt.Sprintf("tracklist for \"%s\"", ep.Title)
		}
		bodyMd, err := renderTracklistTemplate(tmpl.Body, vars)
		if err != nil {
			slog.Error("tracklist email: render body", "episode", ep.ID, "error", err)
			bodyMd = fmt.Sprintf("your show **%s** has ended. [add your tracklist](%s)", ep.Title, editURL)
		}

		logoURL := ""
		if ep.LogoURL != nil {
			logoURL = *ep.LogoURL
		}
		html, plain, err := RenderStationEmail(ep.StationName, logoURL, hermes.Email{
			Body: hermes.Body{
				FreeMarkdown: hermes.Markdown(bodyMd),
			},
		})
		if err != nil {
			slog.Error("tracklist email: render failed", "episode", ep.ID, "error", err)
			continue
		}

		if err := w.mailer.Send(ctx, ep.ContactEmail, subject, html, plain); err != nil {
			slog.Error("tracklist email: send failed", "episode", ep.ID, "to", ep.ContactEmail, "error", err)
			continue
		}

		if _, err := w.db.Exec(ctx, `
			insert into sent_tracklist_emails (episode_id) values ($1::uuid) on conflict do nothing
		`, ep.ID); err != nil {
			slog.Warn("tracklist email: mark sent failed", "episode", ep.ID, "error", err)
		}

		slog.Info("tracklist email: sent", "episode", ep.ID, "to", ep.ContactEmail)
	}

	return nil
}

func renderTracklistTemplate(tmplStr string, data tracklistEmailVars) (string, error) {
	t, err := template.New("").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("parse: %w", err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("exec: %w", err)
	}
	return buf.String(), nil
}
