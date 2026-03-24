package ui_cli_socket

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"tocadormusica/commands"
	"tocadormusica/domain"
	"tocadormusica/ports/ui"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

var _ ui.InputHandler = (*CLIWebSocket)(nil)
var _ ui.OutputHandler = (*CLIWebSocket)(nil)

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

	yellowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226"))

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

type CLIWebSocket struct {
	reader       *bufio.Reader
	inputChan    chan string
	responseChan chan string
	waitingType  string
	wg           sync.WaitGroup
	mu           sync.Mutex
	closed       bool

	perfil         domain.PerfilInterface
	profileName    string
	queue          []string
	nowPlaying     string
	duration       string
	isSearch       bool
	lastMessage    string
	isError        bool
	isPaused       bool
	volume         int
	autoplay       bool
	socketStatus   int
	extraCommands  []string
	wsClient       *wsClient
	defaultAddress string
}

const (
	SocketStatusDisconnected = 0
	SocketStatusConnecting   = 1
	SocketStatusReconnecting = 2
	SocketStatusConnected    = 3
)

func NewCLIWebSocket(defaultAddress string) *CLIWebSocket {
	cliWS := &CLIWebSocket{
		reader:         bufio.NewReader(os.Stdin),
		inputChan:      make(chan string, 10),
		responseChan:   make(chan string),
		socketStatus:   SocketStatusDisconnected,
		defaultAddress: defaultAddress,
		extraCommands: []string{
			"  address     : Set/show WebSocket address",
			"  connect     : Connect to WebSocket server (ws://ip:port/path)",
			"  disconnect  : Disconnect from WebSocket server",
			"  status      : Show WebSocket connection status",
		},
		wsClient: newWSClient(),
	}
	cliWS.wsClient.SetCallbacks(
		cliWS.onWSConnect,
		cliWS.onWSDisconnect,
		cliWS.onWSReconnecting,
	)
	cliWS.wsClient.SetMessageCallback(cliWS.onWSMessage)
	return cliWS
}

func (c *CLIWebSocket) SetPerfil(perfil domain.PerfilInterface) {
	c.perfil = perfil
}

func (c *CLIWebSocket) Input() <-chan string {
	return c.inputChan
}

func (c *CLIWebSocket) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.closed {
		c.closed = true
		close(c.inputChan)
		close(c.responseChan)
	}
}

