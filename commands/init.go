package commands

func init() {
	Register(&AddCommand{})
	Register(&VolumeCommand{})
	Register(&SkipCommand{})
	Register(&PauseCommand{})
	Register(&ResumeCommand{})
	Register(&StopCommand{})
	Register(&QueueCommand{})
	Register(&QuitCommand{})
}
