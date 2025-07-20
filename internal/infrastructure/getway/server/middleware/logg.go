package middleware

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log/slog"
	"time"
)

func LoggingMiddleware(log *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		log.Info(fmt.Sprintf("[START] %s %s | IP: %s",
			c.Method(),
			c.Path(),
			c.IP(),
		))

		err := c.Next()
		logErr := err
		// хотелось бы фидбек на то насколько это антипаттерн или наоборот,
		// потому что по логике эта хуйня нихуя не явная, и это по сути ContextWithValue(в методе так написано)
		if errHandler, ok := c.Locals("handlerError").(error); ok {
			logErr = errHandler
		}

		duration := time.Since(start)
		status := c.Response().StatusCode() // в падлу исправлять, но по сути статус может поменяться в дальнейшем и в логах будет некорректная информация

		if logErr != nil {
			log.Error(fmt.Sprintf("[ERROR] %s %s | Status: %d | Duration: %v",
				c.Method(),
				c.Path(),
				status,
				duration,
			),
				slog.String("error", logErr.Error()),
			)
		} else {
			log.Info(fmt.Sprintf("[END] %s %s | Status: %d | Duration: %v",
				c.Method(),
				c.Path(),
				status,
				duration,
			))
		}

		return err
	}
}
