package api

import (
	"fmt"
	"hivem/pkg/api/auth"
	"hivem/pkg/api/sse"
	"hivem/pkg/store"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Set(c *fiber.Ctx) error {
	var key, value string

	if c.Method() == "POST" {
		var body struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "invalid JSON body",
			})
		}
		key = body.Key
		value = body.Value
	} else {
		key = c.Query("key")
		value = c.Query("value")
	}

	if key == "" || value == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "key and value are required",
		})
	}

	identity := c.Locals("identity").(string)
	role := c.Locals("role").(string)
	if _, allowed := auth.KeyScopeAccessAllowed(key, role, auth.OpWrite); !allowed {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "forbidden",
		})
	}
	scopedKey := fmt.Sprintf("%s:%s", identity, key)

	if err := store.SaveKey(scopedKey, []byte(value)); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to save key",
		})
	}

	sse.Hub.Broadcast(identity, &sse.Message{
		Identity:  identity,
		Key:       key,
		Value:     value,
		Operation: "set",
		Timestamp: time.Now().Unix(),
	})

	return c.JSON(fiber.Map{
		"status": "ok",
		"key":    key,
	})
}

func Get(c *fiber.Ctx) error {
	key := c.Query("key")

	if key == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "key is required",
		})
	}

	identity := c.Locals("identity").(string)
	role := c.Locals("role").(string)
	if _, allowed := auth.KeyScopeAccessAllowed(key, role, auth.OpRead); !allowed {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "forbidden",
		})
	}
	scopedKey := fmt.Sprintf("%s:%s", identity, key)

	value, err := store.GetKey(scopedKey)
	if err != nil {
		if store.IsNotFound(err) {
			return c.Status(404).JSON(fiber.Map{
				"error": "key not found",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to get key",
		})
	}
	return c.JSON(fiber.Map{
		"key":   key,
		"value": value,
	})
}

func Sign(c *fiber.Ctx) error {
	identity := c.Locals("identity").(string)

	cfg := c.Locals("config").(auth.AuthConfig)
	if len(cfg.SigningKeys) == 0 {
		return c.Status(500).JSON(fiber.Map{"error": "signing keys not configured"})
	}

	signingKey := ""
	for key := range cfg.SigningKeys {
		signingKey = key
		break
	}

	if signingKey == "" {
		return c.Status(500).JSON(fiber.Map{"error": "no signing keys available"})
	}

	// uses premade middleware util
	signature := auth.SignIdentity(identity, signingKey)

	return c.JSON(fiber.Map{
		"identity":  identity,
		"signature": signature,
	})
}
