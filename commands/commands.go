package commands

import (
	"sort"

	"tocadormusica/domain"
)

type Command interface {
	Name() string
	Description() string
	Execute(perfil domain.PerfilInterface, args []string) error
}

var registry = make(map[string]Command)

func Register(cmd Command) {
	registry[cmd.Name()] = cmd
}

func Get(name string) Command {
	return registry[name]
}

func List() []Command {
	cmds := make([]Command, 0, len(registry))
	for _, cmd := range registry {
		cmds = append(cmds, cmd)
	}
	sort.Slice(cmds, func(i, j int) bool {
		return cmds[i].Name() < cmds[j].Name()
	})
	return cmds
}
