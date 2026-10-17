package sse

import (
	"encoding/json"
	"sync"

	"hivem/pkg/utils/logger"
)

type Message struct {
	Identity  string `json:"identity"`
	Key       string `json:"key"`
	Value     string `json:"value,omitempty"`
	Operation string `json:"operation"`
	Timestamp int64  `json:"timestamp"`
}

func (m *Message) ToJSON() string {
	bytes, _ := json.Marshal(m)
	return string(bytes)
}

type SSEHub struct {
	subscribers map[string][]chan *Message
	mu          sync.RWMutex
}

func NewHub() *SSEHub {
	return &SSEHub{
		subscribers: make(map[string][]chan *Message),
	}
}

func (h *SSEHub) Subscribe(identity string) chan *Message {
	h.mu.Lock()
	defer h.mu.Unlock()

	ch := make(chan *Message, 100)
	h.subscribers[identity] = append(h.subscribers[identity], ch)

	logger.Debug("sse_subscribe", "identity", identity, "total_subscribers", len(h.subscribers[identity]))

	return ch
}

func (h *SSEHub) Unsubscribe(identity string, ch chan *Message) {
	h.mu.Lock()
	defer h.mu.Unlock()

	channels := h.subscribers[identity]
	for i, c := range channels {
		if c == ch {
			h.subscribers[identity] = append(channels[:i], channels[i+1:]...)
			close(ch)
			logger.Debug("sse_unsubscribe", "identity", identity, "remaining_subscribers", len(h.subscribers[identity]))
			break
		}
	}

	if len(h.subscribers[identity]) == 0 {
		delete(h.subscribers, identity)
		logger.Debug("sse_identity_removed", "identity", identity)
	}
}

func (h *SSEHub) Broadcast(identity string, msg *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	channels := h.subscribers[identity]
	count := len(channels)

	if count == 0 {
		logger.Debug("sse_broadcast_sync_skipped", "identity", identity, "reason", "no_subscribers")
		return
	}

	logger.Debug("sse_broadcast", "identity", identity, "subscribers", count, "key", msg.Key, "operation", msg.Operation)

	for _, ch := range channels {
		select {
		case ch <- msg:
		default:
		}
	}
}
