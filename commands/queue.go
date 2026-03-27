package commands

import (
	"fmt"

	"tocadormusica/domain"
)

type QueueCommand struct{}

func (c *QueueCommand) Name() string        { return "queue" }
func (c *QueueCommand) Description() string { return "Show queue size" }

func (c *QueueCommand) Execute(p domain.PerfilInterface, args []string) error {
	size := p.Queue().Size()
	p.Output().Display(fmt.Sprintf("Queue size: %d", size))
	p.Output().ShowQueue(p.GetQueueItems())
	return nil
}

var _ Command = (*QueueCommand)(nil)
