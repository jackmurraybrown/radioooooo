package notify

// ⊹ ࣪ ˖ show form reminders — emails a form link 10 days before an episode
// runs once daily at 9am UTC

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
)

type ReminderArgs struct{}

func (ReminderArgs) Kind() string { return "show_reminder" }

type ReminderWorker struct {
	river.WorkerDefaults[ReminderArgs]
	db        *pgxpool.Pool
	mailer    Mailer
	templates *TemplateStore
}

func NewReminderWorker(db *pgxpool.Pool, mailer Mailer, templates *TemplateStore) *ReminderWorker {
	return &ReminderWorker{db: db, mailer: mailer, templates: templates}
}

// ⋆˙⟡ template variables available to station admins
type reminderVars struct {
	Title       string
	DaysUntil   int
	StartTime   string
	StationName string
}

type upcomingEpisode struct {
	ID           string    `db:"id"`
	ChannelID    string    `db:"channel_id"`
	StationID    string    `db:"station_id"`
	Title        string    `db:"title"`
	StartTime    time.Time `db:"start_time"`
	ContactEmail string    `db:"contact_email"`
	StationName  string    `db:"station_name"`
	LogoURL      *string   `db:"logo_url"`
}

func (w *ReminderWorker) Work(ctx context.Context, job *river.Job[ReminderArgs]) error {
	// find episodes starting in the next 10 days that haven't been reminded ⊹ ˖
	rows, err := w.db.Query(ctx, `
		select e.id::text, e.channel_id::text, st.id::text as station_id,
		       e.title, e.start_time, e.contact_email,
		       st.name as station_name, st.logo_url
		from episodes e
		join channels c on c.id = e.channel_id
		join stations st on st.id = c.station_id
		where e.start_time > now()
		  and e.start_time <= now() + interval '10 days'
		  and e.contact_email is not null
		  and e.id not in (select episode_id from sent_reminders)
	`)
	if err != nil {
		return fmt.Errorf("reminder: query: %w", err)
	}

	episodes, err := pgx.CollectRows(rows, pgx.RowToStructByName[upcomingEpisode])
	if err != nil {
		return fmt.Errorf("reminder: collect: %w", err)
	}

	for _, ep := range episodes {
		daysUntil := int(time.Until(ep.StartTime).Hours() / 24)

		vars := reminderVars{
			Title:       ep.Title,
			DaysUntil:   daysUntil,
			StartTime:   ep.StartTime.Format("Mon 2 Jan at 15:04 UTC"),
			StationName: ep.StationName,
		}

		tmpl, _ := w.templates.Get(ctx, ep.StationID, "show_reminder")
		subject, err := renderTemplate(tmpl.Subject, vars)
		if err != nil {
			slog.Error("reminder: render subject", "episode", ep.ID, "error", err)
			subject = fmt.Sprintf("your show \"%s\" is in %d days", ep.Title, daysUntil)
		}
		bodyMd, err := renderTemplate(tmpl.Body, vars)
		if err != nil {
			slog.Error("reminder: render body", "episode", ep.ID, "error", err)
			bodyMd = fmt.Sprintf("your show **%s** is scheduled in %d days (%s).",
				ep.Title, daysUntil, vars.StartTime)
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
			slog.Error("reminder: render failed", "episode", ep.ID, "error", err)
			continue
		}

		if err := w.mailer.Send(ctx, ep.ContactEmail, subject, html, plain); err != nil {
			slog.Error("reminder: send failed", "episode", ep.ID, "to", ep.ContactEmail, "error", err)
			continue
		}

		if _, err := w.db.Exec(ctx, `
			insert into sent_reminders (episode_id) values ($1::uuid) on conflict do nothing
		`, ep.ID); err != nil {
			slog.Warn("reminder: mark sent failed", "episode", ep.ID, "error", err)
		}

		slog.Info("reminder: sent", "episode", ep.ID, "to", ep.ContactEmail, "days_until", daysUntil)
	}

	return nil
}

type DailyAt9AM struct{}

func (DailyAt9AM) Next(after time.Time) time.Time {
	next := time.Date(after.Year(), after.Month(), after.Day(), 9, 0, 0, 0, time.UTC)
	if !next.After(after) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}

func renderTemplate(tmplStr string, data reminderVars) (string, error) {
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
