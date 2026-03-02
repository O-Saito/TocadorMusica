package models

import "fmt"

type Track struct {
	ID          string
	Title       string
	URL         string // Video URL (for display in queue)
	AudioURL    string // Direct audio stream URL (for ffmpeg)
	Duration    int
	Thumbnail   string
	Description string
}

func (t Track) DurationFormatted() string {
	minutes := t.Duration / 60
	seconds := t.Duration % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}
