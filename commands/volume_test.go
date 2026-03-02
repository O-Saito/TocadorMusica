package commands

import (
	"testing"

	"tocadormusica/config"
	"tocadormusica/models"
	"tocadormusica/services"
)

func TestVolumeCommand_Name(t *testing.T) {
	cmd := &VolumeCommand{}
	if cmd.Name() != "volume" {
		t.Errorf("Name() = %v, want volume", cmd.Name())
	}
}

func TestVolumeCommand_Description(t *testing.T) {
	cmd := &VolumeCommand{}
	desc := cmd.Description()
	if desc == "" {
		t.Error("Description() returned empty string")
	}
}

func TestVolumeCommand_Execute_Get(t *testing.T) {
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

	cmd := &VolumeCommand{}
	err = cmd.Execute(ctx, []string{})
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
}

func TestVolumeCommand_Execute_Set(t *testing.T) {
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

	cmd := &VolumeCommand{}
	err = cmd.Execute(ctx, []string{"75"})
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if player.Volume() != 0.75 {
		t.Errorf("Volume() = %v, want 0.75", player.Volume())
	}
}

func TestVolumeCommand_Execute_Invalid(t *testing.T) {
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

	cmd := &VolumeCommand{}
	err = cmd.Execute(ctx, []string{"invalid"})
	if err == nil {
		t.Error("Expected error for invalid volume")
	}
}

func TestVolumeCommand_Execute_OutOfRange(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"negative", []string{"-1"}},
		{"over_100", []string{"101"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			cmd := &VolumeCommand{}
			err = cmd.Execute(ctx, tt.args)
			if err == nil {
				t.Error("Expected error for out of range volume")
			}
		})
	}
}

func TestVolumeCommand_Execute_Boundary(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want float64
	}{
		{"zero", []string{"0"}, 0.0},
		{"fifty", []string{"50"}, 0.5},
		{"hundred", []string{"100"}, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			cmd := &VolumeCommand{}
			err = cmd.Execute(ctx, tt.args)
			if err != nil {
				t.Errorf("Execute() error = %v", err)
			}

			if player.Volume() != tt.want {
				t.Errorf("Volume() = %v, want %v", player.Volume(), tt.want)
			}
		})
	}
}
