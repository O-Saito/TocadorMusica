package commands

import (
	"testing"

	"tocadormusica/config"
	"tocadormusica/models"
	"tocadormusica/services"
)

func TestPauseCommand_Name(t *testing.T) {
	cmd := &PauseCommand{}
	if cmd.Name() != "pause" {
		t.Errorf("Name() = %v, want pause", cmd.Name())
	}
}

func TestPauseCommand_Description(t *testing.T) {
	cmd := &PauseCommand{}
	desc := cmd.Description()
	if desc == "" {
		t.Error("Description() returned empty string")
	}
}

func TestPauseCommand_Execute(t *testing.T) {
	player, err := services.NewAudioPlayer()
	if err != nil {
		t.Skipf("No audio device: %v", err)
	}
	defer player.Stop()

	cfg := &config.Config{
		Volume:        0.5,
		SearchResults: 5,
	}

	ctx := &CommandContext{
		Queue:  models.NewQueue(),
		Player: player,
		Config: cfg,
		Reader: nil,
	}

	cmd := &PauseCommand{}
	err = cmd.Execute(ctx, []string{})
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
}
