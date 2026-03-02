package services

import (
	"sync"
	"testing"
)

var (
	testPlayer     *AudioPlayer
	testPlayerOnce sync.Once
	testPlayerErr  error
)

func getTestPlayer(t *testing.T) *AudioPlayer {
	testPlayerOnce.Do(func() {
		testPlayer, testPlayerErr = NewAudioPlayer()
	})
	if testPlayerErr != nil {
		t.Skipf("No audio device: %v", testPlayerErr)
	}
	return testPlayer
}

func TestNewAudioPlayer(t *testing.T) {
	player, err := NewAudioPlayer()
	if err != nil {
		t.Skipf("No audio device: %v", err)
	}
	if player == nil {
		t.Fatal("Expected player, got nil")
	}
	defer player.Stop()
}

func TestAudioPlayer_SetVolume_Normal(t *testing.T) {
	player := getTestPlayer(t)

	tests := []struct {
		name string
		want float64
	}{
		{"zero", 0.0},
		{"quarter", 0.25},
		{"half", 0.5},
		{"three_quarter", 0.75},
		{"full", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player.SetVolume(tt.want)
			if player.Volume() != tt.want {
				t.Errorf("Volume() = %v, want %v", player.Volume(), tt.want)
			}
		})
	}
}

func TestAudioPlayer_SetVolume_Bounds(t *testing.T) {
	player := getTestPlayer(t)

	tests := []struct {
		name string
		set  float64
		want float64
	}{
		{"negative_clamped", -0.5, 0.0},
		{"below_zero", -1.0, 0.0},
		{"above_one_clamped", 1.5, 1.0},
		{"above_one", 2.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player.SetVolume(tt.set)
			if player.Volume() != tt.want {
				t.Errorf("Volume() = %v, want %v", player.Volume(), tt.want)
			}
		})
	}
}

func TestAudioPlayer_IsPlaying_Initial(t *testing.T) {
	player := getTestPlayer(t)

	if player.IsPlaying() {
		t.Error("IsPlaying() = true, want false initially")
	}
}

func TestAudioPlayer_IsPaused_Initial(t *testing.T) {
	player := getTestPlayer(t)

	if player.IsPaused() {
		t.Error("IsPaused() = true, want false initially")
	}
}

func TestAudioPlayer_Stop(t *testing.T) {
	player := getTestPlayer(t)

	player.Stop()

	if player.IsPlaying() {
		t.Error("IsPlaying() = true, want false after Stop()")
	}
	if player.IsPaused() {
		t.Error("IsPaused() = true, want false after Stop()")
	}
}

func TestAudioPlayer_Pause(t *testing.T) {
	player := getTestPlayer(t)

	player.Pause()

	if !player.IsPaused() {
		t.Error("IsPaused() = false, want true after Pause()")
	}
}

func TestAudioPlayer_Resume(t *testing.T) {
	player := getTestPlayer(t)

	player.Pause()
	player.Resume()

	if player.IsPaused() {
		t.Error("IsPaused() = true, want false after Resume()")
	}
}
