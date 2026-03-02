package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

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
	fmt.Println("Commands: add, volume, skip, pause, resume, stop, queue, quit")
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
	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}

		parts := strings.Fields(input)
		cmd := strings.ToLower(parts[0])

		switch cmd {
		case "add":
			input := strings.Join(parts[1:], " ")

			if len(parts) == 0 {
				fmt.Print("Paste YouTube URL or search: ")
				input, _ = reader.ReadString('\n')
				input = strings.TrimSpace(input)
			}

			if input == "" {
				fmt.Println("No input provided")
				continue
			}

			parseURLAddToQueue := func(url string) {
				fmt.Println("Fetching tracks...")
				// It's a URL - stream tracks from channel
				ch, err := services.ParseURL(url)
				if err != nil {
					services.Error("Error parsing URL: %v", err)
					fmt.Printf("Error: %v\n", err)
					return
				}

				// Add tracks to queue as they arrive
				trackCount := 0
				for track := range ch {
					queue.Add([]models.Track{track})
					fmt.Printf("Added: %s\n", track.Title)
					trackCount++
				}

				if trackCount == 0 {
					fmt.Println("No tracks found")
				}
			}

			// Check if URL or search query
			if services.IsURL(input) {
				parseURLAddToQueue(input)
				continue
			}
			// It's a search query
			fmt.Printf("Searching for: %s\n", input)
			ch, err := services.Search(input, cfg.SearchResults)
			if err != nil {
				services.Error("Error searching: %v", err)
				fmt.Printf("Error: %v\n", err)
				continue
			}

			// Collect all results for selection
			results := services.CollectTracks(ch)

			if len(results) == 0 {
				fmt.Println("No results found")
				continue
			}

			// Display results
			fmt.Println("Search results:")
			for i, r := range results {
				fmt.Printf("%d. %s [%s]\n", i+1, r.Title, r.DurationFormatted())
			}

			// Get user selection
			fmt.Print("Select (number): ")
			selection, _ := reader.ReadString('\n')
			selection = strings.TrimSpace(selection)

			// Cancel on empty or invalid
			if selection == "" {
				fmt.Println("Search cancelled")
				continue
			}

			// Parse selection
			idx, err := strconv.Atoi(selection)
			if err != nil || idx < 1 || idx > len(results) {
				fmt.Println("Invalid selection")
				continue
			}
			// Get the selected track
			selected := results[idx-1]
			parseURLAddToQueue(selected.URL)

		case "volume":
			if len(parts) > 1 {
				v, err := strconv.Atoi(parts[1])
				if err != nil || v < 0 || v > 100 {
					fmt.Println("Usage: volume 0-100")
					continue
				}
				player.SetVolume(float64(v) / 100)
				cfg.Volume = player.Volume()
				cfg.Save()
				fmt.Printf("Volume: %d%%\n", v)
			} else {
				fmt.Printf("Volume: %d%%\n", int(player.Volume()*100))
			}

		case "skip":
			player.Stop()
			fmt.Println("Skipped")

		case "pause":
			player.Pause()
			fmt.Println("Paused")

		case "resume":
			player.Resume()
			fmt.Println("Resumed")

		case "stop":
			player.Stop()
			queue.Clear()
			fmt.Println("Stopped and queue cleared")

		case "queue":
			list := queue.List()
			if len(list) == 0 {
				fmt.Println("Queue is empty")
			} else {
				fmt.Println("=== Queue ===")
				for i, t := range list {
					fmt.Printf("%2d. %s [%s]\n", i+1, t.Title, t.DurationFormatted())
				}
			}

		case "quit", "exit":
			player.Stop()
			services.Info("Application exiting")
			fmt.Println("Goodbye!")
			return

		default:
			fmt.Println("Commands: add, volume, skip, pause, resume, stop, queue, quit")
		}
	}
}
