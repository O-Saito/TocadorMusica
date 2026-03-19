package domain

import (
	"context"
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
	GetQueueItems() []string
	GetNowPlaying() string
	NotifyVolumeChanged(volume int)
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

func (p *perfil) GetQueueItems() []string {
	tracks := p.queue.All()
	items := make([]string, len(tracks))
	for i, track := range tracks {
		items[i] = track.Title()
	}
	return items
}

func (p *perfil) GetNowPlaying() string {
	if p.player.IsPlaying() {
		track, err := p.queue.Peek()
		if err == nil {
			return track.Title()
		}
	}
	return ""
}

func (p *perfil) NotifyVolumeChanged(volume int) {
	_, profile := p.cfg.GetProfile(p.name)
	p.output.ShowVolumeAndAutoplay(volume, profile.AutoPlay)
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

	p.ExecuteCommand(cmd, args)
}
