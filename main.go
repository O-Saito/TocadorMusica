package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	audioadapter "tocadormusica/adapters/audio"
	cliui "tocadormusica/adapters/ui"
	"tocadormusica/commands"
	"tocadormusica/config"
	"tocadormusica/domain"
	"tocadormusica/logger"
	"tocadormusica/services/yt-dlp"
)

type cmdExecutor struct {
	perfil domain.PerfilInterface
}

func (e *cmdExecutor) ExecuteCommand(name string, args []string) {
	cmd := commands.Get(name)
	if cmd != nil {
		err := cmd.Execute(e.perfil, args)
		if err != nil {
			e.perfil.Output().Display("Error: " + err.Error())
		}
		e.perfil.Output().Refresh()
	} else {
		e.perfil.Output().FindUnknownCommand()
	}
}

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

	ytService := ytdlp.NewWithRunnerAndLogger(nil, log)

	queue := domain.NewQueue(global.MaxQueueSize)

	player, err := audioadapter.NewOtoPlayer(global.SampleRate, log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create audio player: %v\n", err)
		os.Exit(1)
	}
	player.SetVolume(profile.Volume)

	cliinterface := cliui.NewCLIinterface()
	cliinterface.SetProfileName(profileName)

	perfil := domain.NewPerfil(
		profileName,
		queue,
		player,
		cliinterface,
		cliinterface,
		ytService,
		cfg,
		log,
		nil,
	)

	executor := &cmdExecutor{perfil: perfil}
	perfil.SetCommandExecutor(executor)

	volume := int(player.Volume() * 100)
	perfil.NotifyVolumeChanged(volume)

	cliinterface.Refresh()

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
