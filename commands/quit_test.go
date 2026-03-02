package commands

import (
	"testing"

	"tocadormusica/config"
	"tocadormusica/models"
	"tocadormusica/services"
)

func TestQuitCommand_Name(t *testing.T) {
	cmd := &QuitCommand{}
	if cmd.Name() != "quit" {
		t.Errorf("Name() = %v, want quit", cmd.Name())
	}
}

func TestQuitCommand_Description(t *testing.T) {
	cmd := &QuitCommand{}
	desc := cmd.Description()
	if desc == "" {
		t.Error("Description() returned empty string")
	}
}

func TestQuitCommand_Execute_ReturnsExitError(t *testing.T) {
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

	cmd := &QuitCommand{}
	err = cmd.Execute(ctx, []string{})
	if err == nil {
		t.Fatal("Expected ExitError")
	}
	if !IsExitError(err) {
		t.Errorf("Expected ExitError, got %v", err)
	}
}
