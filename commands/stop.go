package commands

import "fmt"

type StopCommand struct{}

func (c *StopCommand) Name() string        { return "stop" }
func (c *StopCommand) Description() string { return "Stop playback and clear queue" }

func (c *StopCommand) Execute(ctx *CommandContext, args []string) error {
	ctx.Player.Stop()
	ctx.Queue.Clear()
	fmt.Println("Stopped and queue cleared")
	return nil
}