func (c *CLIWebSocket) Run(ctx context.Context) {
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

		if strings.HasPrefix(line, "connect ") {
			c.handleConnect(strings.Fields(line)[1:])
			continue
		}

		if strings.TrimSpace(line) == "connect" {
			c.handleConnect([]string{})
			continue
		}

		if strings.HasPrefix(line, "address ") {
			c.handleAddress(strings.Fields(line)[1:])
			continue
		}

		if strings.TrimSpace(line) == "address" {
			c.handleAddress([]string{})
			continue
		}

		if strings.TrimSpace(line) == "disconnect" {
			c.handleDisconnect()
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

func (c *CLIWebSocket) Display(message string) {
	c.mu.Lock()
	c.lastMessage = message
	c.isError = strings.HasPrefix(message, "Error:")
	c.mu.Unlock()

	c.RenderBox()
}

func (c *CLIWebSocket) RequestInput(prompt string) <-chan string {
	c.waitingType = "input"
	return c.responseChan
}

func (c *CLIWebSocket) DisplayOptions(options []string) <-chan int {
	c.waitingType = "option"
	ch := make(chan int)

	go func() {
		idxStr := <-c.responseChan
		var idx int
		fmt.Sscanf(idxStr, "%d", &idx)
		ch <- idx - 1
	}()

	c.isSearch = true
	c.clearScreen()
	c.renderSearch(options)

	return ch
}

func (c *CLIWebSocket) FindUnknownCommand() {
	c.RenderBox()
	fmt.Println("Unknown command")
}

func (c *CLIWebSocket) ShowQueue(items []string) {
	c.mu.Lock()
	c.queue = items
	c.isSearch = false
	c.mu.Unlock()
	c.RenderBox()
}

func (c *CLIWebSocket) ShowNowPlaying(track string) {
	c.mu.Lock()
	c.nowPlaying = track
	c.isPaused = track == ""
	c.isSearch = false
	c.mu.Unlock()
	c.RenderBox()
}

func (c *CLIWebSocket) ShowNowPlayingInfo(track string, isPaused bool, duration string) {
	c.mu.Lock()
	c.nowPlaying = track
	c.isPaused = isPaused
	c.duration = duration
	c.isSearch = false
	c.mu.Unlock()
	c.RenderBox()
}

func (c *CLIWebSocket) Refresh() {
	c.mu.Lock()
	c.isSearch = false
	c.mu.Unlock()
	c.RenderBox()
}

func (c *CLIWebSocket) ShowVolumeAndAutoplay(volume int, autoplay bool) {
	c.mu.Lock()
	c.volume = volume
	c.autoplay = autoplay
	c.mu.Unlock()
	c.RenderBox()
}

func (c *CLIWebSocket) SetProfileName(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.profileName = name
}

func (c *CLIWebSocket) GetExtraCommands() []string {
	return c.extraCommands
}

func (c *CLIWebSocket) GetSocketStatus() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.socketStatus
}

func (c *CLIWebSocket) SetAddress(addr string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.defaultAddress = addr
}

func (c *CLIWebSocket) onWSConnect() {
	c.mu.Lock()
	c.socketStatus = SocketStatusConnected
	c.mu.Unlock()
	c.RenderBox()

	go func() {
		c.wsClient.Send(`{"type":"upgrade-conn","data":{"conn":"musica.lua"}}`)
		c.wsClient.Send(`{"type":"upgrade-conn","data":{"conn":"ignore-broadcast"}}`)
	}()
}

func (c *CLIWebSocket) onWSDisconnect() {
	c.mu.Lock()
	c.socketStatus = SocketStatusDisconnected
	c.mu.Unlock()
	c.RenderBox()
}

func (c *CLIWebSocket) onWSReconnecting() {
	c.mu.Lock()
	c.socketStatus = SocketStatusReconnecting
	c.mu.Unlock()
	c.RenderBox()
}

func (c *CLIWebSocket) onWSMessage(msg string) {
	var wsMsg struct {
		Type string `json:"type"`
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
		Filter string `json:"filter"`
	}

	if err := json.Unmarshal([]byte(msg), &wsMsg); err != nil {
		return
	}

	if wsMsg.Type != "music_request" {
		return
	}

	formattedURL := c.formatYouTubeURL(wsMsg.Data.URL)
	if formattedURL == "" {
		return
	}

	if c.perfil != nil {
		tracks := c.perfil.Queue().All()
		for _, track := range tracks {
			if track.URL() == formattedURL {
				return
			}
		}
	}

	select {
	case c.inputChan <- "add " + formattedURL:
	default:
	}
}

func (c *CLIWebSocket) formatYouTubeURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	host := parsed.Host
	if host != "www.youtube.com" && host != "youtube.com" && host != "youtu.be" {
		return ""
	}

	if host == "youtu.be" {
		videoID := strings.TrimPrefix(parsed.Path, "/")
		if videoID == "" {
			return ""
		}
		return "https://www.youtube.com/watch?v=" + videoID
	}

	if parsed.Path != "/watch" {
		return ""
	}

	videoID := parsed.Query().Get("v")
	if videoID == "" {
		return ""
	}

	return "https://www.youtube.com/watch?v=" + videoID
}

