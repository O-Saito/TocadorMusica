package commands

import (
	"bufio"
	"sort"

	"tocadormusica/config"
	"tocadormusica/models"
	"tocadormusica/services"
)

type CommandContext struct {
	Queue  *models.Queue
	Player *services.AudioPlayer
	Config *config.Config
	Reader *bufio.Reader
}

type Command interface {
	Name() string
	Description() string
	Execute(ctx *CommandContext, args []string) error
}

var registry = make(map[string]Command)

func Register(cmd Command) {
	registry[cmd.Name()] = cmd
}

func Get(name string) Command {
	return registry[name]
}

func List() []Command {
	commands := make([]Command, 0, len(registry))
	for _, cmd := range registry {
		commands = append(commands, cmd)
	}
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name() < commands[j].Name()
	})
	return commands
}
