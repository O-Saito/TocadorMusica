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

	titles := make([]string, len(tracks))
	for i, t := range tracks {
		titles[i] = t.Title()
	}

	p.Output().Display("Select a file to add:")
	ch := p.Output().DisplayOptions(titles)
	idx := <-ch

	if idx < 0 || idx >= len(tracks) {
		p.Output().Display("Invalid selection")
		return nil
	}

	track := tracks[idx]
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

var _ Command = (*ListCommand)(nil)
