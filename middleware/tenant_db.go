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
			// Cache Miss: Fetch from Master DB
			// Preload Plan taaki limits bhi cache ho jayen
			if dbErr := config.MasterDB.Preload("Plan").First(&tenant, tenantID).Error; dbErr != nil {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Tenant not found or inactive"})
				return
			}
			// Cache Set (30 Minutes Expiry)
			cacheService := services.NewCacheService()
			_ = cacheService.Set(cacheKey, tenant, 30*time.Minute)
		}

		if !tenant.IsActive {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Tenant account is suspended"})
			return
		}

		// 3. Get Tenant Database Connection
		tenantDB, err := config.TenantManager.GetTenantDB(&tenant)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to tenant database"})
			return
		}

		// Shared DB Scope Logic
		if tenant.DatabaseType == models.SharedDB {
			// Agar shared DB hai, to har query mein `WHERE tenant_id = ?` lagana zaroori hai
			// GORM ka scoped session use karenge
			scopedDB := tenantDB.Where("tenant_id = ?", tenant.ID)
			c.Set("tenantDB", scopedDB)
		} else {
			c.Set("tenantDB", tenantDB)
		}

		// Tenant Info context mein rakho (Limits check karne ke liye kaam ayegi)
		c.Set("tenant", &tenant)

		c.Next()
	}
}
