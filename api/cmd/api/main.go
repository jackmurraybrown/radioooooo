package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/riverqueue/river"
	"radioooooo/internal/analytics"
	"radioooooo/internal/broadcast"
	"radioooooo/internal/channel"
	"radioooooo/internal/config"
	"radioooooo/internal/database"
	"radioooooo/internal/episode"
	"radioooooo/internal/gapfill"
	"radioooooo/internal/ical"
	"radioooooo/internal/jobs"
	"radioooooo/internal/media"
	"radioooooo/internal/playlist"
	"radioooooo/internal/server"
	"radioooooo/internal/show"
	"radioooooo/internal/station"
	"radioooooo/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := database.MigrateRiver(ctx, db); err != nil {
		slog.Error("failed to run river migrations", "error", err)
		os.Exit(1)
	}

	if err := database.Migrate(cfg.DatabaseURL); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	slog.Info("migrations applied")

	workers := river.NewWorkers()
	river.AddWorker(workers, jobs.NewLoudnessAnalysisWorker(media.NewStore(db)))

	periodic := []*river.PeriodicJob{
		river.NewPeriodicJob(
			river.PeriodicInterval(30*time.Second),
			func() (river.JobArgs, *river.InsertOpts) {
				return analytics.CollectorArgs{}, nil
			},
			nil,
		),
		river.NewPeriodicJob(
			river.PeriodicInterval(7*24*time.Hour),
			func() (river.JobArgs, *river.InsertOpts) {
				return analytics.GeoUpdateArgs{}, nil
			},
			nil,
		),
		river.NewPeriodicJob(
			river.PeriodicInterval(24*time.Hour),
			func() (river.JobArgs, *river.InsertOpts) {
				return jobs.ShowExpansionArgs{}, nil
			},
			nil,
		),
		river.NewPeriodicJob(
			river.PeriodicInterval(15*time.Minute),
			func() (river.JobArgs, *river.InsertOpts) {
				return ical.SyncArgs{}, nil
			},
			nil,
		),
		river.NewPeriodicJob(
			river.PeriodicInterval(24*time.Hour),
			func() (river.JobArgs, *river.InsertOpts) {
				return gapfill.FillArgs{}, nil
			},
			nil,
		),
	}

	riverClient, err := jobs.NewClient(db, workers, periodic)
	if err != nil {
		slog.Error("failed to build river client", "error", err)
		os.Exit(1)
	}
	if err := riverClient.Start(ctx); err != nil {
		slog.Error("failed to start river client", "error", err)
		os.Exit(1)
	}
	slog.Info("river client started")

	// ⊹ ࣪ ˖ storage
	var store storage.Store
	switch cfg.StorageDriver {
	case "s3":
		store = storage.NewS3(storage.S3Config{
			Endpoint:  cfg.S3Endpoint,
			Bucket:    cfg.S3Bucket,
			Region:    cfg.S3Region,
			AccessKey: cfg.S3AccessKey,
			SecretKey: cfg.S3SecretKey,
		})
		slog.Info("storage: s3", "endpoint", cfg.S3Endpoint, "bucket", cfg.S3Bucket)
	default:
		ls, err := storage.NewLocal(cfg.StorageLocalRoot)
		if err != nil {
			slog.Error("storage: local init failed", "error", err)
			os.Exit(1)
		}
		store = ls
		slog.Info("storage: local", "root", cfg.StorageLocalRoot)
	}
	_ = store

	// ✮⋆‧° listener analytics
	icecastSource := analytics.NewIcecastSource(analytics.IcecastConfig{
		BaseURL:  cfg.IcecastURL,
		User:     cfg.IcecastAdminUser,
		Password: cfg.IcecastAdminPass,
		Mounts:   cfg.IcecastMounts,
	})

	var geoResolver *analytics.GeoResolver
	if cfg.GeoIPDatabasePath != "" {
		gr, err := analytics.NewGeoResolver(cfg.GeoIPDatabasePath)
		if err != nil {
			slog.Warn("analytics: geoip disabled", "error", err)
		} else {
			geoResolver = gr
			defer gr.Close()
			slog.Info("analytics: geoip loaded", "path", cfg.GeoIPDatabasePath)
		}
	}

	analyticsStore := analytics.NewStore(db)
	channelStore := channel.NewStore(db, cfg.EncryptionKey)

	river.AddWorker(workers, analytics.NewCollectorWorker(icecastSource, geoResolver, analyticsStore, channelStore))
	river.AddWorker(workers, jobs.NewShowExpansionWorker(show.NewStore(db), station.NewStore(db)))
	river.AddWorker(workers, ical.NewSyncWorker(ical.NewStore(db)))
	river.AddWorker(workers, gapfill.NewFillWorker(gapfill.NewStore(db)))
	river.AddWorker(workers, analytics.NewGeoUpdateWorker(cfg.GeoIPDatabasePath, geoResolver))

	// ⋆˙⟡ broadcast manager — one controller per channel
	var mgr *broadcast.Manager
	if cfg.LiquidsoapSocket != "" {
		liq, err := broadcast.Dial(cfg.LiquidsoapSocket)
		if err != nil {
			slog.Warn("broadcast: liquidsoap not reachable, controllers disabled", "error", err)
		} else {
			mgr = broadcast.NewManager(liq, channelStore, episode.NewStore(db), media.NewStore(db), playlist.NewStore(db))
			if err := mgr.Start(ctx); err != nil {
				slog.Error("broadcast: manager start failed", "error", err)
			}
		}
	}

	srv := server.New(cfg, db, icecastSource)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      srv,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// listen for shutdown signals before starting so we don't miss them
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server starting", "port", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	slog.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	if mgr != nil {
		mgr.Stop()
	}

	if err := riverClient.Stop(shutdownCtx); err != nil {
		slog.Error("river client stop failed", "error", err)
	}

	slog.Info("server stopped")
}
