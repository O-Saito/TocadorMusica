package ui

type InputHandler interface {
	Input() <-chan string
	Close()
}
