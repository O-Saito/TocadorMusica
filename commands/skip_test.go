package commands

import (
	"testing"

	"tocadormusica/config"
	"tocadormusica/models"
	"tocadormusica/services"
)

func TestSkipCommand_Name(t *testing.T) {
	cmd := &SkipCommand{}
	if cmd.Name() != "skip" {
		t.Errorf("Name() = %v, want skip", cmd.Name())
	}
}

func TestSkipCommand_Description(t *testing.T) {
	cmd := &SkipCommand{}
	desc := cmd.Description()
	if desc == "" {
		t.Error("Description() returned empty string")
	}
}

func TestSkipCommand_Execute(t *testing.T) {
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

	cmd := &SkipCommand{}
	err = cmd.Execute(ctx, []string{})
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
}
