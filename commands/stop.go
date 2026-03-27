package commands

import "tocadormusica/domain"

type StopCommand struct{}

func (c *StopCommand) Name() string        { return "stop" }
func (c *StopCommand) Description() string { return "Stop current track" }

func (c *StopCommand) Execute(p domain.PerfilInterface, args []string) error {
	p.Player().Stop()
	p.Output().Display("Stopped")
	p.Output().ShowNowPlaying("")
	return nil
}

var _ Command = (*StopCommand)(nil)
