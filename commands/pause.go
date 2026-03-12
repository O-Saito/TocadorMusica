package commands

type PauseCommand struct{}

func (c *PauseCommand) Name() string        { return "pause" }
func (c *PauseCommand) Description() string { return "Pause current track" }

func (c *PauseCommand) Execute(ctx CommandContext, args []string) error {
	ctx.Player.Pause()
	ctx.Output.Display("Paused")
	return nil
}

var _ Command = (*PauseCommand)(nil)

type ResumeCommand struct{}

func (c *ResumeCommand) Name() string        { return "resume" }
func (c *ResumeCommand) Description() string { return "Resume paused track" }

func (c *ResumeCommand) Execute(ctx CommandContext, args []string) error {
	ctx.Player.Resume()
	ctx.Output.Display("Resumed")
	return nil
}

var _ Command = (*ResumeCommand)(nil)
