package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	audioadapter "tocadormusica/adapters/audio"
	cliui "tocadormusica/adapters/ui"
	cliwebsocket "tocadormusica/adapters/ui/cli_socket"
	"tocadormusica/commands"
	"tocadormusica/config"
	discordadapter "tocadormusica/discord"
	"tocadormusica/domain"
	"tocadormusica/logger"
	portsaudio "tocadormusica/ports/audio"
	"tocadormusica/ports/ui"
	"tocadormusica/services/dependencies"
	"tocadormusica/services/file"
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
			e.perfil.Logger().Error("command failed", "command", name, "error", err)
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

func GetCLIWebSocket(profileName string, cfg config.Config) (ui.InputHandler, ui.OutputHandler, func(context.Context)) {
	_, profileCustomData := cfg.GetCustomData(profileName)
	address := ""
	if profileCustomData != nil {
		if cliWSData, ok := profileCustomData["cliwebsocket"]; ok {
			address = cliWSData["address"]
		}
	}

	cliWS := cliwebsocket.NewCLIWebSocket(address)
	cliWS.SetProfileName(profileName)
	cliWS.Refresh()

	return cliWS, cliWS, func(ctx context.Context) {
		cliWS.Run(ctx)
	}
}

func getDiscord(profileName string, cfg config.Config, ffmpegPath string) (portsaudio.Player, ui.InputHandler, ui.OutputHandler, func(context.Context), error) {
	_, profileCustomData := cfg.GetCustomData(profileName)
	token := ""
	if profileCustomData != nil {
		if discordData, ok := profileCustomData["discord"]; ok {
			token = discordData["token"]
		}
	}

	discordBot, err := discordadapter.NewBot(token)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create Discord bot: %w", err)
	}

	discordPlayer, err := discordadapter.NewDiscordPlayer(discordBot, ffmpegPath)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create Discord player: %w", err)
	}

	discordUI := discordadapter.NewUI(discordBot)
	discordUI.SetProfileName(profileName)
	discordUI.Refresh()

	return discordPlayer, discordUI, discordUI, func(ctx context.Context) {
		discordUI.Run(ctx)
	}, nil
}

func checkDependencies() (string, string, string, error) {
	deps := []dependencies.Dependency{
		{Name: "yt-dlp", Required: true, DisplayName: "yt-dlp"},
		{Name: "ffmpeg", Required: true, DisplayName: "ffmpeg"},
		{Name: "deno", Required: false, DisplayName: "deno"},
	}

	ytDlpPath := ""
	ffmpegPath := ""
	denoPath := ""

	reader := bufio.NewReader(os.Stdin)

	for _, dep := range deps {
		found, path := dependencies.FindCommand(dep.Name)
		if found {
			fmt.Printf("%s found at: %s\n", dep.DisplayName, path)
			switch dep.Name {
			case "yt-dlp":
				ytDlpPath = path
			case "ffmpeg":
				ffmpegPath = path
			case "deno":
				denoPath = path
			}
			continue
		}

		if dep.Name == "ffmpeg" {
			fmt.Println("ffmpeg not found.")
			fmt.Println(dependencies.GetInstallMessage("ffmpeg"))
			fmt.Println("Please install ffmpeg and run the program again.")
			os.Exit(1)
		}

		requiredStr := "required"
		if !dep.Required {
			requiredStr = "optional"
		}

		for {
			fmt.Printf("%s not found (%s). Download to ./src? (y/n): ", dep.DisplayName, requiredStr)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)

			if input != "y" && input != "Y" {
				fmt.Println(dependencies.GetInstallMessage(dep.Name))
				if dep.Required {
					fmt.Println("This dependency is required. Please accept to continue.")
					continue
				}
				fmt.Printf("Skipping %s (optional)\n", dep.DisplayName)
				break
			}

			for {
				downloader := dependencies.NewDownloader()
				if err := downloader.Download(&dep); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to download %s: %v\n", dep.Name, err)
					fmt.Println("Download failed. Try again? (y/n)")
					input2, _ := reader.ReadString('\n')
					input2 = strings.TrimSpace(input2)
					if input2 != "y" && input2 != "Y" {
						if dep.Required {
							os.Exit(1)
						}
						fmt.Printf("Skipping %s (optional)\n", dep.DisplayName)
						break
					}
					continue
				}

				found, localPath := dependencies.FindCommand(dep.Name)
				if !found {
					fmt.Fprintf(os.Stderr, "Failed to find %s after download\n", dep.DisplayName)
					if dep.Required {
						os.Exit(1)
					}
					fmt.Printf("Skipping %s (optional)\n", dep.DisplayName)
					break
				}
				switch dep.Name {
				case "yt-dlp":
					ytDlpPath = localPath
				case "ffmpeg":
					ffmpegPath = localPath
				case "deno":
					denoPath = localPath
				}
				break
			}
			break
		}
	}

	return ytDlpPath, ffmpegPath, denoPath, nil
}

