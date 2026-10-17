package logger

import (
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var Logger *slog.Logger

func InitLogger(level string) {
	var slogLevel slog.Level
	switch level {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: slogLevel,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	Logger = slog.New(handler)
}

func Info(msg string, args ...any) {
	Logger.Info(msg, args...)
}

func Debug(msg string, args ...any) {
	Logger.Debug(msg, args...)
}

func Warn(msg string, args ...any) {
	Logger.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	Logger.Error(msg, args...)
}

func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// start
		requestID := uuid.New().String()
		c.Locals("request_id", requestID)

		args := []any{
			"id", requestID,
			"method", c.Method(),
			"path", c.Path(),
		}
		Logger.Info("startr", args...)

		// continue
		err := c.Next()

		// end
		resultArgs := []any{
			"id", requestID,
			"status", c.Response().StatusCode(),
		}

		if authDetail := c.Locals("auth"); authDetail != nil {
			resultArgs = append(resultArgs, "auth", authDetail)
		}

		Logger.Info("endr", resultArgs...)
		return err
	}
}
