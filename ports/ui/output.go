package ui

type OutputHandler interface {
	Display(message string)
	RequestInput(prompt string) <-chan string
	DisplayOptions(options []string) <-chan int
	DisplayOptionsPage(options []string, currentPage int, totalPages int, showYouTubeOption bool) <-chan int
	FindUnknownCommand()
	ShowQueue(items []string)
	ShowNowPlaying(track string)
	ShowVolumeAndAutoplay(volume int, autoplay bool)
	ShowBackground(track string, position int, isPlaying bool, isPaused bool)
	Refresh()
}
