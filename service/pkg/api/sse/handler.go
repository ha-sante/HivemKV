package sse

import (
	"bufio"
	"fmt"
	"hivem/pkg/api/auth"
	"hivem/pkg/store"
	"hivem/pkg/utils/logger"
	"time"

	"github.com/gofiber/fiber/v2"
)

var Hub *SSEHub

func init() {
	Hub = NewHub()
}

func SSE(c *fiber.Ctx) error {
	identityValue := c.Locals("identity")
	if identityValue == nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	identity := identityValue.(string)
	role := c.Locals("role").(string)

	ch := Hub.Subscribe(identity)

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer Hub.Unsubscribe(identity, ch)

		initialMessages, err := getInitialState(identity, role)
		if err == nil {
			for _, msg := range initialMessages {
				fmt.Fprintf(w, "data: %s\n\n", msg.ToJSON())
				w.Flush()
			}
		}

		keepalive := time.NewTicker(15 * time.Second)
		defer keepalive.Stop()

		for {
			select {
			case msg, ok := <-ch:
				if !ok {
					return
				}
				if _, allowed := auth.KeyScopeAccessAllowed(msg.Key, role, auth.OpRead); !allowed {
					continue
				}

				fmt.Fprintf(w, "data: %s\n\n", msg.ToJSON())
				if err := w.Flush(); err != nil {
					logger.Info("sse_disconnect", "identity", identity, "reason", "write_failed")
					return
				}
			case <-keepalive.C:
				fmt.Fprintf(w, ": keepalive\n\n")
				if err := w.Flush(); err != nil {
					logger.Info("sse_disconnect", "identity", identity, "reason", "keepalive_failed")
					return
				}
			}
		}
	})

	return nil
}

func getInitialState(identity, role string) ([]*Message, error) {
	messages := []*Message{}
	identityPrefix := fmt.Sprintf("%s:", identity)

	iter, err := store.Iter()
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	for iter.SeekGE([]byte(identityPrefix)); iter.Valid(); iter.Next() {
		key := string(iter.Key())
		value := string(iter.Value())

		if len(key) <= len(identityPrefix) {
			break
		}
		valuePostFix := key[:len(identityPrefix)]
		if valuePostFix != identityPrefix {
			break
		}

		valueKey := key[len(identityPrefix):]
		if _, allowed := auth.KeyScopeAccessAllowed(valueKey, role, auth.OpRead); !allowed {
			continue
		}
		messages = append(messages, &Message{
			Identity:  identity,
			Key:       valueKey,
			Value:     value,
			Operation: "set",
			Timestamp: 0,
		})
	}

	// log inital sends
	if len(messages) > 0 {
		for _, m := range messages {
			logger.Info("sse_send_initial", "identity", identity, "key", m.Key)
		}
	} else {
		logger.Info("sse_send_initial_empty", "identity", identity)
	}

	return messages, nil
}
