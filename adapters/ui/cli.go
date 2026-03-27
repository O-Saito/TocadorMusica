package ui

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"tocadormusica/commands"
	"tocadormusica/ports/ui"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

var _ ui.InputHandler = (*CLIinterface)(nil)
var _ ui.OutputHandler = (*CLIinterface)(nil)

const (
	BoxWidth    = 103
	TitleMaxLen = 22
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			Width(BoxWidth - 2).
			Align(lipgloss.Center)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("69"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82"))

	greenStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82"))

	redStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("69")).
			Width(BoxWidth - 2)

	searchBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("75")).
			Width(BoxWidth - 2)

	nowPlayingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Width(BoxWidth - 2)

	messageStyle = lipgloss.NewStyle().
			Width(BoxWidth - 2)

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("75"))
)

type CLIinterface struct {
	reader       *bufio.Reader
	inputChan    chan string
	responseChan chan string
	waitingType  string
	wg           sync.WaitGroup
	mu           sync.Mutex
	closed       bool

	perfil interface {
		GetBackgroundTrack() string
		GetBackgroundPosition() int
		IsBackgroundPlaying() bool
		IsBackgroundPaused() bool
	}
	profileName        string
	queue              []string
	nowPlaying         string
	duration           string
	isSearch           bool
	lastMessage        string
	isError            bool
	isPaused           bool
	volume             int
	autoplay           bool
	backgroundTrack    string
	backgroundPosition int
	backgroundStatus   string
}

func NewCLIinterface() *CLIinterface {
	return &CLIinterface{
		reader:       bufio.NewReader(os.Stdin),
		inputChan:    make(chan string, 10),
		responseChan: make(chan string),
	}
}

func (c *CLIinterface) Input() <-chan string {
	return c.inputChan
}

func (c *CLIinterface) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.closed {
		c.closed = true
		close(c.inputChan)
		close(c.responseChan)
	}
}

func (c *CLIinterface) Run(ctx context.Context) {
	c.wg.Add(1)
	defer c.wg.Done()

	for {
		select {
		case <-ctx.Done():
			c.Close()
			return
		default:
		}

		line, err := c.reader.ReadString('\n')
		line = strings.TrimSpace(line)

		if err != nil {
			continue
		}

		if c.waitingType != "" {
			c.responseChan <- line
			c.waitingType = ""
		} else if line != "" {
			select {
			case c.inputChan <- line:
			case <-ctx.Done():
				c.Close()
				return
			}
		}
	}
}

func (c *CLIinterface) Display(message string) {
	c.mu.Lock()
	c.lastMessage = message
	c.isError = strings.HasPrefix(message, "Error:")
	c.mu.Unlock()

	c.RenderBox()
}

func (c *CLIinterface) RequestInput(prompt string) <-chan string {
	c.waitingType = "input"
	return c.responseChan
}

func (c *CLIinterface) DisplayOptions(options []string) <-chan int {
	return c.DisplayOptionsPage(options, 0, 0, false)
}

func (c *CLIinterface) DisplayOptionsPage(options []string, currentPage int, totalPages int, showYouTubeOption bool) <-chan int {
	c.waitingType = "option"
	ch := make(chan int)

	go func() {
		idxStr := <-c.responseChan
		idxStr = strings.TrimSpace(idxStr)

		if idxStr == "n" || idxStr == "N" {
			ch <- -1
			return
		}
		if idxStr == "p" || idxStr == "P" {
			ch <- -2
			return
		}
		if idxStr == "yt" || idxStr == "YT" || idxStr == "ytube" || idxStr == "YTUBE" {
			ch <- -3
			return
		}
		if idxStr == "q" || idxStr == "Q" || idxStr == "quit" || idxStr == "QUIT" {
			ch <- -4
			return
		}

		var idx int
		fmt.Sscanf(idxStr, "%d", &idx)
		if showYouTubeOption && idx == 0 {
			ch <- -3
		} else {
			ch <- idx - 1
		}
	}()

	c.isSearch = true
	c.clearScreen()
	c.renderSearchPage(options, currentPage, totalPages, showYouTubeOption)

	return ch
}

func (c *CLIinterface) FindUnknownCommand() {
	c.RenderBox()
	fmt.Println("Unknown command")
}

func (c *CLIinterface) ShowQueue(items []string) {
	c.mu.Lock()
	c.queue = items
	c.isSearch = false
	c.mu.Unlock()
	c.RenderBox()
}

