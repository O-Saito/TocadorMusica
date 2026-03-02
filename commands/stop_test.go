package commands

import (
	"testing"

	"tocadormusica/config"
	"tocadormusica/models"
	"tocadormusica/services"
)

func TestStopCommand_Name(t *testing.T) {
	cmd := &StopCommand{}
	if cmd.Name() != "stop" {
		t.Errorf("Name() = %v, want stop", cmd.Name())
	}
}

func TestStopCommand_Description(t *testing.T) {
	cmd := &StopCommand{}
	desc := cmd.Description()
	if desc == "" {
		t.Error("Description() returned empty string")
	}
}

func TestStopCommand_Execute(t *testing.T) {
	player, err := services.NewAudioPlayer()
	if err != nil {
		t.Skipf("No audio device: %v", err)
	}
	defer player.Stop()

	cfg := &config.Config{
		Volume:        0.5,
		SearchResults: 5,
	}

	queue := models.NewQueue()
	queue.Add([]models.Track{{ID: "1", Title: "Test Track"}})

	ctx := &CommandContext{
		Queue:  queue,
		Player: player,
		Config: cfg,
		Reader: nil,
	}

	cmd := &StopCommand{}
	err = cmd.Execute(ctx, []string{})
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if !queue.IsEmpty() {
		t.Error("Queue should be empty after stop")
	}
}
