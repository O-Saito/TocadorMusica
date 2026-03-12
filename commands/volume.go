package commands

import (
	"fmt"
	"strconv"
)

type VolumeCommand struct{}

func (c *VolumeCommand) Name() string        { return "volume" }
func (c *VolumeCommand) Description() string { return "Get/set volume (0-100)" }

func (c *VolumeCommand) Execute(ctx CommandContext, args []string) error {
	global, profile := ctx.Config.GetProfile(ctx.ProfileName)
	_ = global

	if len(args) == 0 {
		ctx.Output.Display(fmt.Sprintf("Volume: %.0f%%", profile.Volume*100))
		return nil
	}

	vol, err := strconv.Atoi(args[0])
	if err != nil || vol < 0 || vol > 100 {
		return fmt.Errorf("usage: volume [0-100]")
	}

	ctx.Player.SetVolume(float64(vol) / 100)
	ctx.Config.SetVolume(ctx.ProfileName, float64(vol)/100)
	ctx.Output.Display(fmt.Sprintf("Volume: %d%%", vol))
	return nil
}

var _ Command = (*VolumeCommand)(nil)
