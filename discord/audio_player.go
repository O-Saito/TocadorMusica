package discord

import (
	"bytes"
	"io"
	"os/exec"
	"strconv"
	"sync"
	"sync/atomic"

	"tocadormusica/ports/audio"
)

const (
	sampleRate = 48000
	channels   = 2
)

var _ audio.Player = (*DiscordPlayer)(nil)

type DiscordPlayer struct {
	bot        *Bot
	volume     float64
	playing    int32
	stopped    int32
	ffmpegPath string
	onFinished func()
	ffmpegCmd  *exec.Cmd
	volumeMu   sync.Mutex
	playerMu   sync.Mutex
	ffmpegMu   sync.Mutex
}

func NewDiscordPlayer(bot *Bot, ffmpegPath string) (*DiscordPlayer, error) {
	return &DiscordPlayer{
		bot:        bot,
		volume:     1.0,
		ffmpegPath: ffmpegPath,
	}, nil
}

func (p *DiscordPlayer) Play(reader io.Reader) error {
	return nil
}

func (p *DiscordPlayer) PlayURL(url string, sampleRate int) error {
	return p.PlayURLWithSeek(url, sampleRate, 0)
}

func (p *DiscordPlayer) PlayURLWithSeek(url string, sampleRate int, seekSeconds int) error {
	p.Stop()

	atomic.StoreInt32(&p.stopped, 0)

	p.volumeMu.Lock()
	vol := p.volume
	p.volumeMu.Unlock()

	ffmpegReader, ffmpegWriter := io.Pipe()
	stderrBuf := &bytes.Buffer{}

	p.ffmpegMu.Lock()
	cmdArgs := []string{}
	if seekSeconds > 0 {
		cmdArgs = append(cmdArgs, "-ss", strconv.Itoa(seekSeconds))
	}
	cmdArgs = append(cmdArgs,
		"-reconnect", "1",
		"-reconnect_streamed", "1",
		"-reconnect_delay_max", "5",
		"-fflags", "+genpts",
		"-loglevel", "error",
		"-i", url,
		"-vn",
		"-filter:a", "volume="+strconv.FormatFloat(vol, 'f', -1, 64),
		"-c:a", "libopus",
		"-frame_size", "20",
		"-application", "voip",
		"-f", "opus",
		"pipe:1")
	p.ffmpegCmd = exec.Command(p.ffmpegPath, cmdArgs...)
	p.ffmpegCmd.Stdout = ffmpegWriter
	p.ffmpegCmd.Stderr = stderrBuf

	if err := p.ffmpegCmd.Start(); err != nil {
		p.ffmpegMu.Unlock()
		return err
	}
	p.ffmpegMu.Unlock()

	vc := p.bot.VoiceConnection()
	if vc == nil {
		return nil
	}

	if err := vc.Speaking(true); err != nil {
		return err
	}

	atomic.StoreInt32(&p.playing, 1)

	go func() {
		defer ffmpegReader.Close()
		defer vc.Speaking(false)

		buf := make([]byte, 4096)

		for {
			n, err := ffmpegReader.Read(buf)
			if err != nil {
				if err == io.EOF {
					break
				}
				continue
			}

			if n > 0 {
				select {
				case vc.OpusSend <- buf[:n]:
				default:
				}
			}
		}

		if atomic.LoadInt32(&p.stopped) == 0 {
			atomic.StoreInt32(&p.playing, 0)
			if p.onFinished != nil {
				p.onFinished()
			}
		}
	}()

	return nil
}

func (p *DiscordPlayer) Pause() {
	atomic.StoreInt32(&p.playing, 0)

	vc := p.bot.VoiceConnection()
	if vc != nil {
		vc.Speaking(false)
	}
}

func (p *DiscordPlayer) Resume() {
	atomic.StoreInt32(&p.playing, 1)

	vc := p.bot.VoiceConnection()
	if vc != nil {
		vc.Speaking(true)
	}
}

func (p *DiscordPlayer) Stop() {
	if p.onFinished != nil {
		p.onFinished = nil
	}

	atomic.StoreInt32(&p.stopped, 1)

	p.ffmpegMu.Lock()
	if p.ffmpegCmd != nil && p.ffmpegCmd.Process != nil {
		p.ffmpegCmd.Process.Kill()
		p.ffmpegCmd = nil
	}
	p.ffmpegMu.Unlock()

	vc := p.bot.VoiceConnection()
	if vc != nil {
		vc.Speaking(false)
	}

	atomic.StoreInt32(&p.playing, 0)
}

func (p *DiscordPlayer) SetVolume(volume float64) {
	if volume < 0 {
		volume = 0
	}
	if volume > 1 {
		volume = 1
	}

	p.volumeMu.Lock()
	p.volume = volume
	p.volumeMu.Unlock()
}

func (p *DiscordPlayer) Volume() float64 {
	p.volumeMu.Lock()
	defer p.volumeMu.Unlock()
	return p.volume
}

func (p *DiscordPlayer) IsPlaying() bool {
	return atomic.LoadInt32(&p.playing) == 1
}

func (p *DiscordPlayer) SetOnFinishedCallback(fn func()) {
	p.onFinished = fn
}
