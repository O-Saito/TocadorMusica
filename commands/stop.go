package commands

type StopCommand struct{}

func (c *StopCommand) Name() string        { return "stop" }
func (c *StopCommand) Description() string { return "Stop current track" }

func (c *StopCommand) Execute(ctx CommandContext, args []string) error {
	ctx.Player.Stop()
	ctx.Output.Display("Stopped")
	return nil
}

var _ Command = (*StopCommand)(nil)
