package commands

import "fmt"

type ResumeCommand struct{}

func (c *ResumeCommand) Name() string        { return "resume" }
func (c *ResumeCommand) Description() string { return "Resume playback" }

func (c *ResumeCommand) Execute(ctx *CommandContext, args []string) error {
	ctx.Player.Resume()
	fmt.Println("Resumed")
	return nil
}
