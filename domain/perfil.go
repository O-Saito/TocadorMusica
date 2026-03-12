package domain

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"tocadormusica/config"
	"tocadormusica/logger"
	"tocadormusica/ports/audio"
	"tocadormusica/ports/ui"
)

type PerfilInterface interface {
	Name() string
	Queue() Queue
	Player() audio.Player
	Config() config.Config
	YtService() YouTubeService
	Output() ui.OutputHandler
	Logger() logger.Logger
	Context() context.Context
	ExecuteCommand(name string, args []string)
	SetCommandExecutor(exec CommandExecutor)
	Start(ctx context.Context) error
	Wait()
}

type CommandExecutor interface {
	ExecuteCommand(name string, args []string)
}

type perfil struct {
	name        string
	queue       Queue
	player      audio.Player
	input       ui.InputHandler
	output      ui.OutputHandler
	ytSvc       YouTubeService
	cfg         config.Config
	log         logger.Logger
	cmdExecutor CommandExecutor
	ctx         context.Context
	cancel      context.CancelFunc
	done        chan struct{}
}

func NewPerfil(
	name string,
	queue Queue,
	player audio.Player,
	input ui.InputHandler,
	output ui.OutputHandler,
	ytSvc YouTubeService,
	cfg config.Config,
	log logger.Logger,
	cmdExecutor CommandExecutor,
) PerfilInterface {
	return &perfil{
		name:        name,
		queue:       queue,
		player:      player,
		input:       input,
		output:      output,
		ytSvc:       ytSvc,
		cfg:         cfg,
		log:         log.WithProfile(name),
		cmdExecutor: cmdExecutor,
		done:        make(chan struct{}),
	}
}

func (p *perfil) ExecuteCommand(name string, args []string) {
	if p.cmdExecutor != nil {
		p.cmdExecutor.ExecuteCommand(name, args)
	}
}

func (p *perfil) SetCommandExecutor(exec CommandExecutor) {
	p.cmdExecutor = exec
}

func (p *perfil) Name() string {
	return p.name
}

func (p *perfil) Queue() Queue {
	return p.queue
}

func (p *perfil) Player() audio.Player {
	return p.player
}

func (p *perfil) Config() config.Config {
	return p.cfg
}

func (p *perfil) YtService() YouTubeService {
	return p.ytSvc
}

func (p *perfil) Output() ui.OutputHandler {
	return p.output
}

func (p *perfil) Logger() logger.Logger {
	return p.log
}

func (p *perfil) Context() context.Context {
	return p.ctx
}

func (p *perfil) Start(ctx context.Context) error {
	p.ctx, p.cancel = context.WithCancel(ctx)

	go p.run()

	p.log.Info("perfil started")
	return nil
}

func (p *perfil) Wait() {
	<-p.done
}

func (p *perfil) run() {
	defer func() {
		p.cancel()
		close(p.done)
		p.log.Info("perfil stopped")
	}()

	for {
		select {
		case <-p.ctx.Done():
			return
		case input, ok := <-p.input.Input():
			if !ok {
				return
			}
			p.handleInput(input)
		}
	}
}

func (p *perfil) handleInput(input string) {
	p.log.Info("received input", "input", input)

	parts := strings.Fields(input)
	if len(parts) == 0 {
		return
	}

	cmd := parts[0]
	args := parts[1:]

	if cmd == "add" && len(args) == 0 {
		p.output.Display("Usage: add <url or search query>")
		return
	}

	p.ExecuteCommand(cmd, args)
}

func (p *perfil) handleAdd(arg string) {
	if arg == "" {
		p.output.Display("Usage: add <url or search query>")
		return
	}

	if isYouTubeURL(arg) {
		p.handleAddURL(arg)
	} else {
		p.handleSearch(arg)
	}
}

func (p *perfil) handleVolume(arg string) {
	if arg == "" {
		vol := p.player.Volume() * 100
		p.output.Display(fmt.Sprintf("Volume: %.0f%%", vol))
		return
	}

	vol, err := strconv.ParseFloat(arg, 64)
	if err != nil || vol < 0 || vol > 100 {
		p.output.Display("Usage: volume [0-100]")
		return
	}

	p.player.SetVolume(vol / 100)
	p.cfg.SetVolume(p.name, vol/100)
	p.output.Display(fmt.Sprintf("Volume: %.0f%%", vol))
}

func (p *perfil) handleAddURL(url string) {
	p.output.Display("Fetching track...")

	track, err := p.ytSvc.ParseURL(p.ctx, url)
	if err != nil {
		p.output.Display("Error: " + err.Error())
		return
	}

	err = p.queue.Enqueue(track)
	if err != nil {
		p.output.Display("Error adding to queue: " + err.Error())
		return
	}

	p.output.Display("Added: " + track.Title())
}

func (p *perfil) handleSearch(query string) {
	p.output.Display("Searching...")

	_, profile := p.cfg.GetProfile(p.name)
	results, err := p.ytSvc.Search(p.ctx, query, profile.SearchResults)
	if err != nil {
		p.output.Display("Error searching: " + err.Error())
		return
	}

	if len(results) == 0 {
		p.output.Display("No results found")
		return
	}

	titles := make([]string, len(results))
	for i, r := range results {
		titles[i] = fmt.Sprintf("%s - %s", r.Title, r.Duration)
	}

	p.output.Display("Select a track:")
	ch := p.output.DisplayOptions(titles)
	idx := <-ch

	if idx < 0 || idx >= len(results) {
		p.output.Display("Invalid selection")
		return
	}

	p.handleAddURL(results[idx].URL)
}

func isYouTubeURL(input string) bool {
	return strings.Contains(input, "youtube.com") ||
		strings.Contains(input, "youtu.be")
}

func (p *perfil) handlePlay() {
	track, err := p.queue.Peek()
	if err != nil {
		p.output.Display("Queue is empty")
		return
	}

	p.log.Debug("playing track", "title", track.Title(), "url", track.URL(), "audioURL", track.AudioURL())

	if track.AudioURL() == "" {
		p.output.Display("Error: No audio URL available for this track")
		return
	}

	p.output.Display("Streaming: " + track.Title())

	global, _ := p.cfg.GetProfile(p.name)
	err = p.player.PlayURL(track.AudioURL(), global.SampleRate)
	if err != nil {
		p.output.Display("Error: " + err.Error())
		return
	}

	p.output.Display("Playing: " + track.Title())
}

func (p *perfil) handlePause() {
	p.player.Pause()
	p.output.Display("Paused")
}

func (p *perfil) handleResume() {
	p.player.Resume()
	p.output.Display("Resumed")
}

func (p *perfil) handleStop() {
	p.player.Stop()
	p.output.Display("Stopped")
}

func (p *perfil) handleNext() {
	if p.queue.IsEmpty() {
		p.output.Display("Queue is empty")
		return
	}

	_, err := p.queue.Dequeue()
	if err != nil {
		p.output.Display("Error skipping track")
		return
	}

	if !p.queue.IsEmpty() {
		p.handlePlay()
	} else {
		p.output.Display("Queue is empty")
		p.player.Stop()
	}
}

func (p *perfil) handleQueue() {
	size := p.queue.Size()
	p.output.Display(fmt.Sprintf("Queue size: %d", size))
}
