package commands

import (
	"tocadormusica/domain"
)

type ListCommand struct{}

func (c *ListCommand) Name() string        { return "list" }
func (c *ListCommand) Description() string { return "List all music files in configured folders" }

func (c *ListCommand) Execute(p domain.PerfilInterface, args []string) error {
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

		p.Output().Display("Select a file to add:")
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

		if idx < 0 || idx >= len(pageTracks) {
			p.Output().Display("Invalid selection")
			continue
		}

		track := pageTracks[idx]
		err = p.Queue().Enqueue(track)
		if err != nil {
			return err
		}

		p.Output().Display("Added: " + track.Title())
		p.Output().ShowQueue(p.GetQueueItems())

		if p.Config().GetAutoPlay(p.Name()) {
			p.ExecuteCommand("play", nil)
		}

		return nil
	}
}

var _ Command = (*ListCommand)(nil)
