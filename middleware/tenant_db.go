package middleware

import (
	"fmt"
	"go-multi-tenant/config"
	"go-multi-tenant/models"
	"go-multi-tenant/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func TenantDBMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get TenantID from previous middleware (AuthMiddleware)
		tenantIDInterface, exists := c.Get("tenantID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found in context"})
			return
		}
		tenantID := tenantIDInterface.(uint)

		// 2. Fetch Tenant Info (Redis First -> Then Master DB)
		var tenant models.Tenant
		cacheKey := fmt.Sprintf("tenant_info:%d", tenantID)

		// Try Cache
		cacheService := services.NewCacheService()
		err := cacheService.Get(cacheKey, &tenant)
		if err != nil {
			if dbErr := config.MasterDB.Preload("Plan").First(&tenant, tenantID).Error; dbErr != nil {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Tenant not found or inactive"})
				return
			}
			// Cache Set (30 Minutes Expiry)
			_ = cacheService.Set(cacheKey, tenant, 30*time.Minute)
		}

		if !tenant.IsActive {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Tenant account is suspended"})
			return
		}
		// 3. Connect to Tenant DB
		tenantDB, err := config.TenantManager.GetTenantDB(&tenant)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to tenant database"})
			return
		}
		c.Set("tenantDB", tenantDB)

		c.Set("currentTenant", &tenant)

		c.Next()
	}
}
