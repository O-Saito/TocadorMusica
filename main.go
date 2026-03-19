package main

import (
	"context"
	"flag"
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
	"tocadormusica/ports/ui"
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

type Flags struct {
	Interface string
	Profile   string
}

func ParseFlags() *Flags {
	interfaceName := flag.String("interface", "cli", "UI interface to use (cli)")
	profileName := flag.String("profile", "main-perfil", "Profile name to use")
	flag.Parse()

	return &Flags{
		Interface: *interfaceName,
		Profile:   *profileName,
	}
}

func GetCLI(profileName string) (ui.InputHandler, ui.OutputHandler, func(context.Context)) {
	cliinterface := cliui.NewCLIinterface()
	cliinterface.SetProfileName(profileName)
	cliinterface.Refresh()

	return cliinterface, cliinterface, func(ctx context.Context) {
		cliinterface.Run(ctx)
	}
}

func main() {
	flags := ParseFlags()
	profileName := flags.Profile

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

	var input ui.InputHandler
	var output ui.OutputHandler
	var runCLI func(context.Context)

	switch flags.Interface {
	case "cli":
		input, output, runCLI = GetCLI(profileName)
	default:
		fmt.Fprintf(os.Stderr, "Unknown interface: %s\n", flags.Interface)
		fmt.Fprintf(os.Stderr, "Available interfaces: cli\n")
		os.Exit(1)
	}

	perfil := domain.NewPerfil(
		profileName,
		queue,
		player,
		input,
		output,
		ytService,
		cfg,
		log,
		nil,
	)

	executor := &cmdExecutor{perfil: perfil}
	perfil.SetCommandExecutor(executor)

	volume := int(player.Volume() * 100)
	perfil.NotifyVolumeChanged(volume)

	ctx, cancel := context.WithCancel(context.Background())

	if err := perfil.Start(ctx); err != nil {
		log.Error("failed to start perfil", "error", err)
		return
	}

	go runCLI(ctx)

	log.Info("application ready, press Ctrl+C to exit")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("shutting down...")
	cancel()
	perfil.Wait()
	log.Info("shutdown complete")
}
