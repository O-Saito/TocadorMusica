package ui

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"tocadormusica/ports/ui"
)

var _ ui.InputHandler = (*CLIinterface)(nil)
var _ ui.OutputHandler = (*CLIinterface)(nil)

type CLIinterface struct {
	reader       *bufio.Reader
	inputChan    chan string
	responseChan chan string
	waitingType  string
	wg           sync.WaitGroup
	mu           sync.Mutex
	closed       bool
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

		fmt.Print("> ")
		line, err := c.reader.ReadString('\n')
		line = strings.TrimSpace(line)

		fmt.Fprintf(os.Stderr, "[DEBUG CLI] Received: %q, err=%v, waiting=%q\n", line, err, c.waitingType)

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
	fmt.Println(message)
}

func (c *CLIinterface) RequestInput(prompt string) <-chan string {
	c.waitingType = "input"
	return c.responseChan
}

func (c *CLIinterface) DisplayOptions(options []string) <-chan int {
	c.waitingType = "option"
	ch := make(chan int)

	go func() {
		idxStr := <-c.responseChan
		var idx int
		fmt.Sscanf(idxStr, "%d", &idx)
		ch <- idx
	}()

	for i, opt := range options {
		fmt.Printf("  %d: %s\n", i, opt)
	}
	fmt.Print("Select option: ")

	return ch
}

func (c *CLIinterface) FindUnknownCommand() {
	fmt.Println("Unknown command. Available commands:")
	fmt.Println("  add     : Add a track to queue (url or search query)")
	fmt.Println("  next    : Skip to next track in queue")
	fmt.Println("  pause   : Pause current track")
	fmt.Println("  play    : Play the first track in queue")
	fmt.Println("  queue   : Show queue size")
	fmt.Println("  resume  : Resume paused track")
	fmt.Println("  stop    : Stop current track")
	fmt.Println("  volume  : Get/set volume (0-100)")
}
