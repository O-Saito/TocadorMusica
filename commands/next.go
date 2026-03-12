package commands

import (
	"tocadormusica/domain"
)

type NextCommand struct{}

func (c *NextCommand) Name() string        { return "next" }
func (c *NextCommand) Description() string { return "Skip to next track in queue" }

func (c *NextCommand) Execute(p domain.PerfilInterface, args []string) error {
	if p.Queue().IsEmpty() {
		p.Output().Display("Queue is empty")
		return nil
	}

	_, err := p.Queue().Dequeue()
	if err != nil {
		return err
	}

	p.Output().ShowQueue(p.GetQueueItems())

	p.ExecuteCommand("play", nil)
	return nil
}

var _ Command = (*NextCommand)(nil)
