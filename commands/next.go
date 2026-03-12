package commands

import (
	"fmt"

	"tocadormusica/domain"
)

type NextCommand struct{}

func (c *NextCommand) Name() string        { return "next" }
func (c *NextCommand) Description() string { return "Skip to next track in queue" }

func (c *NextCommand) Execute(p domain.PerfilInterface, args []string) error {
	if p.Queue().IsEmpty() {
		p.Output().Display("Queue is empty")
		return nil
	}

	_, err := p.Queue().Dequeue()
	if err != nil {
		return err
	}

	p.Output().ShowQueue(p.GetQueueItems())

	if !p.Queue().IsEmpty() {
		track, err := p.Queue().Peek()
		if err != nil {
			p.Output().Display("Queue is empty")
			p.Player().Stop()
			p.Output().ShowNowPlaying("")
			return nil
		}

		p.Logger().Debug("playing track", "title", track.Title(), "url", track.URL())

		audioURL := track.AudioURL()
		if audioURL == "" {
			p.Output().Display("Fetching audio URL...")
			audioURL, err = p.YtService().GetAudioURL(p.Context(), track.URL())
			if err != nil {
				return fmt.Errorf("failed to get audio URL: %w", err)
			}
			track.SetAudioURL(audioURL)
		}

		p.Output().Display("Streaming: " + track.Title())

		global, _ := p.Config().GetProfile(p.Name())
		err = p.Player().PlayURL(audioURL, global.SampleRate)
		if err != nil {
			return err
		}

		p.Output().Display("Playing: " + track.Title())
		p.Output().ShowNowPlaying(track.Title())
	} else {
		p.Output().Display("Queue is empty")
		p.Player().Stop()
		p.Output().ShowNowPlaying("")
	}

	return nil
}

var _ Command = (*NextCommand)(nil)
