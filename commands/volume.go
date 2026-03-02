package commands

import (
	"fmt"
	"strconv"
)

type VolumeCommand struct{}

func (c *VolumeCommand) Name() string        { return "volume" }
func (c *VolumeCommand) Description() string { return "Get/set volume (0-100)" }

func (c *VolumeCommand) Execute(ctx *CommandContext, args []string) error {
	if len(args) == 0 {
		fmt.Printf("Volume: %d%%\n", int(ctx.Player.Volume()*100))
		return nil
	}

	v, err := strconv.Atoi(args[0])
	if err != nil || v < 0 || v > 100 {
		return fmt.Errorf("usage: volume 0-100")
	}

	ctx.Player.SetVolume(float64(v) / 100)
	ctx.Config.Volume = ctx.Player.Volume()
	ctx.Config.Save()
	fmt.Printf("Volume: %d%%\n", v)
	return nil
}
