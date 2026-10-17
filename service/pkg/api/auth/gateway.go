package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"strings"

	"hivem/pkg/utils/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func ApplySecurityMiddlewares(app *fiber.App, cfg AuthConfig) {
	if len(cfg.SigningKeys) == 0 {
		panic("HIVEM_SIGNING_KEYS must be set when frontend signature verification is enabled")
	}

	app.Use(healthcheck.New(healthcheck.Config{
		LivenessProbe:  func(c *fiber.Ctx) bool { return true },
		ReadinessProbe: func(c *fiber.Ctx) bool { return true },
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Join(cfg.AllowedOrigins, ","),
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Authorization,Content-Type,X-Identity,X-Signature",
		ExposeHeaders:    "X-Role-Name",
		AllowCredentials: false,
	}))

	if len(cfg.IPWhitelist) > 0 {
		app.Use(createIPWhitelistMiddleware(cfg.IPWhitelist))
	} else {
		logger.Debug("skipped ip whitelist", "ip whitelist", cfg.IPWhitelist)
	}

	app.Use(func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" {
			c.Locals("auth", "bearer token not present")
		}
		return c.Next()
	})

	app.Use(keyauth.New(keyauth.Config{
		KeyLookup:  "header:Authorization",
		AuthScheme: "Bearer",
		Validator:  createKeyValidator(cfg),
	}))

	app.Use(createRoleAccessMiddleware(cfg))

	app.Use(limiter.New(limiter.Config{
		Max:        cfg.RateLimitMax,
		Expiration: cfg.RateLimitDuration,
	}))
}

func createIPWhitelistMiddleware(whitelist []string) fiber.Handler {
	whitelistMap := make(map[string]bool)
	for _, ip := range whitelist {
		whitelistMap[ip] = true
	}

	return func(c *fiber.Ctx) error {
		ip := c.IP()
		if !whitelistMap[ip] {
			c.Locals("auth", "ip not in whitelist")
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
		}
		return c.Next()
	}
}

func createKeyValidator(cfg AuthConfig) func(c *fiber.Ctx, key string) (bool, error) {
	return func(c *fiber.Ctx, key string) (bool, error) {
		if cfg.FrontendKeys[key] {
			c.Locals("role", "frontend")
			return true, nil
		}
		if cfg.BackendKeys[key] {
			c.Locals("role", "backend")
			return true, nil
		}
		c.Locals("auth", "invalid api key")
		return false, keyauth.ErrMissingOrMalformedAPIKey
	}
}

func createRoleAccessMiddleware(cfg AuthConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role")
		path := c.Path()

		switch role {
		case "frontend":
			if !fiber.RoutePatternMatch(path, "/api/frontend/*") {
				c.Locals("auth", "forbidden")
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
			}

			identity := c.Get("X-Identity")
			signature := c.Get("X-Signature")

			if identity == "" {
				c.Locals("auth", "missing X-Identity header")
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing signature"})
			}
			if signature == "" {
				c.Locals("auth", "missing X-Signature header")
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing signature"})
			}

			if !verifySignature(identity, signature, cfg.SigningKeys) {
				log.Println("Verifying signature for identity:", identity, "signature:", signature)
				c.Locals("auth", "invalid signature")
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid signature"})
			}

			c.Locals("identity", identity)
			c.Set("X-Role-Name", "frontend")
			return c.Next()
		case "backend":
			backendMatch := fiber.RoutePatternMatch(path, "/api/backend/*") || fiber.RoutePatternMatch(path, "/api/frontend/*")
			if !backendMatch {
				c.Locals("auth", "forbidden")
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
			}

			identity := c.Get("X-Identity")
			if identity == "" {
				c.Locals("auth", "missing X-Identity header")
				return c.Status(401).JSON(fiber.Map{"error": "X-Identity header required for backend"})
			}

			c.Locals("identity", identity)
			c.Set("X-Role-Name", "backend")
			return c.Next()
		default:
			c.Locals("auth", "no role")
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
		}
	}
}

func verifySignature(identity, signature string, signingKeys map[string]bool) bool {
	for key := range signingKeys {
		expected := SignIdentity(identity, key)
		if hmac.Equal([]byte(expected), []byte(signature)) {
			return true
		}
	}
	return false
}

func SignIdentity(identity, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(identity))
	return hex.EncodeToString(mac.Sum(nil))
}
