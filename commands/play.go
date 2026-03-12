package commands

import (
	"fmt"

	"tocadormusica/domain"
)

type PlayCommand struct{}

func (c *PlayCommand) Name() string        { return "play" }
func (c *PlayCommand) Description() string { return "Play the first track in queue" }

func (c *PlayCommand) Execute(p domain.PerfilInterface, args []string) error {
	track, err := p.Queue().Peek()
	if err != nil {
		p.Output().Display("Queue is empty")
		return nil
	}

	p.Logger().Debug("playing track", "title", track.Title(), "url", track.URL(), "audioURL", track.AudioURL())

	if track.AudioURL() == "" {
		return fmt.Errorf("no audio URL available for this track")
	}

	autoPlay := p.Config().GetAutoPlay(p.Name())

	if autoPlay {
		p.Player().SetOnFinishedCallback(func() {
			if p.Config().GetAutoPlay(p.Name()) {
				_, err := p.Queue().Dequeue()
				if err != nil {
					return
				}
				p.Output().ShowQueue(p.GetQueueItems())
				p.ExecuteCommand("play", nil)
			}
		})
	} else {
		p.Player().SetOnFinishedCallback(nil)
	}

	p.Output().Display("Streaming: " + track.Title())

	global, _ := p.Config().GetProfile(p.Name())
	err = p.Player().PlayURL(track.AudioURL(), global.SampleRate)
	if err != nil {
		p.Player().SetOnFinishedCallback(nil)
		return fmt.Errorf("failed to play: %w", err)
	}

	p.Output().Display("Playing: " + track.Title())
	p.Output().ShowNowPlaying(track.Title())
	return nil
}

var _ Command = (*PlayCommand)(nil)
