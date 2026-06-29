package media

import "time"

//river job args — lives here to avoid import cycles (jobs imports media)
type LoudnessAnalysisArgs struct {
	MediaID  string `json:"media_id"`
	FilePath string `json:"file_path"`
}

func (LoudnessAnalysisArgs) Kind() string { return "loudness_analysis" }

// --- constants ✦ ✧ ✦ ---

const (
	FormatMP3 = "mp3"
	FormatAAC = "aac"
	FormatM4A = "m4a"

	DownloadStatusNotRequired = "not_required"
	DownloadStatusPending     = "pending"
	DownloadStatusDownloading = "downloading"
	DownloadStatusReady       = "ready"
	DownloadStatusFailed      = "failed"
)

// --- types ˚₊✧ ---

type Media struct {
	ID             string     `json:"id"                      db:"id"`
	StationID      string     `json:"stationId"               db:"station_id"`
	Title          string     `json:"title"                   db:"title"`
	Artist         *string    `json:"artist,omitempty"        db:"artist"`
	Duration       *int       `json:"duration,omitempty"      db:"duration"`
	ArtworkRef     *string    `json:"artworkRef,omitempty"    db:"artwork_ref"`
	FileFormat     *string    `json:"fileFormat,omitempty"    db:"file_format"`
	FileSizeBytes  *int64     `json:"fileSizeBytes,omitempty" db:"file_size_bytes"`
	SourceAdapter  string     `json:"sourceAdapter"           db:"source_adapter"`
	SourceRef      string     `json:"sourceRef"               db:"source_ref"`
	LocalRef       *string    `json:"localRef,omitempty"      db:"local_ref"`
	DownloadStatus string     `json:"downloadStatus"          db:"download_status"`
	DownloadError  *string    `json:"downloadError,omitempty" db:"download_error"`
	DownloadedAt   *time.Time `json:"downloadedAt,omitempty"  db:"downloaded_at"`
	LoudnessLUFS   *float64   `json:"loudnessLufs,omitempty"  db:"loudness_lufs"`
	TruePeakDB     *float64   `json:"truePeakDb,omitempty"    db:"true_peak_db"`
	CreatedAt      time.Time  `json:"createdAt"               db:"created_at"`
	UpdatedAt      time.Time  `json:"updatedAt"               db:"updated_at"`
}
