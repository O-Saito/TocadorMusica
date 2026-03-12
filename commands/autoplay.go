package commands

import (
	"tocadormusica/domain"
)

type AutoplayCommand struct{}

func (c *AutoplayCommand) Name() string        { return "autoplay" }
func (c *AutoplayCommand) Description() string { return "Toggle autoplay on/off" }

func (c *AutoplayCommand) Execute(p domain.PerfilInterface, args []string) error {
	current := p.Config().GetAutoPlay(p.Name())
	newValue := !current
	p.Config().SetAutoPlay(p.Name(), newValue)

	if newValue {
		p.Output().Display("Autoplay: enabled")
	} else {
		p.Output().Display("Autoplay: disabled")
	}
	return nil
}

var _ Command = (*AutoplayCommand)(nil)
