package domain

import (
	"context"
	"io"
	"testing"
	"time"

	"tocadormusica/config"
	"tocadormusica/logger"
)

type mockPlayerPerfil struct {
	playing bool
	volume  float64
}

func (m *mockPlayerPerfil) Play(reader io.Reader) error {
	m.playing = true
	return nil
}

func (m *mockPlayerPerfil) PlayURL(url string, sampleRate int) error {
	m.playing = true
	return nil
}

func (m *mockPlayerPerfil) Pause()  { m.playing = false }
func (m *mockPlayerPerfil) Resume() { m.playing = true }
func (m *mockPlayerPerfil) Stop()   { m.playing = false }
func (m *mockPlayerPerfil) SetVolume(volume float64) {
	m.volume = volume
}

func (m *mockPlayerPerfil) Volume() float64 {
	return m.volume
}

func (m *mockPlayerPerfil) IsPlaying() bool { return m.playing }

type mockInputHandlerPerfil struct {
	inputChan chan string
}

func (m *mockInputHandlerPerfil) Input() <-chan string {
	return m.inputChan
}

func (m *mockInputHandlerPerfil) Close() {
	close(m.inputChan)
}

type mockOutputHandlerPerfil struct {
	messages []string
}

func (m *mockOutputHandlerPerfil) Display(message string) {
	m.messages = append(m.messages, message)
}

func (m *mockOutputHandlerPerfil) RequestInput(prompt string) <-chan string {
	ch := make(chan string)
	go func() {
		time.Sleep(10 * time.Millisecond)
		ch <- "mock input"
	}()
	return ch
}

func (m *mockOutputHandlerPerfil) DisplayOptions(options []string) <-chan int {
	ch := make(chan int)
	go func() {
		time.Sleep(10 * time.Millisecond)
		ch <- 0
	}()
	return ch
}

func (m *mockOutputHandlerPerfil) FindUnknownCommand() {}

type mockYouTubeService struct{}

func (m *mockYouTubeService) ParseURL(ctx context.Context, url string) (Track, error) {
	return NewTrackFromYouTube(url, "Test Track", "Description", "http://audio.url"), nil
}

func (m *mockYouTubeService) Search(ctx context.Context, query string, maxResults int) ([]SearchResult, error) {
	return []SearchResult{
		{Title: "Result 1", URL: "http://yt.com/1", Duration: "3:45"},
		{Title: "Result 2", URL: "http://yt.com/2", Duration: "4:30"},
	}, nil
}

type mockConfig struct{}

func (m *mockConfig) GetProfile(profileName string) (config.GlobalConfig, config.ProfileConfig) {
	return config.GlobalConfig{
			LogLevel:     "debug",
			MaxQueueSize: 500,
			SampleRate:   44100,
		}, config.ProfileConfig{
			Volume:        0.5,
			SearchResults: 10,
		}
}

func (m *mockConfig) SetVolume(profileName string, volume float64)     {}
func (m *mockConfig) SetSearchResults(profileName string, results int) {}
func (m *mockConfig) Save() error                                      { return nil }
func (m *mockConfig) Validate() error                                  { return nil }

func TestPerfil_New(t *testing.T) {
	queue := NewQueue(100)
	player := &mockPlayerPerfil{}
	input := &mockInputHandlerPerfil{inputChan: make(chan string)}
	output := &mockOutputHandlerPerfil{}
	ytSvc := &mockYouTubeService{}
	cfg := &mockConfig{}
	log := &mockLogger{}

	perfil := NewPerfil("test-perfil", queue, player, input, output, ytSvc, cfg, log, nil)

	if perfil.Name() != "test-perfil" {
		t.Errorf("expected name 'test-perfil', got '%s'", perfil.Name())
	}
}

func TestPerfil_Components(t *testing.T) {
	queue := NewQueue(100)
	player := &mockPlayerPerfil{}
	input := &mockInputHandlerPerfil{inputChan: make(chan string)}
	output := &mockOutputHandlerPerfil{}
	ytSvc := &mockYouTubeService{}
	cfg := &mockConfig{}
	log := &mockLogger{}

	perfil := NewPerfil("test", queue, player, input, output, ytSvc, cfg, log, nil)

	if perfil.Queue() == nil {
		t.Error("expected queue to be set")
	}
	if perfil.Player() == nil {
		t.Error("expected player to be set")
	}
}

func TestPerfil_Start(t *testing.T) {
	queue := NewQueue(100)
	player := &mockPlayerPerfil{}
	input := &mockInputHandlerPerfil{inputChan: make(chan string)}
	output := &mockOutputHandlerPerfil{}
	ytSvc := &mockYouTubeService{}
	cfg := &mockConfig{}
	log := &mockLogger{}

	perfil := NewPerfil("test", queue, player, input, output, ytSvc, cfg, log, nil)
	ctx, cancel := context.WithCancel(context.Background())

	err := perfil.Start(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	cancel()
	perfil.Wait()
}

func TestPerfil_Shutdown(t *testing.T) {
	queue := NewQueue(100)
	player := &mockPlayerPerfil{}
	input := &mockInputHandlerPerfil{inputChan: make(chan string)}
	output := &mockOutputHandlerPerfil{}
	ytSvc := &mockYouTubeService{}
	cfg := &mockConfig{}
	log := &mockLogger{}

	perfil := NewPerfil("test", queue, player, input, output, ytSvc, cfg, log, nil)
	ctx, cancel := context.WithCancel(context.Background())

	perfil.Start(ctx)

	time.Sleep(50 * time.Millisecond)
	cancel()

	perfil.Wait()
}

func TestPerfil_InputRouting(t *testing.T) {
	queue := NewQueue(100)
	player := &mockPlayerPerfil{}
	input := &mockInputHandlerPerfil{inputChan: make(chan string, 1)}
	output := &mockOutputHandlerPerfil{}
	ytSvc := &mockYouTubeService{}
	cfg := &mockConfig{}
	log := &mockLogger{}

	perfil := NewPerfil("test", queue, player, input, output, ytSvc, cfg, log, nil)
	ctx, cancel := context.WithCancel(context.Background())

	perfil.Start(ctx)

	input.inputChan <- "test command"

	time.Sleep(50 * time.Millisecond)

	cancel()
	perfil.Wait()
}

type mockLogger struct{}

func (m *mockLogger) Debug(msg string, args ...interface{}) {}
func (m *mockLogger) Info(msg string, args ...interface{})  {}
func (m *mockLogger) Warn(msg string, args ...interface{})  {}
func (m *mockLogger) Error(msg string, args ...interface{}) {}
func (m *mockLogger) WithProfile(name string) logger.Logger { return m }
