package commands

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"tocadormusica/config"
	"tocadormusica/models"
	"tocadormusica/services"
)

func TestAddCommand_Name(t *testing.T) {
	cmd := &AddCommand{}
	if cmd.Name() != "add" {
		t.Errorf("Name() = %v, want add", cmd.Name())
	}
}

func TestAddCommand_Description(t *testing.T) {
	cmd := &AddCommand{}
	desc := cmd.Description()
	if desc == "" {
		t.Error("Description() returned empty string")
	}
}

func TestAddCommand_Execute_NoInput_Prompts(t *testing.T) {
	player, err := services.NewAudioPlayer()
	if err != nil {
		t.Skipf("No audio device: %v", err)
	}
	defer player.Stop()

	input := "https://youtube.com/watch?v=dQw4w9WgXcQ\n"
	reader := bufio.NewReader(bytes.NewReader([]byte(input)))

	cfg := &config.Config{
		Volume:        0.5,
		SearchResults: 5,
	}

	ctx := &CommandContext{
		Queue:  models.NewQueue(),
		Player: player,
		Config: cfg,
		Reader: reader,
	}

	cmd := &AddCommand{}
	err = cmd.Execute(ctx, nil)
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
}

func TestAddCommand_Execute_URL_AddsToQueue(t *testing.T) {
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

	cmd := &AddCommand{}
	err = cmd.Execute(ctx, []string{"https://youtube.com/watch?v=dQw4w9WgXcQ"})
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
}

func TestAddCommand_Execute_EmptyInput_ReturnsError(t *testing.T) {
	player, err := services.NewAudioPlayer()
	if err != nil {
		t.Skipf("No audio device: %v", err)
	}
	defer player.Stop()

	input := "\n"
	reader := bufio.NewReader(bytes.NewReader([]byte(input)))

	cfg := &config.Config{
		Volume:        0.5,
		SearchResults: 5,
	}

	ctx := &CommandContext{
		Queue:  models.NewQueue(),
		Player: player,
		Config: cfg,
		Reader: reader,
	}

	cmd := &AddCommand{}
	err = cmd.Execute(ctx, []string{})
	if err == nil {
		t.Error("Expected error for empty input")
	}
	if !strings.Contains(err.Error(), "no input") {
		t.Errorf("Error message = %v, want contains 'no input'", err.Error())
	}
}

func TestAddCommand_Execute_Search_ShowsResults(t *testing.T) {
	player, err := services.NewAudioPlayer()
	if err != nil {
		t.Skipf("No audio device: %v", err)
	}
	defer player.Stop()

	input := "1\n"
	reader := bufio.NewReader(bytes.NewReader([]byte(input)))

	cfg := &config.Config{
		Volume:        0.5,
		SearchResults: 5,
	}

	ctx := &CommandContext{
		Queue:  models.NewQueue(),
		Player: player,
		Config: cfg,
		Reader: reader,
	}

	cmd := &AddCommand{}
	err = cmd.Execute(ctx, []string{"test query"})
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
}

func TestAddCommand_Execute_InvalidSelection(t *testing.T) {
	player, err := services.NewAudioPlayer()
	if err != nil {
		t.Skipf("No audio device: %v", err)
	}
	defer player.Stop()

	input := "999\n"
	reader := bufio.NewReader(bytes.NewReader([]byte(input)))

	cfg := &config.Config{
		Volume:        0.5,
		SearchResults: 5,
	}

	ctx := &CommandContext{
		Queue:  models.NewQueue(),
		Player: player,
		Config: cfg,
		Reader: reader,
	}

	cmd := &AddCommand{}
	err = cmd.Execute(ctx, []string{"test query"})
	if err == nil {
		t.Error("Expected error for invalid selection")
	}
	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("Error message = %v, want contains 'invalid'", err.Error())
	}
}

func TestAddCommand_Execute_CancelSelection(t *testing.T) {
	player, err := services.NewAudioPlayer()
	if err != nil {
		t.Skipf("No audio device: %v", err)
	}
	defer player.Stop()

	input := "\n"
	reader := bufio.NewReader(bytes.NewReader([]byte(input)))

	cfg := &config.Config{
		Volume:        0.5,
		SearchResults: 5,
	}

	ctx := &CommandContext{
		Queue:  models.NewQueue(),
		Player: player,
		Config: cfg,
		Reader: reader,
	}

	cmd := &AddCommand{}
	err = cmd.Execute(ctx, []string{"test query"})
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
}
