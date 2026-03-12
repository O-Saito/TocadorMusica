package commands

import (
	"fmt"
	"strconv"

	"tocadormusica/domain"
)

type VolumeCommand struct{}

func (c *VolumeCommand) Name() string        { return "volume" }
func (c *VolumeCommand) Description() string { return "Get/set volume (0-100)" }

func (c *VolumeCommand) Execute(p domain.PerfilInterface, args []string) error {
	_, profile := p.Config().GetProfile(p.Name())

	if len(args) == 0 {
		p.Output().Display(fmt.Sprintf("Volume: %.0f%%", profile.Volume*100))
		return nil
	}

	vol, err := strconv.Atoi(args[0])
	if err != nil || vol < 0 || vol > 100 {
		return fmt.Errorf("usage: volume [0-100]")
	}

	p.Player().SetVolume(float64(vol) / 100)
	p.Config().SetVolume(p.Name(), float64(vol)/100)
	p.Output().Display(fmt.Sprintf("Volume: %d%%", vol))
	return nil
}

var _ Command = (*VolumeCommand)(nil)
