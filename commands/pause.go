package commands

import "fmt"

type PauseCommand struct{}

func (c *PauseCommand) Name() string        { return "pause" }
func (c *PauseCommand) Description() string { return "Pause playback" }

func (c *PauseCommand) Execute(ctx *CommandContext, args []string) error {
	ctx.Player.Pause()
	fmt.Println("Paused")
	return nil
}
