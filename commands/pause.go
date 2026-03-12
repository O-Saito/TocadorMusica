package commands

import (
	"tocadormusica/domain"
)

type PauseCommand struct{}

func (c *PauseCommand) Name() string        { return "pause" }
func (c *PauseCommand) Description() string { return "Pause current track" }

func (c *PauseCommand) Execute(p domain.PerfilInterface, args []string) error {
	p.Player().Pause()
	p.Output().Display("Paused")
	return nil
}

var _ Command = (*PauseCommand)(nil)
