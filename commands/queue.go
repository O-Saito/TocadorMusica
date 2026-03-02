package commands

import "fmt"

type QueueCommand struct{}

func (c *QueueCommand) Name() string        { return "queue" }
func (c *QueueCommand) Description() string { return "Show queued tracks" }

func (c *QueueCommand) Execute(ctx *CommandContext, args []string) error {
	list := ctx.Queue.List()
	if len(list) == 0 {
		fmt.Println("Queue is empty")
		return nil
	}

	fmt.Println("=== Queue ===")
	for i, t := range list {
		fmt.Printf("%2d. %s [%s]\n", i+1, t.Title, t.DurationFormatted())
	}
	return nil
}
