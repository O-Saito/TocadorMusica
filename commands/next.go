package commands

type NextCommand struct{}

func (c *NextCommand) Name() string        { return "next" }
func (c *NextCommand) Description() string { return "Skip to next track in queue" }

func (c *NextCommand) Execute(ctx CommandContext, args []string) error {
	if ctx.Queue.IsEmpty() {
		ctx.Output.Display("Queue is empty")
		return nil
	}

	_, err := ctx.Queue.Dequeue()
	if err != nil {
		return err
	}

	if !ctx.Queue.IsEmpty() {
		track, err := ctx.Queue.Peek()
		if err != nil {
			ctx.Output.Display("Queue is empty")
			ctx.Player.Stop()
			return nil
		}

		ctx.Logger.Debug("playing track", "title", track.Title(), "url", track.URL(), "audioURL", track.AudioURL())

		if track.AudioURL() == "" {
			ctx.Output.Display("Error: No audio URL available for this track")
			return nil
		}

		ctx.Output.Display("Streaming: " + track.Title())

		global, _ := ctx.Config.GetProfile(ctx.ProfileName)
		err = ctx.Player.PlayURL(track.AudioURL(), global.SampleRate)
		if err != nil {
			return err
		}

		ctx.Output.Display("Playing: " + track.Title())
	} else {
		ctx.Output.Display("Queue is empty")
		ctx.Player.Stop()
	}

	return nil
}

var _ Command = (*NextCommand)(nil)
