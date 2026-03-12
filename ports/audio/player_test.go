package audio

import (
	"io"
	"testing"
)

func TestPlayer_Interface(t *testing.T) {
	var _ Player = &mockPlayer{}
}

type mockPlayer struct {
	playing bool
	volume  float64
}

func (m *mockPlayer) Play(reader io.Reader) error {
	m.playing = true
	return nil
}

func (m *mockPlayer) PlayURL(url string, sampleRate int) error {
	m.playing = true
	return nil
}

func (m *mockPlayer) Pause() {
	m.playing = false
}

func (m *mockPlayer) Resume() {
	m.playing = true
}

func (m *mockPlayer) Stop() {
	m.playing = false
}

func (m *mockPlayer) SetVolume(volume float64) {
	m.volume = volume
}

func (m *mockPlayer) Volume() float64 {
	return m.volume
}

func (m *mockPlayer) IsPlaying() bool {
	return m.playing
}

func TestPlayer_MockPlay(t *testing.T) {
	p := &mockPlayer{}

	err := p.Play(nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !p.IsPlaying() {
		t.Error("expected playing after Play")
	}
}

func TestPlayer_MockPause(t *testing.T) {
	p := &mockPlayer{}
	p.Play(nil)

	p.Pause()

	if p.IsPlaying() {
		t.Error("expected not playing after Pause")
	}
}

func TestPlayer_MockResume(t *testing.T) {
	p := &mockPlayer{}
	p.Play(nil)
	p.Pause()

	p.Resume()

	if !p.IsPlaying() {
		t.Error("expected playing after Resume")
	}
}

func TestPlayer_MockStop(t *testing.T) {
	p := &mockPlayer{}
	p.Play(nil)

	p.Stop()

	if p.IsPlaying() {
		t.Error("expected not playing after Stop")
	}
}

func TestPlayer_MockSetVolume(t *testing.T) {
	p := &mockPlayer{}

	p.SetVolume(0.5)
	if p.volume != 0.5 {
		t.Errorf("expected volume 0.5, got %f", p.volume)
	}

	p.SetVolume(1.0)
	if p.volume != 1.0 {
		t.Errorf("expected volume 1.0, got %f", p.volume)
	}
}
