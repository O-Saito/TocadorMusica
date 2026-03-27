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
			if p.GetBackgroundTrack().Title() == "" {
				p.Output().Display("No background music set")
			} else if p.IsBackgroundPlaying() {
				p.Output().Display(fmt.Sprintf("Background: %s (playing)", p.GetBackgroundTrack().Title()))
			} else if p.IsBackgroundPaused() {
				p.Output().Display(fmt.Sprintf("Background: %s (paused at %ds)", p.GetBackgroundTrack().Title(), p.GetBackgroundPosition()))
			} else {
				p.Output().Display(fmt.Sprintf("Background: %s (stopped)", p.GetBackgroundTrack().Title()))
			}
			return nil
		}
	}

	if p.GetBackgroundTrack().Title() != "" {
		p.Output().Display(fmt.Sprintf("Current background: %s", p.GetBackgroundTrack().Title()))
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
	_, profile := p.Config().GetProfile(p.Name())
	pageSize := profile.SearchResults
	currentPage := 0
	totalPages := (len(tracks) + pageSize - 1) / pageSize

	for {
		start := currentPage * pageSize
		end := start + pageSize
		if end > len(tracks) {
			end = len(tracks)
		}

		pageTracks := tracks[start:end]
		titles := make([]string, len(pageTracks))
		for i, t := range pageTracks {
			titles[i] = t.Title()
		}

		p.Output().Display("Select background music:")
		ch := p.Output().DisplayOptionsPage(titles, currentPage, totalPages, false)
		idx := <-ch

		if idx == -3 {
			p.Output().Display("Enter search query:")
			ch := p.Output().RequestInput("Search:")
			query := <-ch
			if query == "" {
				p.Output().Display("Empty search query")
				continue
			}
			return searchAndSetBackgroundYouTube(p, query)
		}

		if idx == -4 {
			p.Output().Display("Cancelled")
			return nil
		}

		if idx == -1 {
			if currentPage < totalPages-1 {
				currentPage++
			}
			continue
		}

		if idx == -2 {
			if currentPage > 0 {
				currentPage--
			}
			continue
		}

		if idx < 0 || idx >= len(pageTracks) {
			p.Output().Display("Invalid selection")
			continue
		}

		track := pageTracks[idx]
		err := p.SetBackground(track.AudioURL())
		if err != nil {
			return fmt.Errorf("failed to set background: %w", err)
		}

		return nil
	}
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

	pageSize := profile.SearchResults
	currentPage := 0
	totalPages := (len(results) + pageSize - 1) / pageSize

	for {
		start := currentPage * pageSize
		end := start + pageSize
		if end > len(results) {
			end = len(results)
		}

		pageResults := results[start:end]
		titles := make([]string, len(pageResults))
		for i, r := range pageResults {
			titles[i] = fmt.Sprintf("%s - %s", r.Title, r.Duration)
		}

		p.Output().Display("Select a track:")
		ch := p.Output().DisplayOptionsPage(titles, currentPage, totalPages, false)
		idx := <-ch

		if idx == -4 {
			p.Output().Display("Cancelled")
			return nil
		}

		if idx == -1 {
			if currentPage < totalPages-1 {
				currentPage++
			}
			continue
		}

		if idx == -2 {
			if currentPage > 0 {
				currentPage--
			}
			continue
		}

		if idx < 0 || idx >= len(pageResults) {
			p.Output().Display("Invalid selection")
			continue
		}

		result := pageResults[idx]
		track := domain.NewTrackFromYouTube(result.URL, result.Title, "", "")

		err = p.SetBackground(track.AudioURL())
		if err != nil {
			return fmt.Errorf("failed to set background: %w", err)
		}

		return nil
	}
}

var _ Command = (*BackgroundCommand)(nil)
