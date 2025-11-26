package middleware

import (
	"go-multi-tenant/models"
	"go-multi-tenant/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func PermissionMiddleware(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userInterface, exists := c.Get("user")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not found in context", nil)
			c.Abort()
			return
		}

		user := userInterface.(*models.User)

		if !user.HasPermission(requiredPermission) {
			utils.ErrorResponse(c, http.StatusForbidden, "Insufficient permissions", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}
