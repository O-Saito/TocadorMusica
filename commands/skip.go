package commands

import "fmt"

type SkipCommand struct{}

func (c *SkipCommand) Name() string        { return "skip" }
func (c *SkipCommand) Description() string { return "Skip current track" }

func (c *SkipCommand) Execute(ctx *CommandContext, args []string) error {
	ctx.Player.Stop()
	fmt.Println("Skipped")
	return nil
}
