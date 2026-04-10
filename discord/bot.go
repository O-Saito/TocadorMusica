package discord

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	voiceChannelID = "703658058437361729"
	guildID        = "703658058437361724"
)

type Bot struct {
	session     *discordgo.Session
	token       string
	voiceConn   *discordgo.VoiceConnection
	voiceJoined bool
}

func NewBot(token string) (*Bot, error) {
	if token == "" {
		return nil, fmt.Errorf("discord token is required")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord session: %w", err)
	}

	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildVoiceStates

	return &Bot{
		session: dg,
		token:   token,
	}, nil
}

func (b *Bot) Session() *discordgo.Session {
	return b.session
}

func (b *Bot) Open() error {
	return b.session.Open()
}

func (b *Bot) Close() error {
	if b.voiceConn != nil {
		b.voiceConn.Disconnect()
	}
	return b.session.Close()
}

func (b *Bot) AddHandler(handler interface{}) {
	b.session.AddHandler(handler)
}

func (b *Bot) JoinVoice() error {
	if b.voiceConn != nil {
		return nil
	}

	vc, err := b.session.ChannelVoiceJoin(guildID, voiceChannelID, false, true)
	if err != nil {
		return fmt.Errorf("failed to join voice channel: %w", err)
	}

	b.voiceConn = vc

	for i := 0; i < 20; i++ {
		if vc.Ready {
			b.voiceJoined = true
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for voice connection ready")
}

func (b *Bot) IsVoiceJoined() bool {
	return b.voiceJoined
}

func (b *Bot) WaitForSignal() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
