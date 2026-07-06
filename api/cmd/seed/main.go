// ⋆˙⟡ ⋆.˚ ⊹₊⟡ seed — dev database seeder using the store layer
// structured so helpers can be lifted straight into store tests later.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"radioooooo/internal/channel"
	"radioooooo/internal/database"
	"radioooooo/internal/episode"
	"radioooooo/internal/media"
	"radioooooo/internal/station"
	"radioooooo/internal/tracklist"
	"radioooooo/internal/user"
)

// rng is seeded for reproducible index picks ٩(^ᗜ^ )و
var rng = rand.New(rand.NewSource(42))

var genres = []string{
	"house", "techno", "jungle", "drum & bass", "ambient", "jazz",
	"soul", "gabber", "downtempo", "afrobeat", "reggaeton", "experimental", "club", "grime",
	"footwork", "dub", "electro",
}

var showTemplates = []string{
	"%s's %s session",
	"%s frequencies",
	"late night w/ %s",
	"%s radio",
	"the %s hour",
	"%s & friends",
	"%s presents",
	"club %s w/ %s",
	"%s selections",
	"deep %s w/ %s",
	"%s transmissions",
	"the %s show",
}

func showName() string {
	tmpl := showTemplates[rng.Intn(len(showTemplates))]
	genre := genres[rng.Intn(len(genres))]
	name := strings.ToLower(gofakeit.FirstName() + " " + gofakeit.LastName())
	if strings.Count(tmpl, "%s") == 2 {
		return fmt.Sprintf(tmpl, genre, name)
	}
	if rng.Intn(2) == 0 {
		return fmt.Sprintf(tmpl, name)
	}
	return fmt.Sprintf(tmpl, genre)
}

func trackTitle() string {
	genre := genres[rng.Intn(len(genres))]
	artist := gofakeit.FirstName() + " " + gofakeit.LastName()
	tmpls := []string{
		"%s - %s ep",
		"%s / %s",
		"%s (%s edit)",
		"%s [%s remix]",
		"night %s (%s version)",
		"%s on %s",
	}
	return fmt.Sprintf(tmpls[rng.Intn(len(tmpls))], artist, genre)
}

// --- store helpers — portable into test files ✮ ⋆ ˚｡𖦹 ---

func seedStation(ctx context.Context, store *station.Store) (station.Station, error) {
	all, err := store.List(ctx)
	if err != nil {
		return station.Station{}, err
	}
	for _, s := range all {
		if s.Slug == "radiooo" {
			slog.Info("station already exists, skipping")
			return s, nil
		}
	}
	return store.Create(ctx, "radiooo", "radiooo")
}

func seedUser(ctx context.Context, db *pgxpool.Pool, store *user.Store, stationID string) error {
	u, hash, err := store.GetByEmail(ctx, "admin@radiooo.fm")
	if err == nil {
		// user exists — reset password to "password" if it differs . ݁₊ ✶. ݁
		if !user.CheckPassword(hash, "password") {
			if err := resetPassword(ctx, db, u.ID, "password"); err != nil {
				return err
			}
			slog.Info("user password reset", "id", u.ID)
		} else {
			slog.Info("user already exists, skipping", "id", u.ID)
		}
		return nil
	}
	_, err = store.Create(ctx, stationID, "admin@radiooo.fm", "password")
	return err
}

func resetPassword(ctx context.Context, db *pgxpool.Pool, userID, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx, `update users set password_hash=$2 where id=$1::uuid`, userID, string(hash))
	return err
}

func seedChannel(ctx context.Context, store *channel.Store, stationID string) (channel.Channel, error) {
	all, err := store.List(ctx, stationID)
	if err != nil {
		return channel.Channel{}, err
	}
	for _, ch := range all {
		if ch.Slug == "main" {
			slog.Info("channel already exists, skipping")
			return ch, nil
		}
	}
	return store.Create(ctx, stationID, "main", "main")
}

func seedMedia(ctx context.Context, ms *media.Store, stationID string) ([]string, error) {
	// check if already seeded — skip to avoid duplicates on re-run . ݁₊ ✶. ݁
	existing, err := ms.List(ctx, stationID)
	if err != nil {
		return nil, err
	}
	if len(existing) > 0 {
		ids := make([]string, len(existing))
		for i, m := range existing {
			ids[i] = m.ID
		}
		slog.Info("media already seeded, skipping", "count", len(ids))
		return ids, nil
	}

	ids := make([]string, 0, 20)
	for range 20 {
		title := trackTitle()
		artist := gofakeit.FirstName() + " " + gofakeit.LastName()
		format := media.FormatMP3
		durationSec := 180 + rng.Intn(300) // 3–8 min
		sizeBytes := int64(durationSec) * 32 * 1024

		m, err := ms.Create(ctx, media.CreateParams{
			StationID:     stationID,
			Title:         title,
			Artist:        &artist,
			FileFormat:    &format,
			FileSizeBytes: &sizeBytes,
			SourceAdapter: "local",
			SourceRef:     fmt.Sprintf("media/%s.mp3", gofakeit.UUID()),
		})
		if err != nil {
			return nil, err
		}

		// flip to ready with a realistic duration
		if err := ms.UpdateStatus(ctx, m.ID, stationID, media.DownloadStatusReady, &durationSec); err != nil {
			return nil, err
		}
		ids = append(ids, m.ID)
	}
	return ids, nil
}

