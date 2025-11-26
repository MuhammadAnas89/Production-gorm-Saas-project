package middleware

import (
	"go-multi-tenant/config"
	"go-multi-tenant/models"
	"go-multi-tenant/repositories"
	"go-multi-tenant/services"
	"go-multi-tenant/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func TenantDBMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Agar AuthMiddleware ne pehle hi DB set kar diya hai, wahi use karo
		if _, exists := c.Get("tenantDB"); exists {
			c.Next()
			return
		}

		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.Next()
			return
		}

		userInterface, exists := c.Get("user")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not found in context", nil)
			c.Abort()
			return
		}

		user := userInterface.(*models.User)

		// ✅ Super Administrator always uses Master DB
		if user.HasRole("Super Administrator") {
			c.Set("tenantDB", config.MasterDB)
			c.Next()
			return
		}

		// 2. Tenant Info fetch karo
		tenantRepo := repositories.NewTenantRepository(config.MasterDB)
		tenant, err := tenantRepo.GetByID(user.TenantID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get tenant information", err)
			c.Abort()
			return
		}

		if !tenant.IsActive {
			utils.ErrorResponse(c, http.StatusForbidden, "Tenant account is suspended", nil)
			c.Abort()
			return
		}

		// 3. Raw Database Connection lo
		rawDB, err := config.TenantManager.GetTenantDB(tenant)
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to connect to tenant database", err)
			c.Abort()
			return
		}

		// ✅ CRITICAL FIX FOR SHARED DB: Scope Injection
		// Agar Shared DB hai, to hum us par TenantID ka filter permanent laga denge is request ke liye.
		if tenant.DatabaseType == models.SharedDB {
			scopedDB := rawDB.Where("tenant_id = ?", tenant.ID)
			c.Set("tenantDB", scopedDB)
		} else {
			// Dedicated DB mein filter ki zaroorat nahi
			c.Set("tenantDB", rawDB)
		}

		c.Next()
	}
}
