package ytdlp

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"tocadormusica/domain"
)

type CommandRunner interface {
	Run(ctx context.Context, name string, args ...string) (string, error)
}

type youtubeService struct {
	cmdRunner CommandRunner
}

type realCommandRunner struct{}

func (r *realCommandRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func New() domain.YouTubeService {
	return &youtubeService{
		cmdRunner: &realCommandRunner{},
	}
}

func NewWithRunner(runner CommandRunner) domain.YouTubeService {
	return &youtubeService{
		cmdRunner: runner,
	}
}

func (s *youtubeService) ParseURL(ctx context.Context, url string) (domain.Track, error) {
	output, err := s.cmdRunner.Run(ctx, "yt-dlp",
		"--no-warnings",
		"--skip-download",
		"--dump-json",
		url)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return domain.Track{}, fmt.Errorf("timeout: %w", ctx.Err())
		}
		return domain.Track{}, fmt.Errorf("failed to execute yt-dlp: %w", err)
	}

	var video ytDlpVideo
	if err := json.Unmarshal([]byte(output), &video); err != nil {
		return domain.Track{}, fmt.Errorf("failed to parse json: %w", err)
	}

	audioURL := findAudioURL(video.Formats)

	return domain.NewTrackFromYouTube(
		video.WebpageURL,
		video.Title,
		video.Description,
		audioURL,
	), nil
}

func durationToString(d interface{}) string {
	switch v := d.(type) {
	case float64:
		seconds := int(v)
		minutes := seconds / 60
		secs := seconds % 60
		return fmt.Sprintf("%d:%02d", minutes, secs)
	case string:
		return v
	default:
		return ""
	}
}

func (s *youtubeService) Search(ctx context.Context, query string, maxResults int) ([]domain.SearchResult, error) {
	output, err := s.cmdRunner.Run(ctx, "yt-dlp",
		"--no-warnings",
		"--no-playlist",
		"--no-check-certificate",
		"--geo-bypass",
		"--flat-playlist",
		"--skip-download",
		"--dump-json",
		fmt.Sprintf("ytsearch%d:%s", maxResults, query))
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("timeout: %w", ctx.Err())
		}
		return nil, fmt.Errorf("failed to execute yt-dlp: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	results := make([]domain.SearchResult, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var video ytDlpVideo
		if err := json.Unmarshal([]byte(line), &video); err != nil {
			continue
		}

		results = append(results, domain.SearchResult{
			Title:    video.Title,
			URL:      video.WebpageURL,
			Duration: durationToString(video.Duration),
		})
	}

	return results, nil
}

func findAudioURL(formats []ytDlpFormat) string {
	for _, f := range formats {
		if f.Resolution == "audio only" {
			return f.URL
		}
	}
	return ""
}

type ytDlpVideo struct {
	WebpageURL  string        `json:"webpage_url"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Formats     []ytDlpFormat `json:"formats"`
	Duration    interface{}   `json:"duration"`
}

type ytDlpFormat struct {
	URL        string `json:"url"`
	Resolution string `json:"resolution"`
}
