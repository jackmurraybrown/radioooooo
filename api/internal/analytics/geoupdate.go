package analytics

// ✮⋆‧° db-ip lite auto-updater — downloads the free country database monthly
// https://db-ip.com/db/lite.php
// no account or license key needed

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/riverqueue/river"
)

type GeoUpdateArgs struct{}

func (GeoUpdateArgs) Kind() string { return "geo_database_update" }

type GeoUpdateWorker struct {
	river.WorkerDefaults[GeoUpdateArgs]
	dbPath string
	geo    *GeoResolver
}

func NewGeoUpdateWorker(dbPath string, geo *GeoResolver) *GeoUpdateWorker {
	return &GeoUpdateWorker{dbPath: dbPath, geo: geo}
}

func (w *GeoUpdateWorker) Work(ctx context.Context, job *river.Job[GeoUpdateArgs]) error {
	now := time.Now()
	url := fmt.Sprintf("https://download.db-ip.com/free/dbip-country-lite-%d-%02d.mmdb.gz", now.Year(), now.Month())

	slog.Info("geo: downloading db-ip update", "url", url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("geo: build request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("geo: download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("geo: download status %d", resp.StatusCode)
	}

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("geo: gzip reader: %w", err)
	}
	defer gz.Close()

	// ⋆˙⟡ write to a temp file first, then rename — atomic swap
	tmpPath := w.dbPath + ".tmp"
	if err := os.MkdirAll(filepath.Dir(tmpPath), 0o755); err != nil {
		return fmt.Errorf("geo: mkdir: %w", err)
	}

	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("geo: create tmp: %w", err)
	}

	if _, err := io.Copy(f, gz); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("geo: write: %w", err)
	}
	f.Close()

	if err := os.Rename(tmpPath, w.dbPath); err != nil {
		return fmt.Errorf("geo: rename: %w", err)
	}

	// ⊹ ࣪ ˖ reload the geo resolver with the new database
	if w.geo != nil {
		newGeo, err := NewGeoResolver(w.dbPath)
		if err != nil {
			slog.Error("geo: reload failed, keeping old database", "error", err)
			return nil
		}
		oldDB := w.geo.db
		w.geo.db = newGeo.db
		oldDB.Close()
	}

	slog.Info("geo: database updated", "path", w.dbPath)
	return nil
}
