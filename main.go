package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"tocadormusica/commands"
	"tocadormusica/config"
	"tocadormusica/models"
	"tocadormusica/services"
)

func main() {
	if err := services.InitLogger(); err != nil {
		fmt.Printf("Warning: Could not initialize logger: %v\n", err)
	}
	services.Info("Application starting...")

	defer func() {
		if r := recover(); r != nil {
			services.Error("PANIC recovered: %v", r)
			fmt.Printf("\nPANIC recovered: %v\n", r)
			fmt.Println("Check logs folder for details")
		}
	}()

	fmt.Println("=== YouTube Music Player ===")
	fmt.Println()

	cfg, err := config.Load()
	if err != nil {
		services.Error("Error loading config: %v", err)
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	queue := models.NewQueue()
	player, err := services.NewAudioPlayer()
	if err != nil {
		services.Error("Error initializing audio: %v", err)
		fmt.Printf("Error initializing audio: %v\n", err)
		os.Exit(1)
	}

	player.SetVolume(cfg.Volume)

	go func() {
		for {
			if queue.IsEmpty() || player.IsPlaying() || player.IsPaused() {
				time.Sleep(500 * time.Millisecond)
				continue
			}

			track := queue.Next()
			if track == nil {
				time.Sleep(500 * time.Millisecond)
				continue
			}

			fmt.Printf("▶ Now playing: %s [%s]\n", track.Title, track.DurationFormatted())

			if err := player.Play(track.AudioURL); err != nil {
				services.Error("Error playing track: %v", err)
				fmt.Printf("Error playing: %v\n", err)
			}
		}
	}()

	reader := bufio.NewReader(os.Stdin)

	ctx := &commands.CommandContext{
		Queue:  queue,
		Player: player,
		Config: cfg,
		Reader: reader,
	}

	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}

		parts := strings.Fields(input)
		cmd := commands.Get(strings.ToLower(parts[0]))

		if cmd == nil {
			showCommands()
			continue
		}

		err := cmd.Execute(ctx, parts[1:])
		if commands.IsExitError(err) {
			services.Info("Application exiting")
			return
		}
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
}

func showCommands() {
	list := commands.List()
	fmt.Println("Available commands:")
	for _, c := range list {
		fmt.Printf("  - %-8s: %s\n", c.Name(), c.Description())
	}
}
