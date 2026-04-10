package discord

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"

	"tocadormusica/domain"
)

type UI struct {
	bot           *Bot
	perfil        domain.PerfilInterface
	profileName   string
	currentCh     string
	inputChan     chan string
	optionsChan   chan int
	waitingOption bool
	mu            sync.RWMutex
	closed        bool
}

func NewUI(bot *Bot) *UI {
	return &UI{
		bot:         bot,
		profileName: "default",
		inputChan:   make(chan string, 10),
		optionsChan: make(chan int),
	}
}

func (u *UI) SetProfileName(name string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.profileName = name
}

func (u *UI) Refresh() {}

func (u *UI) Run(ctx context.Context) {
	u.bot.AddHandler(u.handleReady)
	u.bot.AddHandler(u.handleMessageCreate)
	u.bot.AddHandler(u.handleVoiceStateUpdate)

	if err := u.bot.Open(); err != nil {
		u.Display("Error: failed to connect to Discord: " + err.Error())
		return
	}

	u.bot.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		u.Display("Connected to Discord! Joining voice channel...")
		if err := u.bot.JoinVoice(); err != nil {
			u.Display("Error: " + err.Error())
			return
		}
		u.Display("Joined voice channel!")
	})

	defer u.bot.Close()

	<-ctx.Done()
}

func (u *UI) handleReady(s *discordgo.Session, r *discordgo.Ready) {
	u.Display("Logged in as: " + s.State.User.Username)
}

func (u *UI) handleVoiceStateUpdate(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	if v.UserID == s.State.User.ID && v.ChannelID != "" {
		u.Display("Bot joined voice channel: " + v.ChannelID)
	}
}

func (u *UI) handleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	content := strings.TrimSpace(m.Content)
	if content == "" {
		return
	}

	u.mu.Lock()
	if u.closed {
		u.mu.Unlock()
		return
	}
	u.currentCh = m.ChannelID

	if u.waitingOption {
		u.waitingOption = false
		idx, err := strconv.Atoi(content)
		if err != nil {
			u.optionsChan <- -1
		} else {
			u.optionsChan <- idx - 1
		}
		u.mu.Unlock()
		return
	}
	u.mu.Unlock()

	select {
	case u.inputChan <- content:
	default:
	}
}

func (u *UI) Input() <-chan string {
	return u.inputChan
}

func (u *UI) Close() {
	u.mu.Lock()
	defer u.mu.Unlock()
	if !u.closed {
		u.closed = true
		close(u.inputChan)
		close(u.optionsChan)
	}
}

func (u *UI) channelID() string {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.currentCh
}

func (u *UI) send(message string) {
	u.bot.Session().ChannelMessageSend(u.channelID(), message)
}

func (u *UI) Display(message string) {
	u.send(message)
}

func (u *UI) RequestInput(prompt string) <-chan string {
	u.send(prompt)
	return u.inputChan
}

func (u *UI) DisplayOptions(options []string) <-chan int {
	return u.DisplayOptionsPage(options, 0, 0, false)
}

func (u *UI) DisplayOptionsPage(options []string, currentPage int, totalPages int, showYouTubeOption bool) <-chan int {
	var sb strings.Builder
	if totalPages > 0 {
		sb.WriteString(fmt.Sprintf("Page %d/%d\n", currentPage+1, totalPages))
	}
	for i, opt := range options {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, opt))
	}
	if showYouTubeOption {
		sb.WriteString(fmt.Sprintf("%d. Search on YouTube\n", len(options)+1))
	}
	u.send(sb.String())

	u.mu.Lock()
	u.waitingOption = true
	u.mu.Unlock()

	return u.optionsChan
}

func (u *UI) FindUnknownCommand() {
	u.send("Unknown command. Available: play, pause, resume, stop, next, list, queue, volume, add, clear")
}

func (u *UI) ShowQueue(items []string) {
	var sb strings.Builder
	sb.WriteString("Queue:\n")
	for i, item := range items {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, item))
	}
	u.send(sb.String())
}

func (u *UI) ShowNowPlaying(track string) {
	u.send("Now playing: " + track)
}

func (u *UI) ShowVolumeAndAutoplay(volume int, autoplay bool) {
	autoplayStr := "off"
	if autoplay {
		autoplayStr = "on"
	}
	u.send(fmt.Sprintf("Volume: %d%% | Autoplay: %s", volume, autoplayStr))
}

func (u *UI) ShowBackground(track string, position int, isPlaying bool, isPaused bool) {
	status := "stopped"
	if isPlaying && !isPaused {
		status = "playing"
	} else if isPaused {
		status = "paused"
	}
	u.send(fmt.Sprintf("%s - %s [%d]", track, status, position))
}

func (u *UI) SetPerfil(perfil domain.PerfilInterface) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.perfil = perfil
}

func (u *UI) NotifyVolumeChanged(volume int) {}

func (u *UI) NotifyTrackChanged(track string) {}

func (u *UI) NotifyPaused() {}

func (u *UI) NotifyPlaying() {}

func (u *UI) NotifyQueueChanged(length int) {}

func (u *UI) NotifyAutoPlayChanged(enabled bool) {}
