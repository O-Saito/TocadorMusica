package services

import (
	"io"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"sync/atomic"

	"tocadormusica/config"

	"github.com/ebitengine/oto/v3"
)

const (
	Channels = 2
	BitDepth = 16
)

type AudioPlayer struct {
	context  *oto.Context
	player   *oto.Player
	playerMu sync.RWMutex
	playing  int32
	paused   int32
	stopped  int32
	volume   float64
	ffCmd    *exec.Cmd
}

func NewAudioPlayer() (*AudioPlayer, error) {
	op := &oto.NewContextOptions{
		SampleRate:   config.FRAME_RATE,
		ChannelCount: Channels,
		Format:       oto.FormatSignedInt16LE,
		BufferSize:   512,
	}

	ctx, ready, err := oto.NewContext(op)
	if err != nil {
		Error("Failed to create audio context: %v", err)
		return nil, err
	}
	<-ready

	Info("Audio player initialized")
	return &AudioPlayer{context: ctx, volume: 1.0}, nil
}

func (p *AudioPlayer) Play(media string) error {
	Info("Starting playback: %s", media)

	p.Stop()

	atomic.StoreInt32(&p.stopped, 0)

	ffmpegReader, ffmpegWriter := io.Pipe()

	p.ffCmd = exec.Command("ffmpeg",
		"-reconnect", "1",
		"-reconnect_streamed", "1",
		"-reconnect_delay_max", "5",
		"-fflags", "+genpts",
		"-loglevel", "error",
		"-i", media,
		"-f", "s16le",
		"-ar", strconv.Itoa(config.FRAME_RATE),
		"-ac", "2",
		"pipe:1")
	p.ffCmd.Stdout = ffmpegWriter
	p.ffCmd.Stderr = os.Stderr

	if err := p.ffCmd.Start(); err != nil {
		Error("Failed to start ffmpeg: %v", err)
		return err
	}

	p.playerMu.Lock()
	p.player = p.context.NewPlayer(ffmpegReader)
	p.player.SetVolume(p.volume)
	p.player.Play()
	p.playerMu.Unlock()

	atomic.StoreInt32(&p.playing, 1)

	Info("Playback started successfully")

	go func() {
		p.ffCmd.Wait()

		if atomic.LoadInt32(&p.stopped) == 0 {
			atomic.StoreInt32(&p.playing, 0)
			Info("Playback finished")
		}

		ffmpegReader.Close()
	}()

	return nil
}

func (p *AudioPlayer) Pause() {
	p.playerMu.RLock()
	player := p.player
	p.playerMu.RUnlock()

	if player != nil && atomic.LoadInt32(&p.playing) == 1 && atomic.LoadInt32(&p.paused) == 0 {
		player.Pause()
		atomic.StoreInt32(&p.paused, 1)
		Info("Playback paused")
	}
}

func (p *AudioPlayer) Resume() {
	p.playerMu.RLock()
	player := p.player
	p.playerMu.RUnlock()

	if player != nil && atomic.LoadInt32(&p.paused) == 1 {
		player.Play()
		atomic.StoreInt32(&p.paused, 0)
		Info("Playback resumed")
	}
}

func (p *AudioPlayer) Stop() {
	atomic.StoreInt32(&p.stopped, 1)

	p.playerMu.Lock()
	if p.player != nil {
		p.player.Close()
		p.player = nil
	}
	p.playerMu.Unlock()

	if p.ffCmd != nil && p.ffCmd.Process != nil {
		p.ffCmd.Process.Kill()
		p.ffCmd = nil
	}

	atomic.StoreInt32(&p.playing, 0)
	atomic.StoreInt32(&p.paused, 0)
	Info("Playback stopped")
}

func (p *AudioPlayer) IsPlaying() bool {
	return atomic.LoadInt32(&p.playing) == 1 && atomic.LoadInt32(&p.paused) == 0
}

func (p *AudioPlayer) IsPaused() bool {
	return atomic.LoadInt32(&p.paused) == 1
}

func (p *AudioPlayer) Volume() float64 {
	return p.volume
}

func (p *AudioPlayer) SetVolume(v float64) {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	p.volume = v

	p.playerMu.RLock()
	player := p.player
	p.playerMu.RUnlock()

	if player != nil {
		player.SetVolume(v)
	}
	Info("Volume set to %.0f%%", v*100)
}
