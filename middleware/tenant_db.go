package middleware

import (
	"fmt"
	"go-multi-tenant/config"
	"go-multi-tenant/models"
	"go-multi-tenant/repositories"
	"go-multi-tenant/services"
	"go-multi-tenant/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func TenantDBMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		if user.HasRole("Super Administrator") {
			c.Set("tenantDB", config.MasterDB)
			c.Next()
			return
		}

		var tenant models.Tenant
		cacheKey := fmt.Sprintf("tenant_info:%d", user.TenantID)

		err := config.GetCacheStruct(cacheKey, &tenant)

		if err == nil {

		} else {

			tenantRepo := repositories.NewTenantRepository(config.MasterDB)
			dbTenant, dbErr := tenantRepo.GetByID(user.TenantID)
			if dbErr != nil {
				utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get tenant information", dbErr)
				c.Abort()
				return
			}
			tenant = *dbTenant

			_ = config.SetCacheStruct(cacheKey, tenant, 30*time.Minute)
		}

		if !tenant.IsActive {
			utils.ErrorResponse(c, http.StatusForbidden, "Tenant account is suspended", nil)
			c.Abort()
			return
		}

		rawDB, err := config.TenantManager.GetTenantDB(&tenant)
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to connect to tenant database", err)
			c.Abort()
			return
		}

		if tenant.DatabaseType == models.SharedDB {
			scopedDB := rawDB.Where("tenant_id = ?", tenant.ID)
			c.Set("tenantDB", scopedDB)
		} else {
			c.Set("tenantDB", rawDB)
		}

		c.Next()
	}
}