func (c *CLIWebSocket) handleConnect(args []string) {
	addr := ""
	if len(args) == 0 {
		if c.defaultAddress == "" {
			c.Display("Error: no address configured. Use 'address ws://ip:port/path' to set one.")
			return
		}
		addr = c.defaultAddress
	} else {
		addr = args[0]
	}

	if !strings.Contains(addr, "://") {
		c.Display("Error: usage: connect ws://ip:port/path")
		return
	}

	if c.perfil != nil {
		c.perfil.Config().SetCustomData(c.profileName, "cliwebsocket", map[string]string{"address": addr})
	}

	c.mu.Lock()
	c.socketStatus = SocketStatusConnecting
	c.mu.Unlock()
	c.RenderBox()

	if err := c.wsClient.Connect(addr); err != nil {
		c.mu.Lock()
		c.socketStatus = SocketStatusDisconnected
		c.mu.Unlock()
		c.Display("Error: " + err.Error())
	}
}

func (c *CLIWebSocket) handleAddress(args []string) {
	_, profileCustomData := c.perfil.Config().GetCustomData(c.profileName)
	currentAddr := ""
	if profileCustomData != nil {
		if data, ok := profileCustomData["cliwebsocket"]; ok {
			currentAddr = data["address"]
		}
	}

	if len(args) == 0 {
		if currentAddr == "" {
			c.Display("No address configured. Usage: address ws://ip:port/path")
			return
		}
		c.Display("Current address: " + currentAddr)
		return
	}

	addr := args[0]
	if !strings.Contains(addr, "://") {
		c.Display("Error: usage: address ws://ip:port/path")
		return
	}

	c.perfil.Config().SetCustomData(c.profileName, "cliwebsocket", map[string]string{"address": addr})
	c.SetAddress(addr)
	c.Display("Address saved: " + addr)
}

func (c *CLIWebSocket) handleDisconnect() {
	c.wsClient.Disconnect()
}

func (c *CLIWebSocket) clearScreen() {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	} else {
		fmt.Print("\033[2J\033[H")
	}
}

func (c *CLIWebSocket) truncateTitle(title string, maxLen int) string {
	runes := []rune(title)
	if len(runes) > maxLen {
		return string(runes[:maxLen-3]) + "..."
	}
	return title
}

func (c *CLIWebSocket) renderBox() {
	divider := strings.Repeat("─", BoxWidth-2)

	titleContent := titleStyle.Width(BoxWidth - 2).Render("TOCADOR DE MUSICA (socket)")

	autoplayStr := greenStyle.Render("Autoplay")
	if !c.autoplay {
		autoplayStr = redStyle.Render("Autoplay")
	}

	socketStatusStr := "Socket: "
	switch c.socketStatus {
	case SocketStatusConnected:
		socketStatusStr += greenStyle.Render("connected")
	case SocketStatusConnecting:
		socketStatusStr += yellowStyle.Render("connecting...")
	case SocketStatusReconnecting:
		socketStatusStr += yellowStyle.Render("reconnecting...")
	default:
		socketStatusStr += redStyle.Render("disconnected")
	}

	statusContent := fmt.Sprintf("Volume: %d%% | %s | %s", c.volume, autoplayStr, socketStatusStr)
	statusStyle := lipgloss.NewStyle().
		Width(BoxWidth - 2).
		Foreground(lipgloss.Color("75"))
	statusRow := statusStyle.Render(statusContent)

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
	if nowPlayingContent != "" {
		content.WriteString(nowPlayingContent + "\n")
		content.WriteString(divider + "\n")
	}
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

func (c *CLIWebSocket) buildQueueCommandsContent() string {
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

	for _, cmd := range c.extraCommands {
		commandLines = append(commandLines, cmd)
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

func (c *CLIWebSocket) renderSearch(options []string) {
	centeredTitle := "SEARCH RESULTS"
	divider := strings.Repeat("─", BoxWidth-2)

	var content strings.Builder
	content.WriteString(centeredTitle + "\n")
	content.WriteString(divider + "\n")

	for i, opt := range options {
		truncated := c.truncateTitle(opt, BoxWidth-10)
		content.WriteString(fmt.Sprintf(" %2d: %s\n", i+1, truncated))
	}

	content.WriteString(divider)

	c.clearScreen()
	fmt.Println(searchBoxStyle.Render(content.String()))
	fmt.Print("Select option: ")
}

func (c *CLIWebSocket) RenderBox() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.renderBox()
}
