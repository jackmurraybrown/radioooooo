package jobs

// ✮⋆‧°—°‧⋆✮ ffmpeg loudness analysis — measures LUFS + true peak
// https://ffmpeg.org/ffmpeg-filters.html#loudnorm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"

	"github.com/riverqueue/river"
	"radioooooo/internal/media"
)

// ⊹ ࣪ ˖ -18 LUFS — between broadcast (-23) and streaming (-14)
// https://wiki.hydrogenaud.io/index.php?title=ReplayGain_2.0_specification
const TargetLUFS = -18.0

type LoudnessAnalysisWorker struct {
	river.WorkerDefaults[media.LoudnessAnalysisArgs]
	store *media.Store
}

func NewLoudnessAnalysisWorker(store *media.Store) *LoudnessAnalysisWorker {
	return &LoudnessAnalysisWorker{store: store}
}

func (w *LoudnessAnalysisWorker) Work(ctx context.Context, job *river.Job[media.LoudnessAnalysisArgs]) error {
	result, err := analyseLoudness(ctx, job.Args.FilePath)
	if err != nil {
		slog.Error("loudness: ffmpeg analysis failed", "media", job.Args.MediaID, "error", err)
		return err
	}

	duration, err := probeDuration(ctx, job.Args.FilePath)
	if err != nil {
		slog.Warn("loudness: duration probe failed", "media", job.Args.MediaID, "error", err)
	}

	if err := w.store.UpdateLoudness(ctx, job.Args.MediaID, result.InputI, result.InputTP, duration); err != nil {
		return fmt.Errorf("loudness: db update failed: %w", err)
	}

	slog.Info("loudness: analysis complete", "media", job.Args.MediaID, "lufs", result.InputI, "duration", duration)
	return nil
}

// ⋆˙⟡ ffmpeg loudnorm output (json from pass 1)
type loudnormResult struct {
	InputI  float64 `json:"input_i,string"`
	InputTP float64 `json:"input_tp,string"`
	InputLRA float64 `json:"input_lra,string"`
}

func analyseLoudness(ctx context.Context, filePath string) (*loudnormResult, error) {
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", filePath,
		"-af", "loudnorm=I=-18:TP=-1:LRA=11:print_format=json",
		"-f", "null", "/dev/null",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg: %w\n%s", err, output)
	}

	outStr := string(output)
	jsonStart := strings.LastIndex(outStr, "{")
	jsonEnd := strings.LastIndex(outStr, "}")
	if jsonStart == -1 || jsonEnd == -1 || jsonEnd <= jsonStart {
		return nil, fmt.Errorf("loudnorm: no json found in output")
	}

	var result loudnormResult
	if err := json.Unmarshal([]byte(outStr[jsonStart:jsonEnd+1]), &result); err != nil {
		return nil, fmt.Errorf("loudnorm: parse failed: %w", err)
	}

	return &result, nil
}

// . ݁₊ ✶ ffprobe for duration in seconds
func probeDuration(ctx context.Context, filePath string) (*int, error) {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "quiet",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filePath,
	)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe: %w", err)
	}
	secs, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return nil, fmt.Errorf("ffprobe: parse duration: %w", err)
	}
	d := int(secs)
	return &d, nil
}
