package ui

type OutputHandler interface {
	Display(message string)
	RequestInput(prompt string) <-chan string
	DisplayOptions(options []string) <-chan int
	FindUnknownCommand()
	ShowQueue(items []string)
	ShowNowPlaying(track string)
	Refresh()
}