func (c *CLIinterface) ShowNowPlaying(track string) {
	c.mu.Lock()
	c.nowPlaying = track
	c.isPaused = track == ""
	c.isSearch = false
	c.mu.Unlock()
	c.RenderBox()
}

func (c *CLIinterface) ShowNowPlayingInfo(track string, isPaused bool, duration string) {
	c.mu.Lock()
	c.nowPlaying = track
	c.isPaused = isPaused
	c.duration = duration
	c.isSearch = false
	c.mu.Unlock()
	c.RenderBox()
}

func (c *CLIinterface) Refresh() {
	c.mu.Lock()
	c.isSearch = false
	c.mu.Unlock()
	c.RenderBox()
}

func (c *CLIinterface) ShowVolumeAndAutoplay(volume int, autoplay bool) {
	c.mu.Lock()
	c.volume = volume
	c.autoplay = autoplay
	c.mu.Unlock()
	c.RenderBox()
}

func (c *CLIinterface) SetPerfil(perfil interface {
	GetBackgroundTrack() string
	GetBackgroundPosition() int
	IsBackgroundPlaying() bool
	IsBackgroundPaused() bool
}) {
	c.mu.Lock()
	c.perfil = perfil
	c.mu.Unlock()
}

func (c *CLIinterface) ShowBackground(track string, position int, isPlaying bool, isPaused bool) {
	c.mu.Lock()
	c.backgroundTrack = track
	c.backgroundPosition = position
	if track == "" {
		c.backgroundStatus = ""
	} else if isPlaying {
		c.backgroundStatus = "playing"
	} else if isPaused {
		c.backgroundStatus = "paused"
	} else {
		c.backgroundStatus = "stopped"
	}
	c.mu.Unlock()
	c.RenderBox()
}

func (c *CLIinterface) SetProfileName(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.profileName = name
}

func (c *CLIinterface) clearScreen() {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	} else {
		fmt.Print("\033[2J\033[H")
	}
}

func (c *CLIinterface) truncateTitle(title string, maxLen int) string {
	runes := []rune(title)
	if len(runes) > maxLen {
		return string(runes[:maxLen-3]) + "..."
	}
	return title
}

func (c *CLIinterface) formatTime(seconds int) string {
	mins := seconds / 60
	secs := seconds % 60
	return fmt.Sprintf("%d:%02d", mins, secs)
}

func (c *CLIinterface) renderBox() {
	divider := strings.Repeat("─", BoxWidth-2)

	titleContent := titleStyle.Width(BoxWidth - 2).Render("TOCADOR DE MUSICA")

	autoplayStr := greenStyle.Render("Autoplay")
	if !c.autoplay {
		autoplayStr = redStyle.Render("Autoplay")
	}
	statusContent := fmt.Sprintf("Volume: %d%% | %s", c.volume, autoplayStr)
	statusStyle := lipgloss.NewStyle().
		Width(BoxWidth - 2).
		Foreground(lipgloss.Color("75"))
	statusRow := statusStyle.Render(statusContent)

	var backgroundContent string
	if c.backgroundTrack != "" {
		posStr := c.formatTime(c.backgroundPosition)
		var bgText string
		if c.backgroundStatus == "playing" {
			bgText = fmt.Sprintf("Background: %s (%s)", c.backgroundTrack, posStr)
		} else if c.backgroundStatus == "paused" {
			bgText = fmt.Sprintf("Background: %s (paused at %s)", c.backgroundTrack, posStr)
		} else {
			bgText = fmt.Sprintf("Background: %s (stopped at %s)", c.backgroundTrack, posStr)
		}
		backgroundContent = statusStyle.Width(BoxWidth - 2).Render(bgText)
	} else {
		backgroundContent = statusStyle.Width(BoxWidth - 2).Render("Background: None")
	}

	var nowPlayingContent string
	if c.nowPlaying != "" {
		if c.isPaused {
			nowPlayingContent = nowPlayingStyle.Width(BoxWidth - 2).Render("Paused: " + c.nowPlaying)
		} else {
			title := "Now Playing: " + c.nowPlaying
			padding := BoxWidth - 2 - len(title) - len(c.duration)
			if padding < 0 {
				padding = 0
			}
			nowPlayingContent = nowPlayingStyle.Width(BoxWidth - 2).Render(
				title + strings.Repeat(" ", padding) + c.duration)
		}
	} else {
		nowPlayingContent = nowPlayingStyle.Width(BoxWidth - 2).Render("Now Playing: None")
	}

	queueCmdContent := c.buildQueueCommandsContent()

	var messageContent string
	if c.lastMessage != "" {
		msg := c.truncateTitle(c.lastMessage, BoxWidth-2)
		if c.isError {
			messageContent = errorStyle.Width(BoxWidth - 2).Render("Error: " + msg)
		} else {
			messageContent = successStyle.Width(BoxWidth - 2).Render(msg)
		}
	}

	var content strings.Builder
	content.WriteString(titleContent + "\n")
	content.WriteString(divider + "\n")
	content.WriteString(statusRow + "\n")
	content.WriteString(divider + "\n")
	content.WriteString(backgroundContent + "\n")
	content.WriteString(divider + "\n")
	content.WriteString(nowPlayingContent + "\n")
	content.WriteString(divider + "\n")
	content.WriteString(queueCmdContent)
	if messageContent != "" {
		content.WriteString(divider + "\n")
		content.WriteString(messageContent + "\n")
	}
	content.WriteString(divider)

	c.clearScreen()
	fmt.Println(boxStyle.Render(content.String()))
	fmt.Print("> Enter command: ")
}

