package commands

import (
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

	if !p.Queue().IsEmpty() {
		track, err := p.Queue().Peek()
		if err != nil {
			p.Output().Display("Queue is empty")
			p.Player().Stop()
			return nil
		}

		p.Logger().Debug("playing track", "title", track.Title(), "url", track.URL(), "audioURL", track.AudioURL())

		if track.AudioURL() == "" {
			p.Output().Display("Error: No audio URL available for this track")
			return nil
		}

		p.Output().Display("Streaming: " + track.Title())

		global, _ := p.Config().GetProfile(p.Name())
		err = p.Player().PlayURL(track.AudioURL(), global.SampleRate)
		if err != nil {
			return err
		}

		p.Output().Display("Playing: " + track.Title())
	} else {
		p.Output().Display("Queue is empty")
		p.Player().Stop()
	}

	return nil
}

var _ Command = (*NextCommand)(nil)
