package audio

import (
	"io"
)

type Player interface {
	Play(reader io.Reader) error
	PlayURL(url string, sampleRate int) error
	Pause()
	Resume()
	Stop()
	SetVolume(volume float64)
	Volume() float64
	IsPlaying() bool
	SetOnFinishedCallback(fn func())
}
