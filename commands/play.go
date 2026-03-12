package commands

import "fmt"

type PlayCommand struct{}

func (c *PlayCommand) Name() string        { return "play" }
func (c *PlayCommand) Description() string { return "Play the first track in queue" }

func (c *PlayCommand) Execute(ctx CommandContext, args []string) error {
	track, err := ctx.Queue.Peek()
	if err != nil {
		ctx.Output.Display("Queue is empty")
		return nil
	}

	ctx.Logger.Debug("playing track", "title", track.Title(), "url", track.URL(), "audioURL", track.AudioURL())

	if track.AudioURL() == "" {
		return fmt.Errorf("no audio URL available for this track")
	}

	ctx.Output.Display("Streaming: " + track.Title())

	global, _ := ctx.Config.GetProfile(ctx.ProfileName)
	err = ctx.Player.PlayURL(track.AudioURL(), global.SampleRate)
	if err != nil {
		return fmt.Errorf("failed to play: %w", err)
	}

	ctx.Output.Display("Playing: " + track.Title())
	return nil
}

var _ Command = (*PlayCommand)(nil)
