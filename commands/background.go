package commands

import (
	"fmt"

	"tocadormusica/domain"
)

type BackgroundCommand struct{}

func (c *BackgroundCommand) Name() string { return "background" }
func (c *BackgroundCommand) Description() string {
	return "Set background music (no args: select file, stop: stop, status: show status)"
}

func (c *BackgroundCommand) Execute(p domain.PerfilInterface, args []string) error {
	if len(args) > 0 {
		switch args[0] {
		case "stop":
			p.ClearBackground()
			return nil
		case "status":
			if p.GetBackgroundTrack() == "" {
				p.Output().Display("No background music set")
			} else if p.IsBackgroundPlaying() {
				p.Output().Display(fmt.Sprintf("Background: %s (playing)", p.GetBackgroundTrack()))
			} else if p.IsBackgroundPaused() {
				p.Output().Display(fmt.Sprintf("Background: %s (paused at %ds)", p.GetBackgroundTrack(), p.GetBackgroundPosition()))
			} else {
				p.Output().Display(fmt.Sprintf("Background: %s (stopped)", p.GetBackgroundTrack()))
			}
			return nil
		}
	}

	if p.GetBackgroundTrack() != "" {
		p.Output().Display(fmt.Sprintf("Current background: %s", p.GetBackgroundTrack()))
	}

	global, _ := p.Config().GetProfile(p.Name())

	if len(global.MusicFolders) == 0 {
		p.Output().Display("No music folders configured")
		return nil
	}

	tracks, err := p.FileService().ListAll(global.MusicFolders, global.RecursiveSearch)
	if err != nil {
		return err
	}

	if len(tracks) == 0 {
		p.Output().Display("No music files found in configured folders")
		return nil
	}

	return selectAndSetBackground(p, tracks)
}

func selectAndSetBackground(p domain.PerfilInterface, tracks []domain.Track) error {
	options := []string{"Search on YouTube"}
	for _, t := range tracks {
		options = append(options, t.Title())
	}

	p.Output().Display("Select background music:")
	ch := p.Output().DisplayOptions(options)
	idx := <-ch

	if idx < 0 || idx > len(tracks) {
		p.Output().Display("Invalid selection")
		return nil
	}

	if idx == 0 {
		p.Output().Display("Enter search query:")
		ch := p.Output().RequestInput("Search:")
		query := <-ch
		if query == "" {
			p.Output().Display("Empty search query")
			return nil
		}
		return searchAndSetBackgroundYouTube(p, query)
	}

	track := tracks[idx-1]
	err := p.SetBackground(track.AudioURL())
	if err != nil {
		return fmt.Errorf("failed to set background: %w", err)
	}

	return nil
}

func searchAndSetBackgroundYouTube(p domain.PerfilInterface, query string) error {
	p.Output().Display("Searching YouTube...")

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

	err = p.SetBackground(track.AudioURL())
	if err != nil {
		return fmt.Errorf("failed to set background: %w", err)
	}

	return nil
}

var _ Command = (*BackgroundCommand)(nil)
