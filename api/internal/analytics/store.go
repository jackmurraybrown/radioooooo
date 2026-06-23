package analytics

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

// ⋆˙⟡ upserts a listener count for the current hour bucket.
// increments samples, updates listeners (running latest), and tracks peak.
func (s *Store) Record(ctx context.Context, channelID string, episodeID *string, countryCode string, count int) error {
	hour := time.Now().Truncate(time.Hour)
	_, err := s.db.Exec(ctx, `
		insert into listener_stats (channel_id, episode_id, hour, country_code, listeners, peak, samples)
		values ($1::uuid, $2, $3, $4, $5, $5, 1)
		on conflict (channel_id, hour, country_code)
		do update set
			episode_id = coalesce($2, listener_stats.episode_id),
			listeners = $5,
			peak = greatest(listener_stats.peak, $5),
			samples = listener_stats.samples + 1
	`, channelID, episodeID, hour, countryCode, count)
	return err
}

// ✮⋆‧° current listeners across all countries for a channel
func (s *Store) CurrentListeners(ctx context.Context, channelID string) (int, error) {
	hour := time.Now().Truncate(time.Hour)
	var total int
	err := s.db.QueryRow(ctx, `
		select coalesce(sum(listeners), 0)
		from listener_stats
		where channel_id = $1::uuid and hour = $2
	`, channelID, hour).Scan(&total)
	return total, err
}

// ⋆˙⟡ peak listeners for a time range
func (s *Store) PeakListeners(ctx context.Context, channelID string, from, to time.Time) (int, error) {
	var peak int
	err := s.db.QueryRow(ctx, `
		select coalesce(max(peak), 0)
		from listener_stats
		where channel_id = $1::uuid and hour >= $2 and hour < $3
	`, channelID, from, to).Scan(&peak)
	return peak, err
}

// ⊹ ࣪ ˖ country breakdown for a time range
func (s *Store) CountryBreakdown(ctx context.Context, channelID string, from, to time.Time) ([]CountryCount, error) {
	rows, err := s.db.Query(ctx, `
		select country_code, sum(listeners) as listeners
		from listener_stats
		where channel_id = $1::uuid and hour >= $2 and hour < $3
		group by country_code
		order by listeners desc
	`, channelID, from, to)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[CountryCount])
}

// . ݁₊ ✶ listeners per hour for a date range — for line charts
func (s *Store) TimeSeries(ctx context.Context, channelID string, from, to time.Time) ([]HourlyCount, error) {
	rows, err := s.db.Query(ctx, `
		select hour, sum(listeners) as listeners, max(peak) as peak
		from listener_stats
		where channel_id = $1::uuid and hour >= $2 and hour < $3
		group by hour
		order by hour
	`, channelID, from, to)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[HourlyCount])
}
