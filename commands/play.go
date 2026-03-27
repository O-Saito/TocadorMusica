package commands

import (
	"fmt"

	"tocadormusica/domain"
)

type PlayCommand struct{}

func (c *PlayCommand) Name() string        { return "play" }
func (c *PlayCommand) Description() string { return "Play the first track in queue" }

func (c *PlayCommand) Execute(p domain.PerfilInterface, args []string) error {
	if p.IsBackgroundPlaying() {
		p.StopBackground()
	}

	track, err := p.Queue().Peek()
	if err != nil {
		p.Output().Display("Queue is empty")
		p.Player().Stop()
		p.Output().ShowNowPlaying("")
		if p.GetBackgroundTrack().Title() != "" {
			p.StartBackground()
		}
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

	autoPlay := p.Config().GetAutoPlay(p.Name())

	if autoPlay {
		p.Player().SetOnFinishedCallback(func() {
			if p.Config().GetAutoPlay(p.Name()) {
				_, err := p.Queue().Dequeue()
				if err != nil {
					if p.GetBackgroundTrack().Title() != "" {
						p.StartBackground()
					}
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
	err = p.Player().PlayURL(audioURL, global.SampleRate)
	if err != nil {
		p.Player().SetOnFinishedCallback(nil)
		return fmt.Errorf("failed to play: %w", err)
	}

	p.Output().Display("Playing: " + track.Title())
	p.Output().ShowNowPlaying(track.Title())
	return nil
}

var _ Command = (*PlayCommand)(nil)
