package notify

// ⊹ ࣪ ˖ show form reminders — emails a form link 10 days before an episode
// runs once daily at 9am UTC

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

type ReminderArgs struct{}

func (ReminderArgs) Kind() string { return "show_reminder" }

type ReminderWorker struct {
	river.WorkerDefaults[ReminderArgs]
	db     *pgxpool.Pool
	mailer Mailer
}

func NewReminderWorker(db *pgxpool.Pool, mailer Mailer) *ReminderWorker {
	return &ReminderWorker{db: db, mailer: mailer}
}

type upcomingEpisode struct {
	ID           string    `db:"id"`
	ChannelID    string    `db:"channel_id"`
	Title        string    `db:"title"`
	StartTime    time.Time `db:"start_time"`
	ContactEmail string    `db:"contact_email"`
}

func (w *ReminderWorker) Work(ctx context.Context, job *river.Job[ReminderArgs]) error {
	// find episodes starting in the next 10 days that also haven't been reminded
	rows, err := w.db.Query(ctx, `
		select e.id::text, e.channel_id::text, e.title, e.start_time, e.contact_email
		from episodes e
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

		subject := fmt.Sprintf("your show \"%s\" is in %d days", ep.Title, daysUntil)
		body := fmt.Sprintf(
			`<p>your show <strong>%s</strong> is scheduled in %d days (%s).</p>`,
			ep.Title, daysUntil, ep.StartTime.Format("Mon 2 Jan at 15:04 UTC"),
		)

		if err := w.mailer.Send(ctx, ep.ContactEmail, subject, body); err != nil {
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
