package commands

import (
	"context"
	"fmt"
	"strings"
)

type AddCommand struct{}

func (c *AddCommand) Name() string        { return "add" }
func (c *AddCommand) Description() string { return "Add a track to queue (url or search query)" }

func (c *AddCommand) Execute(ctx CommandContext, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: add <url or search query>")
	}

	arg := args[0]

	if isYouTubeURL(arg) {
		return addURL(ctx, arg)
	}

	return searchAndAdd(ctx, arg)
}

func isYouTubeURL(input string) bool {
	return strings.Contains(input, "youtube.com") ||
		strings.Contains(input, "youtu.be")
}

func addURL(ctx CommandContext, url string) error {
	ctx.Output.Display("Fetching track...")

	track, err := ctx.YtService.ParseURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("failed to fetch track: %w", err)
	}

	err = ctx.Queue.Enqueue(track)
	if err != nil {
		return fmt.Errorf("failed to add to queue: %w", err)
	}

	ctx.Output.Display("Added: " + track.Title())
	return nil
}

func searchAndAdd(ctx CommandContext, query string) error {
	ctx.Output.Display("Searching...")

	_, profile := ctx.Config.GetProfile(ctx.ProfileName)
	results, err := ctx.YtService.Search(context.Background(), query, profile.SearchResults)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(results) == 0 {
		ctx.Output.Display("No results found")
		return nil
	}

	titles := make([]string, len(results))
	for i, r := range results {
		titles[i] = fmt.Sprintf("%s - %s", r.Title(), r.Duration())
	}

	ctx.Output.Display("Select a track:")
	ch := ctx.Output.DisplayOptions(titles)
	idx := <-ch

	if idx < 0 || idx >= len(results) {
		ctx.Output.Display("Invalid selection")
		return nil
	}

	return addURL(ctx, results[idx].URL())
}

var _ Command = (*AddCommand)(nil)
