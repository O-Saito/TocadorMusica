package commands

import (
	"fmt"
	"strings"

	"tocadormusica/domain"
)

type AddCommand struct{}

func (c *AddCommand) Name() string        { return "add" }
func (c *AddCommand) Description() string { return "Add a track to queue (url or search query)" }

func (c *AddCommand) Execute(p domain.PerfilInterface, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: add <url or search query>")
	}

	arg := args[0]

	if isYouTubeURL(arg) {
		return addURL(p, arg)
	}

	return searchAndAdd(p, arg)
}

func isYouTubeURL(input string) bool {
	return strings.Contains(input, "youtube.com") ||
		strings.Contains(input, "youtu.be")
}

func addURL(p domain.PerfilInterface, url string) error {
	p.Output().Display("Fetching track...")

	track, err := p.YtService().ParseURL(p.Context(), url)
	if err != nil {
		return fmt.Errorf("failed to fetch track: %w", err)
	}

	err = p.Queue().Enqueue(track)
	if err != nil {
		return fmt.Errorf("failed to add to queue: %w", err)
	}

	p.Output().Display("Added: " + track.Title())
	p.Output().ShowQueue(p.GetQueueItems())
	return nil
}

func searchAndAdd(p domain.PerfilInterface, query string) error {
	p.Output().Display("Searching...")

	_, profile := p.Config().GetProfile(p.Name())
	results, err := p.YtService().Search(p.Context(), query, profile.SearchResults)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(results) == 0 {
		p.Output().Display("No results found")
		return nil
	}

	titles := make([]string, len(results))
	for i, r := range results {
		titles[i] = fmt.Sprintf("%s - %s", r.Title, r.Duration)
	}

	p.Output().Display("Select a track:")
	ch := p.Output().DisplayOptions(titles)
	idx := <-ch

	if idx < 0 || idx >= len(results) {
		p.Output().Display("Invalid selection")
		return nil
	}

	result := results[idx]
	track := domain.NewTrackFromYouTube(result.URL, result.Title, "", "")

	err = p.Queue().Enqueue(track)
	if err != nil {
		return fmt.Errorf("failed to add to queue: %w", err)
	}

	p.Output().Display("Added: " + track.Title())
	p.Output().ShowQueue(p.GetQueueItems())
	return nil
}

var _ Command = (*AddCommand)(nil)