func (c *CLIinterface) buildQueueCommandsContent() string {
	queueLen := len(c.queue)
	queueContentWidth := 40

	var queueLines []string
	if queueLen > 0 {
		queueLines = append(queueLines, fmt.Sprintf("Queue (%d):", queueLen))
		for i, item := range c.queue {
			if i >= 10 {
				queueLines = append(queueLines, fmt.Sprintf("  ... +%d more", queueLen-10))
				break
			}
			queueLines = append(queueLines, fmt.Sprintf("  %d. %s", i+1, c.truncateTitle(item, queueContentWidth)))
		}
	} else {
		queueLines = []string{"Queue (0)"}
	}

	cmds := commands.List()
	var commandLines []string
	commandLines = append(commandLines, "Commands:")
	for _, cmd := range cmds {
		commandLines = append(commandLines, fmt.Sprintf("  %-8s : %s", cmd.Name(), cmd.Description()))
	}

	maxLines := len(queueLines)
	if len(commandLines) > maxLines {
		maxLines = len(commandLines)
	}

	t := table.New().
		Width(BoxWidth - 2).
		Border(lipgloss.NormalBorder()).
		BorderColumn(true).
		BorderRow(false).
		BorderLeft(false).
		BorderRight(false).
		BorderTop(false).
		BorderBottom(false).
		Wrap(false).
		StyleFunc(func(row, col int) lipgloss.Style {
			return lipgloss.NewStyle()
		})

	for i := 0; i < maxLines; i++ {
		left := ""
		if i < len(queueLines) {
			left = queueLines[i]
		}
		right := ""
		if i < len(commandLines) {
			right = commandLines[i]
		}
		t.Row(left, right)
	}

	return t.String()
}

func (c *CLIinterface) renderSearch(options []string) {
	c.renderSearchPage(options, 0, 0, false)
}

func (c *CLIinterface) renderSearchPage(options []string, currentPage int, totalPages int, showYouTubeOption bool) {
	centeredTitle := "SEARCH RESULTS"
	divider := strings.Repeat("─", BoxWidth-2)

	var content strings.Builder
	content.WriteString(centeredTitle + "\n")
	content.WriteString(divider + "\n")

	if totalPages > 0 {
		content.WriteString(fmt.Sprintf("Page %d/%d\n", currentPage+1, totalPages))
		content.WriteString(divider + "\n")
	}

	if showYouTubeOption {
		content.WriteString("  0: Search on YouTube\n")
	}

	for i, opt := range options {
		offset := 0
		if showYouTubeOption {
			offset = 1
		}
		truncated := c.truncateTitle(opt, BoxWidth-10)
		content.WriteString(fmt.Sprintf(" %2d: %s\n", i+1+offset, truncated))
	}

	content.WriteString(divider + "\n")

	if totalPages > 0 {
		if showYouTubeOption {
			content.WriteString("(0=yt, n=next, p=prev, q=quit)")
		} else {
			content.WriteString("(n=next, p=prev, q=quit)")
		}
	} else {
		if showYouTubeOption {
			content.WriteString("(0=yt)")
		}
	}

	c.clearScreen()
	fmt.Println(searchBoxStyle.Render(content.String()))
	fmt.Print("Select option: ")
}

func (c *CLIinterface) RenderBox() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.renderBox()
}
