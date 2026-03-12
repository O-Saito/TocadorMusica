package commands

import "fmt"

type QueueCommand struct{}

func (c *QueueCommand) Name() string        { return "queue" }
func (c *QueueCommand) Description() string { return "Show queue size" }

func (c *QueueCommand) Execute(ctx CommandContext, args []string) error {
	size := ctx.Queue.Size()
	ctx.Output.Display(fmt.Sprintf("Queue size: %d", size))
	return nil
}

var _ Command = (*QueueCommand)(nil)
