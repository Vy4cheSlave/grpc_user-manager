package middleware

import (
	"github.com/Vy4cheSlave/grpc_user-manager/internal/infrastructure/rest/dto/response"
	"github.com/Vy4cheSlave/grpc_user-manager/internal/jwt"
	"github.com/gofiber/fiber/v2"
)

func JWTAuthMiddleware(secret *[]byte) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		if token == "" {
			return response.ReturnResponse(
				c,
				fiber.StatusUnauthorized,
				response.WithError(response.ErrCodeMissingAuthHeader, "Authorization header is required"),
			)
		}

		claims, err := jwt.ValidateToken(token, *secret)
		if err != nil {
			return response.ReturnResponse(
				c,
				fiber.StatusUnauthorized,
				response.WithError(response.ErrCodeInvalidToken, "Invalid token"),
				response.WithMeta(err),
			)
		}

		c.Locals("userId", (*claims)["uid"].(string))

		return c.Next()
	}
}
