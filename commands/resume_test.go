package commands

import (
	"testing"

	"tocadormusica/config"
	"tocadormusica/models"
	"tocadormusica/services"
)

func TestResumeCommand_Name(t *testing.T) {
	cmd := &ResumeCommand{}
	if cmd.Name() != "resume" {
		t.Errorf("Name() = %v, want resume", cmd.Name())
	}
}

func TestResumeCommand_Description(t *testing.T) {
	cmd := &ResumeCommand{}
	desc := cmd.Description()
	if desc == "" {
		t.Error("Description() returned empty string")
	}
}

func TestResumeCommand_Execute(t *testing.T) {
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

	cmd := &ResumeCommand{}
	err = cmd.Execute(ctx, []string{})
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
}
