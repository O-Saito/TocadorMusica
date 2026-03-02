package models

import "testing"

func TestTrack_DurationFormatted(t *testing.T) {
	tests := []struct {
		name     string
		duration int
		want     string
	}{
		{"zero", 0, "0:00"},
		{"normal", 225, "3:45"},
		{"exactly one minute", 60, "1:00"},
		{"large", 3600, "60:00"},
		{"single digit seconds", 65, "1:05"},
		{"double digit seconds", 125, "2:05"},
		{"59 seconds", 59, "0:59"},
		{"very large", 3661, "61:01"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			track := Track{Duration: tt.duration}
			got := track.DurationFormatted()
			if got != tt.want {
				t.Errorf("DurationFormatted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrack_DurationFormatted_MinuteBoundaries(t *testing.T) {
	tests := []struct {
		name     string
		duration int
		want     string
	}{
		{"one second", 1, "0:01"},
		{"59 seconds", 59, "0:59"},
		{"60 seconds", 60, "1:00"},
		{"61 seconds", 61, "1:01"},
		{"119 seconds", 119, "1:59"},
		{"120 seconds", 120, "2:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			track := Track{Duration: tt.duration}
			got := track.DurationFormatted()
			if got != tt.want {
				t.Errorf("DurationFormatted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrack_StructFields(t *testing.T) {
	track := Track{
		ID:          "test-id",
		Title:       "Test Title",
		URL:         "https://youtube.com/watch?v=test-id",
		Duration:    180,
		Thumbnail:   "https://example.com/thumb.jpg",
		Description: "Test description",
	}

	if track.ID != "test-id" {
		t.Errorf("ID = %v, want test-id", track.ID)
	}
	if track.Title != "Test Title" {
		t.Errorf("Title = %v, want Test Title", track.Title)
	}
	if track.URL != "https://youtube.com/watch?v=test-id" {
		t.Errorf("URL = %v, want https://youtube.com/watch?v=test-id", track.URL)
	}
	if track.Duration != 180 {
		t.Errorf("Duration = %v, want 180", track.Duration)
	}
	if track.Thumbnail != "https://example.com/thumb.jpg" {
		t.Errorf("Thumbnail = %v, want https://example.com/thumb.jpg", track.Thumbnail)
	}
	if track.Description != "Test description" {
		t.Errorf("Description = %v, want Test description", track.Description)
	}
}
