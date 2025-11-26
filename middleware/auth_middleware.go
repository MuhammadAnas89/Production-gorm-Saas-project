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

		// ✅ UPDATE: User ke sath TenantDB bhi receive kar rahe hain
		user, tenantDB, err := authService.ValidateAPIKey(apiKey)
		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid API key", err)
			c.Abort()
			return
		}

		// User context mein set karo
		c.Set("user", user)

		// ✅ UPDATE: Agar DB bhi mil gaya hai (Tenant User), usay bhi set karo
		// Taaki TenantDBMiddleware ko dobara connect na karna pare
		if tenantDB != nil {
			c.Set("tenantDB", tenantDB)
		}

		c.Next()
	}
}
