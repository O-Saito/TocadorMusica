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

type ResumeCommand struct{}

func (c *ResumeCommand) Name() string        { return "resume" }
func (c *ResumeCommand) Description() string { return "Resume paused track" }

func (c *ResumeCommand) Execute(p domain.PerfilInterface, args []string) error {
	p.Player().Resume()
	p.Output().Display("Resumed")
	return nil
}

var _ Command = (*ResumeCommand)(nil)
