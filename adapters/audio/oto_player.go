package audio

import (
	"bytes"
	"io"
	"os/exec"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/ebitengine/oto/v3"
	"tocadormusica/logger"
	"tocadormusica/ports/audio"
)

var _ audio.Player = (*OtoPlayer)(nil)

type OtoPlayer struct {
	ctx                *oto.Context
	ready              chan struct{}
	player             *oto.Player
	volume             float64
	playing            int32
	stopped            int32
	ffCmd              *exec.Cmd
	volumeMu           sync.Mutex
	playerMu           sync.Mutex
	ffmpegMu           sync.Mutex
	log                logger.Logger
	onFinishedCallback func()
	ffmpegPath         string
}

func NewOtoPlayer(sampleRate int, log logger.Logger) (*OtoPlayer, error) {
	return NewOtoPlayerWithFFmpeg(sampleRate, log, "ffmpeg")
}

func NewOtoPlayerWithFFmpeg(sampleRate int, log logger.Logger, ffmpegPath string) (*OtoPlayer, error) {
	opts := &oto.NewContextOptions{
		SampleRate:   sampleRate,
		ChannelCount: 2,
		Format:       oto.FormatSignedInt16LE,
		BufferSize:   512,
	}

	ctx, ready, err := oto.NewContext(opts)
	if err != nil {
		return nil, err
	}

	<-ready

	return &OtoPlayer{
		ctx:        ctx,
		ready:      ready,
		volume:     1.0,
		log:        log,
		ffmpegPath: ffmpegPath,
	}, nil
}

func (p *OtoPlayer) Play(reader io.Reader) error {
	// This method expects an already streaming reader
	// For URL-based streaming, use PlayURL
	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	if p.player != nil {
		p.player.Close()
	}

	p.player = p.ctx.NewPlayer(reader)
	if p.player == nil {
		return nil
	}

	p.player.SetVolume(p.volume)
	p.player.Play()
	atomic.StoreInt32(&p.playing, 1)

	return nil
}

func (p *OtoPlayer) PlayURL(url string, sampleRate int) error {
	p.Stop()

	atomic.StoreInt32(&p.stopped, 0)

	ffmpegReader, ffmpegWriter := io.Pipe()
	stderrBuf := &bytes.Buffer{}

	p.ffmpegMu.Lock()
	p.ffCmd = exec.Command(p.ffmpegPath,
		"-reconnect", "1",
		"-reconnect_streamed", "1",
		"-reconnect_delay_max", "5",
		"-fflags", "+genpts",
		"-loglevel", "error",
		"-i", url,
		"-f", "s16le",
		"-ar", strconv.Itoa(sampleRate),
		"-ac", "2",
		"pipe:1")
	p.ffCmd.Stdout = ffmpegWriter
	p.ffCmd.Stderr = stderrBuf

	if p.log != nil {
		p.log.Debug("starting ffmpeg", "url_len", len(url))
	}

	if err := p.ffCmd.Start(); err != nil {
		p.ffmpegMu.Unlock()
		return err
	}
	p.ffmpegMu.Unlock()

	p.playerMu.Lock()
	p.player = p.ctx.NewPlayer(ffmpegReader)
	p.player.SetVolume(p.volume)
	p.player.Play()
	p.playerMu.Unlock()

	atomic.StoreInt32(&p.playing, 1)

	go func() {
		err := p.ffCmd.Wait()
		ffmpegReader.Close()

		if p.log != nil {
			if stderr := stderrBuf.String(); stderr != "" {
				p.log.Debug("ffmpeg stderr", "output", stderr)
			}
			if err != nil {
				p.log.Debug("ffmpeg finished", "error", err)
			}
		}

		if atomic.LoadInt32(&p.stopped) == 0 {
			atomic.StoreInt32(&p.playing, 0)
			if p.onFinishedCallback != nil {
				p.onFinishedCallback()
			}
		}
	}()

	return nil
}

func (p *OtoPlayer) Pause() {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	if p.player != nil {
		p.player.Pause()
		atomic.StoreInt32(&p.playing, 0)
	}
}

func (p *OtoPlayer) Resume() {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	if p.player != nil {
		p.player.Play()
		atomic.StoreInt32(&p.playing, 1)
	}
}

func (p *OtoPlayer) Stop() {
	if p.onFinishedCallback != nil {
		p.onFinishedCallback = nil
	}

	atomic.StoreInt32(&p.stopped, 1)

	p.playerMu.Lock()
	if p.player != nil {
		p.player.Close()
		p.player = nil
	}
	p.playerMu.Unlock()

	p.ffmpegMu.Lock()
	if p.ffCmd != nil && p.ffCmd.Process != nil {
		p.ffCmd.Process.Kill()
		p.ffCmd = nil
	}
	p.ffmpegMu.Unlock()

	atomic.StoreInt32(&p.playing, 0)
}

func (p *OtoPlayer) SetVolume(volume float64) {
	p.volumeMu.Lock()
	defer p.volumeMu.Unlock()

	p.volume = volume

	p.playerMu.Lock()
	if p.player != nil {
		p.player.SetVolume(volume)
	}
	p.playerMu.Unlock()
}

func (p *OtoPlayer) Volume() float64 {
	p.volumeMu.Lock()
	defer p.volumeMu.Unlock()
	return p.volume
}

func (p *OtoPlayer) IsPlaying() bool {
	return atomic.LoadInt32(&p.playing) == 1
}

func (p *OtoPlayer) SetOnFinishedCallback(fn func()) {
	p.onFinishedCallback = fn
}
