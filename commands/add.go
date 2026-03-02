package commands

import (
	"fmt"
	"strconv"
	"strings"

	"tocadormusica/models"
	"tocadormusica/services"
)

type AddCommand struct{}

func (c *AddCommand) Name() string        { return "add" }
func (c *AddCommand) Description() string { return "Add YouTube URL or search for tracks" }

func (c *AddCommand) Execute(ctx *CommandContext, args []string) error {
	input := strings.Join(args, " ")

	if input == "" {
		fmt.Print("Paste YouTube URL or search: ")
		input, _ = ctx.Reader.ReadString('\n')
		input = strings.TrimSpace(input)
	}

	if input == "" {
		return fmt.Errorf("no input provided")
	}

	parseURLAddToQueue := func(url string) error {
		fmt.Println("Fetching tracks...")
		ch, err := services.ParseURL(url)
		if err != nil {
			return fmt.Errorf("error parsing URL: %w", err)
		}

		trackCount := 0
		for track := range ch {
			ctx.Queue.Add([]models.Track{track})
			fmt.Printf("Added: %s\n", track.Title)
			trackCount++
		}

		if trackCount == 0 {
			fmt.Println("No tracks found")
		}
		return nil
	}

	if services.IsURL(input) {
		return parseURLAddToQueue(input)
	}

	fmt.Printf("Searching for: %s\n", input)
	ch, err := services.Search(input, ctx.Config.SearchResults)
	if err != nil {
		return fmt.Errorf("error searching: %w", err)
	}

	results := services.CollectTracks(ch)

	if len(results) == 0 {
		fmt.Println("No results found")
		return nil
	}

	fmt.Println("Search results:")
	for i, r := range results {
		fmt.Printf("%d. %s [%s]\n", i+1, r.Title, r.DurationFormatted())
	}

	fmt.Print("Select (number): ")
	selection, _ := ctx.Reader.ReadString('\n')
	selection = strings.TrimSpace(selection)

	if selection == "" {
		fmt.Println("Search cancelled")
		return nil
	}

	idx, err := strconv.Atoi(selection)
	if err != nil || idx < 1 || idx > len(results) {
		return fmt.Errorf("invalid selection")
	}

	selected := results[idx-1]
	return parseURLAddToQueue(selected.URL)
}
