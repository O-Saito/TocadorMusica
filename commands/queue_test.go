package commands

import (
	"testing"

	"tocadormusica/config"
	"tocadormusica/models"
	"tocadormusica/services"
)

func TestQueueCommand_Name(t *testing.T) {
	cmd := &QueueCommand{}
	if cmd.Name() != "queue" {
		t.Errorf("Name() = %v, want queue", cmd.Name())
	}
}

func TestQueueCommand_Description(t *testing.T) {
	cmd := &QueueCommand{}
	desc := cmd.Description()
	if desc == "" {
		t.Error("Description() returned empty string")
	}
}

func TestQueueCommand_Execute_Empty(t *testing.T) {
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

	cmd := &QueueCommand{}
	err = cmd.Execute(ctx, []string{})
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
}

func TestQueueCommand_Execute_WithTracks(t *testing.T) {
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
	queue.Add([]models.Track{
		{ID: "1", Title: "Track 1", Duration: 180},
		{ID: "2", Title: "Track 2", Duration: 240},
	})

	ctx := &CommandContext{
		Queue:  queue,
		Player: player,
		Config: cfg,
		Reader: nil,
	}

	cmd := &QueueCommand{}
	err = cmd.Execute(ctx, []string{})
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
}
