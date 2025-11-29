package middleware

import (
	"fmt"
	"go-multi-tenant/models"
	"go-multi-tenant/services" // ✅ Services import karna zaroori hai
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func PermissionMiddleware(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uint)
		tenantID := c.MustGet("tenantID").(uint)
		tenantDB := c.MustGet("tenantDB").(*gorm.DB)

		// ✅ Service Initialize karo
		cacheService := services.NewCacheService()

		// 1. Check Permissions in Cache first
		cacheKey := fmt.Sprintf("user_perms:%d:%d", tenantID, userID)
		var permissions []string

		// ✅ config.GetCacheStruct ki jagah cacheService.Get use karo
		err := cacheService.Get(cacheKey, &permissions)

		if err == nil {
			// Cache Hit!
			if hasPermission(permissions, requiredPermission) {
				c.Next()
				return
			} else {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions (cached)"})
				return
			}
		}

		// 2. Cache Miss: Fetch from DB (Load Roles & Permissions)
		var user models.User

		if err := tenantDB.Preload("Roles.Permissions").First(&user, userID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		// 3. Extract Permissions
		permissions = user.GetPermissions() // Helper method from models/user.go

		// 4. Save to Redis (10 Minutes Cache)
		// ✅ config.SetCacheStruct ki jagah cacheService.Set use karo
		_ = cacheService.Set(cacheKey, permissions, 10*time.Minute)

		// 5. Verify
		if hasPermission(permissions, requiredPermission) {
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		}
	}
}

// Helper function
func hasPermission(userPerms []string, required string) bool {
	for _, p := range userPerms {
		if p == required || p == "admin:full" { // "admin:full" is a super permission
			return true
		}
	}
	return false
}
