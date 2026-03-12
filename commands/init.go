package commands

func init() {
	Register(&AddCommand{})
	Register(&PauseCommand{})
	Register(&PlayCommand{})
	Register(&QueueCommand{})
	Register(&StopCommand{})
	Register(&NextCommand{})
	Register(&VolumeCommand{})
	Register(&ResumeCommand{})
}