func findConfigPath() string {
	execPath, err := os.Executable()
	if err == nil {
		execDirPath := filepath.Join(filepath.Dir(execPath), ".config")
		if _, err := os.Stat(execDirPath); err == nil {
			return execDirPath
		}
	}

	if _, err := os.Stat(".config"); err == nil {
		return ".config"
	}

	if execPath != "" {
		return filepath.Join(filepath.Dir(execPath), ".config")
	}
	return ".config"
}

func main() {
	flags := ParseFlags()
	profileName := flags.Profile

	ytDlpPath, ffmpegPath, _, err := checkDependencies()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Dependency check failed: %v\n", err)
		os.Exit(1)
	}

	tempLog := logger.New(logger.WithLevel("debug"))
	cfg, err := config.Load(findConfigPath(), tempLog)
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

	cfg.SetLogger(log)

	log.Info("application starting", "volume", profile.Volume, "max_queue", global.MaxQueueSize, "sample_rate", global.SampleRate)

	ytService := ytdlp.NewWithBinaryPathAndLogger(ytDlpPath, log)
	fileService := file.New()

	queue := domain.NewQueue(global.MaxQueueSize)

	var player portsaudio.Player
	var input ui.InputHandler
	var output ui.OutputHandler
	var runCLI func(context.Context)

	switch flags.Interface {
	case "cli":
		player, err = audioadapter.NewOtoPlayerWithFFmpeg(global.SampleRate, log, ffmpegPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create audio player: %v\n", err)
			os.Exit(1)
		}
		input, output, runCLI = GetCLI(profileName)
	case "cliwebsocket":
		player, err = audioadapter.NewOtoPlayerWithFFmpeg(global.SampleRate, log, ffmpegPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create audio player: %v\n", err)
			os.Exit(1)
		}
		input, output, runCLI = GetCLIWebSocket(profileName, cfg)
	case "discord":
		discordPlayer, discordInput, discordOutput, discordRun, err := getDiscord(profileName, cfg, ffmpegPath)
		if err != nil {
			log.Error("failed to create discord interface", "error", err)
			os.Exit(1)
		}
		player = discordPlayer
		input = discordInput
		output = discordOutput
		runCLI = discordRun
	default:
		fmt.Fprintf(os.Stderr, "Unknown interface: %s\n", flags.Interface)
		fmt.Fprintf(os.Stderr, "Available interfaces: cli, cliwebsocket, discord\n")
		os.Exit(1)
	}

	perfil := domain.NewPerfil(
		profileName,
		queue,
		player,
		input,
		output,
		ytService,
		fileService,
		cfg,
		log,
		nil,
	)

	executor := &cmdExecutor{perfil: perfil}
	perfil.SetCommandExecutor(executor)

	if cliWS, ok := input.(*cliwebsocket.CLIWebSocket); ok {
		cliWS.SetPerfil(perfil)
	}

	if cli, ok := input.(*cliui.CLIinterface); ok {
		cli.SetPerfil(perfil)
	}

	if discordUI, ok := input.(*discordadapter.UI); ok {
		discordUI.SetPerfil(perfil)
	}

	player.SetVolume(profile.Volume)
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
