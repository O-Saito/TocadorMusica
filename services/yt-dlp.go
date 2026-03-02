package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"tocadormusica/models"
)

const (
	ERROR_TYPE = iota
	VIDEO_TYPE
	PLAYLIST_TYPE
)

type Format struct {
	FormatID   string `json:"format_id"`
	URL        string `json:"url"`
	Resolution string `json:"resolution"`
	Acodec     string `json:"acodec"`
}

type ytDlpEntry struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Duration    float64      `json:"duration"`
	Thumbnail   string       `json:"thumbnail"`
	Description string       `json:"description"`
	URL         string       `json:"url"`
	Entries     []ytDlpEntry `json:"entries"`
	Formats     []Format     `json:"formats"`
	Type        string       `json:"_type"`
}

func getType(jsonStr string) int {
	if strings.Contains(jsonStr, "upload_date") {
		return VIDEO_TYPE
	}
	if strings.Contains(jsonStr, "_type") {
		return PLAYLIST_TYPE
	}
	return ERROR_TYPE
}

func getAudioURL(formats []Format) string {
	for _, f := range formats {
		if f.Resolution == "audio only" {
			return f.URL
		}
	}
	return ""
}

func ParseURL(url string) ([]models.Track, error) {
	Info("Parsing URL: %s", url)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "yt-dlp", "--no-warnings", "--skip-download", "--dump-json", url)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		Error("Failed to create pipe: %v", err)
		return nil, fmt.Errorf("failed to create pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		Error("Failed to start yt-dlp: %v", err)
		return nil, fmt.Errorf("failed to start yt-dlp: %w", err)
	}

	decoder := json.NewDecoder(stdout)
	var tracks []models.Track
	entryCount := 0

	for {
		var entry ytDlpEntry
		if err := decoder.Decode(&entry); err != nil {
			break
		}
		entryCount++

		if entry.Type == "playlist" && len(entry.Entries) > 0 {
			Info("Found playlist with %d entries", len(entry.Entries))
			for _, e := range entry.Entries {
				tracks = append(tracks, convertEntry(e))
			}
		} else {
			tracks = append(tracks, convertEntry(entry))
		}
	}

	err = cmd.Wait()
	if err != nil {
		Error("yt-dlp command error: %v", err)
	}

	if len(tracks) == 0 {
		Error("No tracks found for URL: %s", url)
		return nil, fmt.Errorf("no tracks found")
	}

	Info("Successfully parsed %d tracks", len(tracks))
	return tracks, nil
}

func convertEntry(e ytDlpEntry) models.Track {
	desc := e.Description
	if len(desc) > 200 {
		desc = desc[:200] + "..."
	}

	thumbnail := e.Thumbnail
	if thumbnail == "" {
		thumbnail = fmt.Sprintf("https://img.youtube.com/vi/%s/default.jpg", e.ID)
	}

	audioURL := e.URL

	Info("Converting entry: %s, Formats count: %d", e.Title, len(e.Formats))

	if len(e.Formats) > 0 {
		if url := getAudioURL(e.Formats); url != "" {
			audioURL = url
			Info("Found audio URL for: %s", e.Title)
		}
	} else {
		Info("No formats found for: %s, will use fallback", e.Title)
	}

	return models.Track{
		ID:          e.ID,
		Title:       e.Title,
		URL:         fmt.Sprintf("https://youtube.com/watch?v=%s", e.ID),
		AudioURL:    audioURL,
		Duration:    int(e.Duration),
		Thumbnail:   thumbnail,
		Description: desc,
	}
}

func GetAudioURL(videoURL string) (string, error) {
	videoID := extractVideoID(videoURL)
	if videoID == "" {
		return "", fmt.Errorf("invalid YouTube URL")
	}
	return fmt.Sprintf("https://youtube.com/watch?v=%s", videoID), nil
}

func IsURL(input string) bool {
	return strings.HasPrefix(input, "http://") ||
		strings.HasPrefix(input, "https://") ||
		strings.Contains(input, "youtube.com") ||
		strings.Contains(input, "youtu.be")
}

func Search(query string, maxResults int) ([]models.Track, error) {
	Info("Searching for: %s (max %d results)", query, maxResults)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "yt-dlp",
		"--no-warnings",
		"--skip-download",
		"--dump-json",
		fmt.Sprintf("ytsearch%d:%s", maxResults, query))

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		Error("Failed to create pipe: %v", err)
		return nil, fmt.Errorf("failed to create pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		Error("Failed to start yt-dlp: %v", err)
		return nil, fmt.Errorf("failed to start yt-dlp: %w", err)
	}

	decoder := json.NewDecoder(stdout)
	var tracks []models.Track

	for {
		var entry ytDlpEntry
		if err := decoder.Decode(&entry); err != nil {
			break
		}

		tracks = append(tracks, convertEntrySearch(entry))
	}

	err = cmd.Wait()
	if err != nil {
		Error("yt-dlp command error: %v", err)
	}

	if len(tracks) == 0 {
		Error("No results found for: %s", query)
		return nil, fmt.Errorf("no results found")
	}

	Info("Found %d results for: %s", len(tracks), query)
	return tracks, nil
}

func convertEntrySearch(e ytDlpEntry) models.Track {
	thumbnail := e.Thumbnail
	if thumbnail == "" {
		thumbnail = fmt.Sprintf("https://img.youtube.com/vi/%s/default.jpg", e.ID)
	}

	return models.Track{
		ID:        e.ID,
		Title:     e.Title,
		URL:       fmt.Sprintf("https://youtube.com/watch?v=%s", e.ID),
		AudioURL:  "", // Will be fetched when selected
		Duration:  int(e.Duration),
		Thumbnail: thumbnail,
	}
}

func extractVideoID(url string) string {
	url = strings.TrimSpace(url)

	if strings.Contains(url, "youtube.com/watch") {
		parts := strings.Split(url, "v=")
		if len(parts) > 1 {
			id := strings.Split(parts[1], "&")[0]
			return id
		}
	}

	if strings.Contains(url, "youtu.be/") {
		parts := strings.Split(url, "youtu.be/")
		if len(parts) > 1 {
			id := strings.Split(parts[1], "?")[0]
			return id
		}
	}

	return ""
}
