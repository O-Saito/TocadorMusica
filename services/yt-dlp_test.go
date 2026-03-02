package services

import "testing"

func TestExtractVideoID(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{"watch with v", "https://youtube.com/watch?v=dQw4w9WgXcQ", "dQw4w9WgXcQ"},
		{"watch with v and params", "https://youtube.com/watch?v=abc123&list=xyz", "abc123"},
		{"watch with extra params", "https://youtube.com/watch?v=video123&t=30&list=playlist123", "video123"},
		{"short link", "https://youtu.be/dQw4w9WgXcQ", "dQw4w9WgXcQ"},
		{"short link with params", "https://youtu.be/abc123?t=30", "abc123"},
		{"short link with embed", "https://youtu.be/xyz123?rel=0", "xyz123"},
		{"invalid domain", "https://example.com/video=dQw4w9WgXcQ", ""},
		{"empty url", "", ""},
		{"no v param", "https://youtube.com/watch?list=playlist123", ""},
		{"v param empty", "https://youtube.com/watch?v=", ""},
		{"http watch", "http://youtube.com/watch?v=test123", "test123"},
		{"http short", "http://youtu.be/test123", "test123"},
		{"youtube music", "https://music.youtube.com/watch?v=abc123", "abc123"},
		{"gaming subdomain", "https://gaming.youtube.com/watch?v=abc123", "abc123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractVideoID(tt.url)
			if got != tt.want {
				t.Errorf("extractVideoID(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestGetAudioURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{"valid watch url", "https://youtube.com/watch?v=dQw4w9WgXcQ", "https://youtube.com/watch?v=dQw4w9WgXcQ", false},
		{"valid short url", "https://youtu.be/dQw4w9WgXcQ", "https://youtube.com/watch?v=dQw4w9WgXcQ", false},
		{"invalid url", "https://example.com", "", true},
		{"empty url", "", "", true},
		{"url without v param", "https://youtube.com/playlist?list=123", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetAudioURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAudioURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetAudioURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertEntry(t *testing.T) {
	tests := []struct {
		name        string
		entry       ytDlpEntry
		wantTitle   string
		wantID      string
		wantURL     string
		wantDescLen int // expected length of truncated description
	}{
		{
			name: "full entry",
			entry: ytDlpEntry{
				ID:          "test123",
				Title:       "Test Video Title",
				Duration:    180.5,
				Thumbnail:   "https://example.com/thumb.jpg",
				Description: "This is a test description",
			},
			wantTitle:   "Test Video Title",
			wantID:      "test123",
			wantURL:     "https://youtube.com/watch?v=test123",
			wantDescLen: 26,
		},
		{
			name: "long description truncation",
			entry: ytDlpEntry{
				ID:          "long123",
				Title:       "Long Video",
				Duration:    300,
				Thumbnail:   "",
				Description: "This is a very long description that should be truncated to 200 characters when converted to a Track struct for display purposes in the queue. It is meant to test the truncation logic in the convertEntry function which should cut off at 200 characters and add ellipsis.",
			},
			wantTitle:   "Long Video",
			wantID:      "long123",
			wantURL:     "https://youtube.com/watch?v=long123",
			wantDescLen: 203, // 200 + "..."
		},
		{
			name: "empty thumbnail uses default",
			entry: ytDlpEntry{
				ID:       "nod thumb",
				Title:    "No Thumbnail",
				Duration: 60,
			},
			wantTitle: "No Thumbnail",
			wantID:    "nod thumb",
			wantURL:   "https://youtube.com/watch?v=nod thumb",
		},
		{
			name: "zero duration",
			entry: ytDlpEntry{
				ID:       "zero",
				Title:    "Zero Duration",
				Duration: 0,
			},
			wantTitle: "Zero Duration",
			wantID:    "zero",
			wantURL:   "https://youtube.com/watch?v=zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			track := convertEntry(tt.entry)

			if track.Title != tt.wantTitle {
				t.Errorf("Title = %v, want %v", track.Title, tt.wantTitle)
			}
			if track.ID != tt.wantID {
				t.Errorf("ID = %v, want %v", track.ID, tt.wantID)
			}
			if track.URL != tt.wantURL {
				t.Errorf("URL = %v, want %v", track.URL, tt.wantURL)
			}
			if track.Duration != int(tt.entry.Duration) {
				t.Errorf("Duration = %v, want %v", track.Duration, int(tt.entry.Duration))
			}

			// Check description truncation
			if tt.entry.Description != "" && len(track.Description) != tt.wantDescLen {
				t.Errorf("Description length = %v, want %v", len(track.Description), tt.wantDescLen)
			}

			// Check default thumbnail
			if tt.entry.Thumbnail == "" && track.Thumbnail == "" {
				t.Error("Expected default thumbnail when empty")
			}
		})
	}
}
