package internal

import (
	"hivem/pkg/api"
	"hivem/pkg/api/auth"
	"hivem/pkg/api/sse"
	"hivem/pkg/utils"
	"hivem/pkg/utils/logger"
	"net"

	"github.com/gofiber/fiber/v2"
)

func RunServer(config utils.Config) {

	app := fiber.New(fiber.Config{
		Prefork:       false,
		CaseSensitive: true,
		StrictRouting: true,
		ServerHeader:  "Hivem",
		AppName:       "Hivem",
	})

	app.Use(logger.RequestLogger())

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("config", config.Auth)
		return c.Next()
	})

	auth.ApplySecurityMiddlewares(app, config.Auth)

	backendRoute := app.Group("/api/backend")
	backendRoute.Post("/sign", api.Sign)

	apiRoute := app.Group("/api/frontend")
	apiRoute.Get("/set", api.Set)
	apiRoute.Post("/set", api.Set)
	apiRoute.Get("/get", api.Get)
	apiRoute.Get("/sse", sse.SSE)

	ln, err := net.Listen("tcp", config.Host+":"+config.Port)
	if err != nil {
		panic(err)
	}
	app.Listener(ln)
}
