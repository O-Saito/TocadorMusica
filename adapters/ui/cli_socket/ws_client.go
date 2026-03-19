package ui_cli_socket

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type wsClient struct {
	conn           *websocket.Conn
	addr           string
	connected      bool
	reconnect      bool
	mu             sync.RWMutex
	onConnect      func()
	onDisconnect   func()
	onReconnecting func()
	onMessage      func(string)
	done           chan struct{}
}

func newWSClient() *wsClient {
	return &wsClient{
		done: make(chan struct{}),
	}
}

func (w *wsClient) SetCallbacks(onConnect, onDisconnect, onReconnecting func()) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.onConnect = onConnect
	w.onDisconnect = onDisconnect
	w.onReconnecting = onReconnecting
}

func (w *wsClient) SetMessageCallback(callback func(string)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.onMessage = callback
}

func (w *wsClient) Connect(addr string) error {
	w.mu.Lock()
	w.addr = addr
	w.reconnect = true
	w.mu.Unlock()

	return w.connect()
}

func (w *wsClient) connect() error {
	w.mu.RLock()
	addr := w.addr
	w.mu.RUnlock()

	conn, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		w.mu.Lock()
		w.connected = false
		w.mu.Unlock()
		return err
	}

	w.mu.Lock()
	w.conn = conn
	w.connected = true
	w.mu.Unlock()

	if w.onConnect != nil {
		go w.onConnect()
	}

	go w.readLoop()
	go w.writeLoop()

	return nil
}

func (w *wsClient) readLoop() {
	defer func() {
		w.mu.Lock()
		w.connected = false
		wasReconnect := w.reconnect
		w.conn = nil
		w.mu.Unlock()

		if wasReconnect && w.onReconnecting != nil {
			go w.onReconnecting()
			go w.reconnectLoop()
		}

		if w.onDisconnect != nil {
			go w.onDisconnect()
		}
	}()

	for {
		_, msg, err := w.conn.ReadMessage()
		if err != nil {
			return
		}

		w.mu.RLock()
		onMsg := w.onMessage
		w.mu.RUnlock()

		if onMsg != nil {
			go onMsg(string(msg))
		}
	}
}

func (w *wsClient) reconnectLoop() {
	for {
		time.Sleep(5 * time.Second)

		w.mu.RLock()
		shouldReconnect := w.reconnect && !w.connected
		w.mu.RUnlock()

		if !shouldReconnect {
			return
		}

		if err := w.connect(); err != nil {
			continue
		}
		return
	}
}

func (w *wsClient) writeLoop() {
	for {
		select {
		case <-w.done:
			return
		}
	}
}

func (w *wsClient) Send(message string) error {
	w.mu.RLock()
	conn := w.conn
	connected := w.connected
	w.mu.RUnlock()

	if !connected || conn == nil {
		return fmt.Errorf("not connected")
	}

	return conn.WriteMessage(websocket.TextMessage, []byte(message))
}

func (w *wsClient) Disconnect() {
	w.mu.Lock()
	w.reconnect = false

	if w.conn != nil {
		w.conn.Close()
		w.conn = nil
	}

	w.connected = false
	w.mu.Unlock()

	if w.onDisconnect != nil {
		go w.onDisconnect()
	}
}

func (w *wsClient) IsConnected() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.connected
}
