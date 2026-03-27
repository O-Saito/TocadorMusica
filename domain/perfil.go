package domain

import (
	"context"
	"strings"
	"time"

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
	FileService() FileService
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
	SetBackground(trackPath string) error
	StopBackground()
	ClearBackground()
	PauseBackground()
	ResumeBackground() error
	StartBackground() error
	IsBackgroundPlaying() bool
	IsBackgroundPaused() bool
	GetBackgroundTrack() Track
	GetBackgroundPosition() int
	SetBackgroundPosition(position int)
}

type CommandExecutor interface {
	ExecuteCommand(name string, args []string)
}

type perfil struct {
	name                string
	queue               Queue
	player              audio.Player
	input               ui.InputHandler
	output              ui.OutputHandler
	ytSvc               YouTubeService
	fileSvc             FileService
	cfg                 config.Config
	log                 logger.Logger
	cmdExecutor         CommandExecutor
	ctx                 context.Context
	cancel              context.CancelFunc
	done                chan struct{}
	backgroundTrack     Track
	backgroundPosition  int
	backgroundIsActive  bool
	backgroundStartedAt time.Time
	backgroundPaused    bool
}

func NewPerfil(
	name string,
	queue Queue,
	player audio.Player,
	input ui.InputHandler,
	output ui.OutputHandler,
	ytSvc YouTubeService,
	fileSvc FileService,
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
		fileSvc:     fileSvc,
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

func (p *perfil) FileService() FileService {
	return p.fileSvc
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

func (p *perfil) SetBackground(trackPath string) error {
	p.backgroundTrack = NewTrackFromFile(trackPath)
	p.backgroundPosition = 0
	p.backgroundIsActive = true
	p.backgroundStartedAt = time.Now()
	p.backgroundPaused = false

	global, _ := p.cfg.GetProfile(p.Name())
	err := p.player.PlayURLWithSeek(trackPath, global.SampleRate, 0)
	if err != nil {
		p.backgroundIsActive = false
		return err
	}

	p.player.SetOnFinishedCallback(func() {
		if p.backgroundIsActive {
			p.backgroundPosition = 0
			p.StartBackground()
		}
	})

	p.Output().ShowBackground(p.backgroundTrack.Title(), 0, true, false)
	p.Output().Display("Background music set: " + p.backgroundTrack.Title())
	return nil
}

func (p *perfil) StopBackground() {
	if p.backgroundIsActive && p.player.IsPlaying() {
		p.backgroundPosition = int(time.Since(p.backgroundStartedAt).Seconds())
	}
	title := p.backgroundTrack.Title()
	pos := p.backgroundPosition
	p.player.Stop()
	p.backgroundIsActive = false
	p.Output().ShowBackground(title, pos, false, false)
}

func (p *perfil) ClearBackground() {
	p.player.Stop()
	p.backgroundIsActive = false
	p.backgroundPaused = false
	p.backgroundTrack = Track{}
	p.backgroundPosition = 0
	p.Output().ShowBackground("", 0, false, false)
	p.Output().Display("Background music cleared")
}

func (p *perfil) PauseBackground() {
	if p.backgroundIsActive && p.player.IsPlaying() {
		p.backgroundPosition = int(time.Since(p.backgroundStartedAt).Seconds())
		title := p.backgroundTrack.Title()
		pos := p.backgroundPosition
		p.player.Pause()
		p.backgroundPaused = true
		p.Output().ShowBackground(title, pos, false, true)
		p.Output().Display("Background paused")
	}
}

func (p *perfil) ResumeBackground() error {
	if p.backgroundPaused && p.backgroundTrack.Title() != "" {
		p.backgroundPaused = false
		p.backgroundStartedAt = time.Now()
		p.backgroundStartedAt = p.backgroundStartedAt.Add(-time.Duration(p.backgroundPosition) * time.Second)

		global, _ := p.cfg.GetProfile(p.Name())
		err := p.player.PlayURLWithSeek(p.backgroundTrack.AudioURL(), global.SampleRate, p.backgroundPosition)
		if err != nil {
			return err
		}

		p.Output().ShowBackground(p.backgroundTrack.Title(), p.backgroundPosition, true, false)
		p.Output().Display("Background resumed")
	}
	return nil
}

func (p *perfil) IsBackgroundPaused() bool {
	return p.backgroundPaused
}

func (p *perfil) IsBackgroundPlaying() bool {
	return p.backgroundIsActive && p.player.IsPlaying()
}

func (p *perfil) GetBackgroundTrack() Track {
	return p.backgroundTrack
}

func (p *perfil) GetBackgroundPosition() int {
	return p.backgroundPosition
}

func (p *perfil) SetBackgroundPosition(position int) {
	p.backgroundPosition = position
}

func (p *perfil) StartBackground() error {
	if p.backgroundTrack.Title() == "" {
		return nil
	}

	p.backgroundIsActive = true
	p.backgroundStartedAt = time.Now()

	global, _ := p.cfg.GetProfile(p.Name())
	err := p.player.PlayURLWithSeek(p.backgroundTrack.AudioURL(), global.SampleRate, p.backgroundPosition)
	if err != nil {
		p.backgroundIsActive = false
		return err
	}

	p.player.SetOnFinishedCallback(func() {
		if p.backgroundIsActive {
			p.backgroundPosition = 0
			p.StartBackground()
		}
	})

	p.Output().ShowBackground(p.backgroundTrack.Title(), p.backgroundPosition, true, false)
	p.Output().Display("Background music resumed")
	return nil
}
