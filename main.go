package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	audioadapter "tocadormusica/adapters/audio"
	cliui "tocadormusica/adapters/ui"
	"tocadormusica/commands"
	"tocadormusica/config"
	"tocadormusica/domain"
	"tocadormusica/logger"
	"tocadormusica/ports/audio"
	"tocadormusica/services/yt-dlp"
)

func main() {
	profileName := "main-perfil"

	cfg, err := config.Load(".config")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid config: %v\n", err)
		os.Exit(1)
	}

	global, profile := cfg.GetProfile(profileName)

	startTime := time.Now().Format("2006-01-02T15-04:05")

	log, closer, err := logger.NewWithFile(
		profileName,
		logger.DEBUG,
		logger.ParseLevel(global.LogLevel),
		startTime,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer closer.Close()

	log.Info("application starting", "volume", profile.Volume, "max_queue", global.MaxQueueSize, "sample_rate", global.SampleRate)

	ytService := ytdlp.New()

	queue := domain.NewQueue(global.MaxQueueSize)

	player, err := audioadapter.NewOtoPlayer(global.SampleRate, log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create audio player: %v\n", err)
		os.Exit(1)
	}
	player.SetVolume(profile.Volume)

	cliinterface := cliui.NewCLIinterface()

	showHelp(cliinterface)

	perfil := domain.NewPerfil(
		profileName,
		queue,
		player,
		cliinterface,
		cliinterface,
		ytService,
		cfg,
		log,
	)

	ctx, cancel := context.WithCancel(context.Background())

	if err := perfil.Start(ctx); err != nil {
		log.Error("failed to start perfil", "error", err)
		return
	}

	go cliinterface.Run(ctx)

	log.Info("application ready, press Ctrl+C to exit")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("shutting down...")
	cancel()
	perfil.Wait()
	log.Info("shutdown complete")
}

func showHelp(cli *cliui.CLIinterface) {
	cli.Display("===== Tocador de Musica =====")
	for _, cmd := range commands.List() {
		cli.Display(fmt.Sprintf("  %-8s: %s", cmd.Name(), cmd.Description()))
	}
}

func handleCommand(input, profileName string, queue domain.Queue, player audio.Player, cfg config.Config, ytService domain.YouTubeService, cli *cliui.CLIinterface, log logger.Logger) {
	log.Info("received input", "input", input)

	parts := strings.Fields(input)
	if len(parts) == 0 {
		return
	}

	commandName := parts[0]
	args := parts[1:]

	cmd := commands.Get(commandName)
	if cmd == nil {
		cli.Display("Unknown command")
		showHelp(cli)
		return
	}

	cmd.Execute(commands.CommandContext{
		ProfileName: profileName,
		Output:      cli,
		Logger:      log,
	}, args)
}