// episode source options ⊹ ࣪ ˖
var epSources = []struct {
	epType  string
	adapter string
	ref     string // empty = use a media id
}{
	{"recorded", "media", ""},
	{"recorded", "media", ""},
	{"recorded", "media", ""},
	{"live", "icecast", "main"},
	{"external", "external", "https://stream.example.com/live"},
}

func seedSchedule(
	ctx context.Context,
	store *episode.Store,
	stationID, channelID string,
	mediaIDs []string,
) ([]episode.Episode, error) {
	// one week back, then 13 weeks forward ⋆˙⟡
	now := time.Now().UTC()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	thisMonday := time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, time.UTC)
	weekStart := thisMonday.Add(-7 * 24 * time.Hour) // start one week back

	var created []episode.Episode
	for week := range 14 { // 1 past + 13 forward
		for day := range 7 {
			dayStart := weekStart.Add(time.Duration(week*7+day) * 24 * time.Hour)
			cur := dayStart.Add(10 * time.Hour) // 10:00 ⊹ ࣪ ˖
			dayEnd := dayStart.Add(22 * time.Hour) // 22:00

			for cur.Before(dayEnd) {
				// 1-3h show, but cap so we don't run past 22:00
				durationH := time.Duration(1+rng.Intn(3)) * time.Hour
				end := cur.Add(durationH)
				if end.After(dayEnd) {
					end = dayEnd
				}

				src := epSources[rng.Intn(len(epSources))]
				ref := src.ref
				if src.epType == "recorded" && len(mediaIDs) > 0 {
					ref = mediaIDs[rng.Intn(len(mediaIDs))]
				}

				ep, err := store.Create(ctx, episode.CreateParams{
					ChannelID:     channelID,
					StationID:     stationID,
					Title:         showName(),
					StartTime:     cur,
					EndTime:       end,
					Type:          src.epType,
					SourceAdapter: src.adapter,
					SourceRef:     ref,
				})
				if err != nil {
					return created, err
				}
				created = append(created, ep)
				cur = end // ✶. ݁ no gap
			}
		}
	}
	return created, nil
}

// ✮ ⋆ ˚｡𖦹 seed tracklists for episodes that have already ended
func seedTracklists(ctx context.Context, store *tracklist.Store, episodes []episode.Episode) (int, error) {
	now := time.Now().UTC()
	count := 0
	for _, ep := range episodes {
		if !ep.EndTime.Before(now) {
			continue // only past episodes
		}
		durationSecs := int(ep.EndTime.Sub(ep.StartTime).Seconds())
		trackCount := 4 + rng.Intn(9) // 4–12 tracks
		inputs := make([]tracklist.TrackInput, 0, trackCount)
		cursor := rng.Intn(300) // first track 0–5 min in
		for range trackCount {
			artist := gofakeit.FirstName() + " " + gofakeit.LastName()
			at := cursor
			inp := tracklist.TrackInput{
				Title:     trackTitle(),
				Artist:    &artist,
				StartedAt: &at,
			}
			inputs = append(inputs, inp)
			cursor += 180 + rng.Intn(420) // 3–10 min per track
			if cursor >= durationSecs {
				break
			}
		}
		if _, err := store.SetTracks(ctx, ep.ID, inputs); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func run(ctx context.Context) error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL not set")
	}

	if err := database.Migrate(dsn); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	db, err := database.Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer db.Close()

	gofakeit.Seed(42) // ✮⋆‧°—°‧⋆✮ reproducible names

	stations := station.NewStore(db)
	users := user.NewStore(db)
	channels := channel.NewStore(db, "")
	media := media.NewStore(db)
	episodes := episode.NewStore(db)
	tracklists := tracklist.NewStore(db)

	st, err := seedStation(ctx, stations)
	if err != nil {
		return fmt.Errorf("station: %w", err)
	}
	slog.Info("station", "id", st.ID, "slug", st.Slug)

	if err := seedUser(ctx, db, users, st.ID); err != nil {
		return fmt.Errorf("user: %w", err)
	}
	slog.Info("user created", "email", "admin@radiooo.fm", "password", "password")

	ch, err := seedChannel(ctx, channels, st.ID)
	if err != nil {
		return fmt.Errorf("channel: %w", err)
	}
	slog.Info("channel", "id", ch.ID, "slug", ch.Slug)

	mediaIDs, err := seedMedia(ctx, media, st.ID)
	if err != nil {
		return fmt.Errorf("media: %w", err)
	}
	slog.Info("media", "count", len(mediaIDs))

	// wipe existing episodes for this channel then re-seed ⊹ ₊
	if _, err := db.Exec(ctx, `delete from episodes where channel_id = $1::uuid`, ch.ID); err != nil {
		return fmt.Errorf("clearing episodes: %w", err)
	}

	created, err := seedSchedule(ctx, episodes, st.ID, ch.ID, mediaIDs)
	if err != nil {
		return fmt.Errorf("schedule: %w", err)
	}
	slog.Info("episodes", "count", len(created))

	tl, err := seedTracklists(ctx, tracklists, created)
	if err != nil {
		return fmt.Errorf("tracklists: %w", err)
	}
	slog.Info("tracklists", "count", tl)

	return nil
}

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		slog.Error("seed failed", "error", err)
		os.Exit(1)
	}
	slog.Info("seed complete")
}
