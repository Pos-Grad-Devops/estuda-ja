package middleware

import (
	"strings"

	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/auth"
	"github.com/Pos-Grad-Devops/estuda-ja/backend/internal/models"
	"github.com/gofiber/fiber/v2"
)

const UserContextKey = "user"

func Authenticate(tokens *auth.TokenService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "não autenticado"})
		}

		claims, err := tokens.Parse(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "token inválido"})
		}

		c.Locals(UserContextKey, claims)
		return c.Next()
	}
}

func RequireRoles(roles ...models.Role) fiber.Handler {
	allowed := make(map[models.Role]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *fiber.Ctx) error {
		claims, ok := c.Locals(UserContextKey).(*auth.Claims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "não autenticado"})
		}
		if _, ok := allowed[claims.Role]; !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "sem permissão"})
		}
		return c.Next()
	}
}

func ClaimsFromContext(c *fiber.Ctx) (*auth.Claims, bool) {
	claims, ok := c.Locals(UserContextKey).(*auth.Claims)
	return claims, ok
}
