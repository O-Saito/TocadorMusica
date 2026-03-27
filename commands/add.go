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

	global, _ := p.Config().GetProfile(p.Name())
	return searchAndAdd(p, arg, global.RecursiveSearch)
}

func isYouTubeURL(input string) bool {
	return strings.Contains(input, "youtube.com") ||
		strings.Contains(input, "youtu.be")
}

func hasPlaylistParam(url string) bool {
	return strings.Contains(url, "list=") ||
		strings.Contains(url, "/playlist")
}

func stripPlaylistParam(url string) string {
	if strings.Contains(url, "list=") {
		parts := strings.Split(url, "list=")
		base := parts[0]
		if idx := strings.Index(base, "?"); idx != -1 {
			base = base[:idx]
		}
		if len(parts) > 1 {
			rest := parts[1]
			if ampIdx := strings.Index(rest, "&"); ampIdx != -1 {
				rest = rest[ampIdx+1:]
				if len(rest) > 0 {
					if !strings.HasSuffix(base, "?") {
						base += "?"
					}
					base += rest
				}
			}
		}
		if strings.HasSuffix(base, "?") {
			base = strings.TrimSuffix(base, "?")
		}
		return base
	}
	return url
}

func addURL(p domain.PerfilInterface, url string) error {
	if hasPlaylistParam(url) {
		p.Output().Display("This URL contains a playlist")

		options := []string{"Add full playlist", "Add current video only"}
		ch := p.Output().DisplayOptions(options)
		choice := <-ch

		if choice == 0 {
			p.Output().Display("Fetching playlist...")
			tracks, err := p.YtService().ParsePlaylist(p.Context(), url)
			if err != nil {
				return fmt.Errorf("failed to fetch playlist: %w", err)
			}

			p.Output().Display(fmt.Sprintf("Adding %d tracks:", len(tracks)))
			for _, track := range tracks {
				err := p.Queue().Enqueue(track)
				if err != nil {
					return fmt.Errorf("failed to add to queue: %w", err)
				}
				p.Output().Display("  + " + track.Title())
			}
			p.Output().ShowQueue(p.GetQueueItems())

			if p.Config().GetAutoPlay(p.Name()) {
				p.ExecuteCommand("play", nil)
			}
			return nil
		}

		url = stripPlaylistParam(url)
	}

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

	if p.Config().GetAutoPlay(p.Name()) {
		p.ExecuteCommand("play", nil)
	}

	return nil
}

func searchAndAdd(p domain.PerfilInterface, query string, recursive bool) error {
	global, _ := p.Config().GetProfile(p.Name())

	fileResults, err := p.FileService().Search(global.MusicFolders, query, recursive)
	if err != nil {
		return fmt.Errorf("file search failed: %w", err)
	}

	if len(fileResults) > 0 {
		return selectAndAddFile(p, fileResults, query)
	}

	return searchAndAddYouTube(p, query)
}

func selectAndAddFile(p domain.PerfilInterface, tracks []domain.Track, query string) error {
	_, profile := p.Config().GetProfile(p.Name())
	pageSize := profile.SearchResults

	if len(tracks) == 1 {
		track := tracks[0]
		err := p.Queue().Enqueue(track)
		if err != nil {
			return fmt.Errorf("failed to add to queue: %w", err)
		}

		p.Output().Display("Added: " + track.Title())
		p.Output().ShowQueue(p.GetQueueItems())

		if p.Config().GetAutoPlay(p.Name()) {
			p.ExecuteCommand("play", nil)
		}

		return nil
	}

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

		p.Output().Display("Select a file:")
		ch := p.Output().DisplayOptionsPage(titles, currentPage, totalPages, true)
		idx := <-ch

		if idx == -3 {
			return searchAndAddYouTube(p, query)
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
		err := p.Queue().Enqueue(track)
		if err != nil {
			return fmt.Errorf("failed to add to queue: %w", err)
		}

		p.Output().Display("Added: " + track.Title())
		p.Output().ShowQueue(p.GetQueueItems())

		if p.Config().GetAutoPlay(p.Name()) {
			p.ExecuteCommand("play", nil)
		}

		return nil
	}
}

func searchAndAddYouTube(p domain.PerfilInterface, query string) error {
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

		err = p.Queue().Enqueue(track)
		if err != nil {
			return fmt.Errorf("failed to add to queue: %w", err)
		}

		p.Output().Display("Added: " + track.Title())
		p.Output().ShowQueue(p.GetQueueItems())

		if p.Config().GetAutoPlay(p.Name()) {
			p.ExecuteCommand("play", nil)
		}

		return nil
	}
}

var _ Command = (*AddCommand)(nil)
