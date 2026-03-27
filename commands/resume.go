package commands

import (
	"tocadormusica/domain"
)

type ResumeCommand struct{}

func (c *ResumeCommand) Name() string        { return "resume" }
func (c *ResumeCommand) Description() string { return "Resume paused track" }

func (c *ResumeCommand) Execute(p domain.PerfilInterface, args []string) error {
	if p.GetBackgroundTrack() != "" && p.IsBackgroundPaused() {
		p.ResumeBackground()
	} else {
		p.Player().Resume()
	}
	p.Output().Display("Resumed")
	return nil
}

var _ Command = (*ResumeCommand)(nil)
