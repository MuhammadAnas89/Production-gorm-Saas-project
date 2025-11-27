package middleware

import (
	"go-multi-tenant/services"
	"go-multi-tenant/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "API key required", nil)
			c.Abort()
			return
		}

		user, tenantDB, err := authService.ValidateAPIKey(apiKey)
		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid API key", err)
			c.Abort()
			return
		}

		c.Set("user", user)

		if tenantDB != nil {
			c.Set("tenantDB", tenantDB)
		}

		c.Next()
	}
}
