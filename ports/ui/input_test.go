package ui

import (
	"testing"
	"time"
)

func TestInputHandler_Interface(t *testing.T) {
	var _ InputHandler = &mockInputHandler{}
}

type mockInputHandler struct {
	inputChan chan string
}

func (m *mockInputHandler) Input() <-chan string {
	return m.inputChan
}

func (m *mockInputHandler) Close() {
	close(m.inputChan)
}

func TestInputHandler_InputChannel(t *testing.T) {
	inputChan := make(chan string, 1)
	handler := &mockInputHandler{inputChan: inputChan}

	inputChan <- "test command"

	select {
	case msg := <-handler.Input():
		if msg != "test command" {
			t.Errorf("expected 'test command', got '%s'", msg)
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for input")
	}
}

func TestInputHandler_Close(t *testing.T) {
	inputChan := make(chan string)
	handler := &mockInputHandler{inputChan: inputChan}

	handler.Close()

	_, open := <-inputChan
	if open {
		t.Error("expected channel to be closed")
	}
}
