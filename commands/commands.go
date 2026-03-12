package commands

import (
	"sort"

	"tocadormusica/config"
	"tocadormusica/logger"
	"tocadormusica/ports/audio"
	"tocadormusica/ports/ui"
)

type Track interface {
	Title() string
	URL() string
	AudioURL() string
}

type SearchResult interface {
	Title() string
	URL() string
	Duration() string
}

type Queue interface {
	Enqueue(track Track) error
	Dequeue() (Track, error)
	Peek() (Track, error)
	Size() int
	IsEmpty() bool
}

type YouTubeService interface {
	ParseURL(ctx interface{}, url string) (Track, error)
	Search(ctx interface{}, query string, maxResults int) ([]SearchResult, error)
}

type CommandContext struct {
	ProfileName string
	Queue       Queue
	Player      audio.Player
	Config      config.Config
	YtService   YouTubeService
	Output      ui.OutputHandler
	Logger      logger.Logger
}

type Command interface {
	Name() string
	Description() string
	Execute(ctx CommandContext, args []string) error
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
