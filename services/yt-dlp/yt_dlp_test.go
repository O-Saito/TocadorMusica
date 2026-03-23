package ytdlp

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockCommandRunner struct {
	output string
	err    error
	delay  time.Duration
}

func (m *mockCommandRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	return m.output, m.err
}

func TestParseURL_Success(t *testing.T) {
	jsonOutput := `{"webpage_url": "https://youtube.com/watch?v=test123", "title": "Test Video", "description": "Test description"}`

	runner := &mockCommandRunner{output: jsonOutput}
	svc := NewWithRunner(runner)

	track, err := svc.ParseURL(context.Background(), "https://youtube.com/watch?v=test123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if track.URL() != "https://youtube.com/watch?v=test123" {
		t.Errorf("expected URL https://youtube.com/watch?v=test123, got %s", track.URL())
	}
	if track.Title() != "Test Video" {
		t.Errorf("expected title Test Video, got %s", track.Title())
	}
}

func TestParseURL_Error(t *testing.T) {
	runner := &mockCommandRunner{err: errors.New("command failed")}
	svc := NewWithRunner(runner)

	_, err := svc.ParseURL(context.Background(), "https://youtube.com/watch?v=test")

	if err == nil {
		t.Error("expected error")
	}
}

func TestSearch_Success(t *testing.T) {
	jsonOutput := `{"webpage_url": "https://youtube.com/watch?v=1", "title": "Video 1", "duration": "3:45"}
{"webpage_url": "https://youtube.com/watch?v=2", "title": "Video 2", "duration": "4:30"}`

	runner := &mockCommandRunner{output: jsonOutput}
	svc := NewWithRunner(runner)

	results, err := svc.Search(context.Background(), "test query", 5)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestSearch_ResultFields(t *testing.T) {
	jsonOutput := `{"webpage_url": "https://youtube.com/watch?v=test", "title": "Test Video", "duration": "5:00"}`

	runner := &mockCommandRunner{output: jsonOutput}
	svc := NewWithRunner(runner)

	results, _ := svc.Search(context.Background(), "test", 1)

	if len(results) == 0 {
		t.Fatal("expected results")
	}

	first := results[0]
	if first.Title != "Test Video" {
		t.Errorf("expected title Test Video, got %s", first.Title)
	}
	if first.URL != "https://youtube.com/watch?v=test" {
		t.Errorf("expected URL https://youtube.com/watch?v=test, got %s", first.URL)
	}
	if first.Duration != "5:00" {
		t.Errorf("expected duration 5:00, got %s", first.Duration)
	}
}

func TestSearch_MaxResults(t *testing.T) {
	jsonOutput := `{"webpage_url": "https://youtube.com/watch?v=1", "title": "V1", "duration": "1:00"}
{"webpage_url": "https://youtube.com/watch?v=2", "title": "V2", "duration": "2:00"}
{"webpage_url": "https://youtube.com/watch?v=3", "title": "V3", "duration": "3:00"}`

	runner := &mockCommandRunner{output: jsonOutput}
	svc := NewWithRunner(runner)

	results, _ := svc.Search(context.Background(), "test", 3)

	if len(results) > 3 {
		t.Errorf("expected at most 3 results, got %d", len(results))
	}
}

func TestSearch_Error(t *testing.T) {
	runner := &mockCommandRunner{err: errors.New("command failed")}
	svc := NewWithRunner(runner)

	_, err := svc.Search(context.Background(), "test", 5)

	if err == nil {
		t.Error("expected error")
	}
}

func TestTimeout(t *testing.T) {
	runner := &mockCommandRunner{delay: 100 * time.Millisecond}
	svc := NewWithRunner(runner)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond)

	_, err := svc.ParseURL(ctx, "https://youtube.com/watch?v=test")

	if err == nil {
		t.Error("expected error on timeout")
	}
}

func TestParsePlaylist_Success(t *testing.T) {
	jsonOutput := `{"webpage_url": "https://youtube.com/watch?v=1", "title": "Track 1", "description": "Desc 1"}
{"webpage_url": "https://youtube.com/watch?v=2", "title": "Track 2", "description": "Desc 2"}`

	runner := &mockCommandRunner{output: jsonOutput}
	svc := NewWithRunner(runner)

	tracks, err := svc.ParsePlaylist(context.Background(), "https://youtube.com/playlist?list=test")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tracks) != 2 {
		t.Errorf("expected 2 tracks, got %d", len(tracks))
	}
	if tracks[0].Title() != "Track 1" {
		t.Errorf("expected first track title 'Track 1', got %s", tracks[0].Title())
	}
	if tracks[1].Title() != "Track 2" {
		t.Errorf("expected second track title 'Track 2', got %s", tracks[1].Title())
	}
}

func TestParsePlaylist_Error(t *testing.T) {
	runner := &mockCommandRunner{err: errors.New("command failed")}
	svc := NewWithRunner(runner)

	_, err := svc.ParsePlaylist(context.Background(), "https://youtube.com/playlist?list=test")

	if err == nil {
		t.Error("expected error")
	}
}
