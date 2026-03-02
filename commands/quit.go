package commands

import "fmt"

type QuitCommand struct{}

func (c *QuitCommand) Name() string        { return "quit" }
func (c *QuitCommand) Description() string { return "Exit the application" }

func (c *QuitCommand) Execute(ctx *CommandContext, args []string) error {
	ctx.Player.Stop()
	fmt.Println("Goodbye!")
	return &ExitError{}
}
