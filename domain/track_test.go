package domain

import (
	"testing"
)

func TestNewTrack(t *testing.T) {
	track := NewTrack("http://example.com", "Test Title", "Description", "http://audio.com", TrackTypeYouTube)

	if track.URL() != "http://example.com" {
		t.Errorf("expected URL 'http://example.com', got '%s'", track.URL())
	}
	if track.Title() != "Test Title" {
		t.Errorf("expected Title 'Test Title', got '%s'", track.Title())
	}
	if track.Description() != "Description" {
		t.Errorf("expected Description 'Description', got '%s'", track.Description())
	}
	if track.AudioURL() != "http://audio.com" {
		t.Errorf("expected AudioURL 'http://audio.com', got '%s'", track.AudioURL())
	}
	if track.Type != TrackTypeYouTube {
		t.Errorf("expected Type YouTube, got '%s'", track.Type)
	}
}

func TestTrack_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		track   Track
		wantErr bool
	}{
		{
			name: "valid YouTube track",
			track: Track{
				url:      "http://youtube.com/watch?v=abc",
				title:    "Test",
				Type:     TrackTypeYouTube,
				audioURL: "http://audio.com",
			},
			wantErr: false,
		},
		{
			name: "valid File track",
			track: Track{
				url:      "C:/Music/song.mp3",
				title:    "Song",
				Type:     TrackTypeFile,
				audioURL: "C:/Music/song.mp3",
			},
			wantErr: false,
		},
		{
			name: "empty URL",
			track: Track{
				url:      "",
				title:    "Test",
				Type:     TrackTypeYouTube,
				audioURL: "http://audio.com",
			},
			wantErr: true,
		},
		{
			name: "empty Title",
			track: Track{
				url:      "http://test.com",
				title:    "",
				Type:     TrackTypeYouTube,
				audioURL: "http://audio.com",
			},
			wantErr: true,
		},
		{
			name: "empty AudioURL",
			track: Track{
				url:      "http://test.com",
				title:    "Test",
				Type:     TrackTypeYouTube,
				audioURL: "",
			},
			wantErr: true,
		},
		{
			name: "empty Type",
			track: Track{
				url:      "http://test.com",
				title:    "Test",
				Type:     "",
				audioURL: "http://audio.com",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.track.IsValid()
			if (err != nil) != tt.wantErr {
				t.Errorf("IsValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTrackType_String(t *testing.T) {
	if TrackTypeYouTube.String() != "YouTube" {
		t.Errorf("expected YouTube, got %s", TrackTypeYouTube.String())
	}
	if TrackTypeFile.String() != "File" {
		t.Errorf("expected File, got %s", TrackTypeFile.String())
	}
}

func TestNewTrackFromYouTube(t *testing.T) {
	track := NewTrackFromYouTube("http://youtube.com/watch?v=abc", "YouTube Video", "Video Description", "http://audio.stream")

	if track.Type != TrackTypeYouTube {
		t.Errorf("expected Type YouTube, got %s", track.Type)
	}
	if track.URL() != "http://youtube.com/watch?v=abc" {
		t.Errorf("expected URL, got %s", track.URL())
	}
	if track.Title() != "YouTube Video" {
		t.Errorf("expected Title, got %s", track.Title())
	}
	if track.Description() != "Video Description" {
		t.Errorf("expected Description, got %s", track.Description())
	}
	if track.AudioURL() != "http://audio.stream" {
		t.Errorf("expected AudioURL, got %s", track.AudioURL())
	}
}

func TestNewTrackFromFile(t *testing.T) {
	track := NewTrackFromFile("C:/Music/song.mp3")

	if track.Type != TrackTypeFile {
		t.Errorf("expected Type File, got %s", track.Type)
	}
	if track.URL() != "C:/Music/song.mp3" {
		t.Errorf("expected URL, got %s", track.URL())
	}
	if track.Title() != "song.mp3" {
		t.Errorf("expected Title 'song.mp3', got '%s'", track.Title())
	}
	if track.Description() != "" {
		t.Errorf("expected empty Description, got '%s'", track.Description())
	}
	if track.AudioURL() != "C:/Music/song.mp3" {
		t.Errorf("expected AudioURL, got %s", track.AudioURL())
	}
}
