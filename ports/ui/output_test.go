package ui

import (
	"testing"
	"time"
)

func TestOutputHandler_Interface(t *testing.T) {
	var _ OutputHandler = &mockOutputHandler{}
}

type mockOutputHandler struct {
	messages []string
}

func (m *mockOutputHandler) Display(message string) {
	m.messages = append(m.messages, message)
}

func (m *mockOutputHandler) RequestInput(prompt string) <-chan string {
	ch := make(chan string)
	go func() {
		time.Sleep(10 * time.Millisecond)
		ch <- "mock input"
	}()
	return ch
}

func (m *mockOutputHandler) DisplayOptions(options []string) <-chan int {
	ch := make(chan int)
	go func() {
		time.Sleep(10 * time.Millisecond)
		ch <- 0
	}()
	return ch
}

func (m *mockOutputHandler) FindUnknownCommand() {}

func TestOutputHandler_Display(t *testing.T) {
	h := &mockOutputHandler{}

	h.Display("Hello")
	h.Display("World")

	if len(h.messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(h.messages))
	}
	if h.messages[0] != "Hello" {
		t.Errorf("expected 'Hello', got '%s'", h.messages[0])
	}
	if h.messages[1] != "World" {
		t.Errorf("expected 'World', got '%s'", h.messages[1])
	}
}

func TestOutputHandler_RequestInput(t *testing.T) {
	h := &mockOutputHandler{}

	ch := h.RequestInput("Enter name:")

	select {
	case input := <-ch:
		if input != "mock input" {
			t.Errorf("expected 'mock input', got '%s'", input)
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for input")
	}
}

func TestOutputHandler_DisplayOptions(t *testing.T) {
	h := &mockOutputHandler{}

	options := []string{"Option 1", "Option 2", "Option 3"}
	ch := h.DisplayOptions(options)

	select {
	case idx := <-ch:
		if idx != 0 {
			t.Errorf("expected 0, got %d", idx)
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for option selection")
	}
}
