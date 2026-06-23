package jobs

// ⋆˙⟡ river client setup

import (
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

func NewClient(db *pgxpool.Pool, workers *river.Workers, periodic []*river.PeriodicJob) (*river.Client[pgx.Tx], error) {
	return river.NewClient(riverpgxv5.New(db), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 10},
		},
		Workers:      workers,
		PeriodicJobs: periodic,
	})
}
