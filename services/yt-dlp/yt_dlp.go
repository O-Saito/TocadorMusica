package ytdlp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"tocadormusica/domain"
	"tocadormusica/logger"
)

type CommandRunner interface {
	Run(ctx context.Context, name string, args ...string) (string, error)
}

type youtubeService struct {
	cmdRunner CommandRunner
	log       logger.Logger
}

type realCommandRunner struct{}

func (r *realCommandRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return stderr.String(), err
	}
	return stdout.String(), nil
}

func New() domain.YouTubeService {
	return &youtubeService{
		cmdRunner: &realCommandRunner{},
		log:       nil,
	}
}

func NewWithRunner(runner CommandRunner) domain.YouTubeService {
	return &youtubeService{
		cmdRunner: runner,
		log:       nil,
	}
}

func NewWithRunnerAndLogger(runner CommandRunner, log logger.Logger) domain.YouTubeService {
	if runner == nil {
		runner = &realCommandRunner{}
	}
	return &youtubeService{
		cmdRunner: runner,
		log:       log,
	}
}

func findFirstJSONLine(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "{") {
			return line
		}
	}
	return ""
}

func (s *youtubeService) logError(msg string) {
	if s.log != nil {
		s.log.Error(msg)
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

		jsonLine := findFirstJSONLine(output)
		if jsonLine == "" {
			errMsg := strings.TrimSpace(output)
			if errMsg == "" {
				errMsg = err.Error()
			}
			s.logError("yt-dlp error: " + errMsg)
			return domain.Track{}, fmt.Errorf("yt-dlp error: %s", errMsg)
		}

		return domain.Track{}, fmt.Errorf("failed to execute yt-dlp: %w", err)
	}

	jsonLine := findFirstJSONLine(output)
	if jsonLine == "" {
		s.logError("yt-dlp no valid JSON: " + output)
		return domain.Track{}, fmt.Errorf("no valid JSON in yt-dlp output: %s", output)
	}

	var video ytDlpVideo
	if err := json.Unmarshal([]byte(jsonLine), &video); err != nil {
		return domain.Track{}, fmt.Errorf("failed to parse json: %w", err)
	}

	return domain.NewTrackFromYouTube(
		video.WebpageURL,
		video.Title,
		video.Description,
		"",
	), nil
}

func (s *youtubeService) GetAudioURL(ctx context.Context, url string) (string, error) {
	output, err := s.cmdRunner.Run(ctx, "yt-dlp",
		"--no-warnings",
		"--skip-download",
		"--get-url",
		"-f", "bestaudio",
		url)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("timeout: %w", ctx.Err())
		}

		errMsg := strings.TrimSpace(output)
		if errMsg == "" {
			errMsg = err.Error()
		}
		s.logError("yt-dlp error getting audio URL: " + errMsg)
		return "", fmt.Errorf("yt-dlp error: %s", errMsg)
	}

	output = strings.TrimSpace(output)
	if output == "" {
		s.logError("yt-dlp returned empty audio URL")
		return "", fmt.Errorf("no audio URL returned")
	}

	return output, nil
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

		jsonLine := findFirstJSONLine(output)
		if jsonLine == "" {
			errMsg := strings.TrimSpace(output)
			if errMsg == "" {
				errMsg = err.Error()
			}
			s.logError("yt-dlp search error: " + errMsg)
			return nil, fmt.Errorf("yt-dlp error: %s", errMsg)
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
