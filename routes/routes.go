package routes

import (
	"go-multi-tenant/config"
	"go-multi-tenant/handlers"
	"go-multi-tenant/middleware"
	"go-multi-tenant/repositories"
	"go-multi-tenant/services"
	"go-multi-tenant/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {

	userRepo := repositories.NewUserRepository(config.MasterDB)
	tenantRepo := repositories.NewTenantRepository(config.MasterDB)

	authService := services.NewAuthService(userRepo, tenantRepo)
	userService := services.NewUserService()
	tenantService := services.NewTenantService(tenantRepo, userRepo)

	// ✅ Services init (Empty constructor, no repo passed)
	moduleService := services.NewModuleService()
	permissionService := services.NewPermissionService()
	roleService := services.NewRoleService()

	queueService := services.NewQueueService(userService, 3)
	queueService.StartWorkers()

	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService, queueService)
	tenantHandler := handlers.NewTenantHandler(tenantService)

	public := router.Group("/api/v1")
	{
		public.POST("/login", authHandler.Login)
	}

	protected := router.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(authService))
	protected.Use(middleware.TenantDBMiddleware(authService))
	{
		auth := protected.Group("/auth")
		{
			auth.POST("/logout", authHandler.Logout)
		}

		users := protected.Group("/users")
		{
			users.POST("", middleware.PermissionMiddleware("user:create"), userHandler.CreateUser)
			users.GET("", middleware.PermissionMiddleware("user:list"), userHandler.ListUsers)
			users.GET("/:id", middleware.PermissionMiddleware("user:read"), userHandler.GetUser)

			users.PUT("/:id", middleware.PermissionMiddleware("user:update"), userHandler.UpdateUser)
			users.DELETE("/:id", middleware.PermissionMiddleware("user:delete"), userHandler.DeleteUser)
			users.GET("/roles/available", userHandler.GetAvailableRoles)
			users.POST("/:id/roles", middleware.PermissionMiddleware("user:update"), userHandler.AssignRole)
			users.DELETE("/:id/roles/:rid", middleware.PermissionMiddleware("user:update"), userHandler.RemoveRole)
		}

		self := protected.Group("/self")
		{
			self.GET("", userHandler.GetSelf)
		}

		tenants := protected.Group("/tenants")
		tenants.Use(middleware.PermissionMiddleware("tenant:create"))
		{
			tenants.POST("", tenantHandler.CreateTenant)
			tenants.GET("", tenantHandler.ListTenants)
		}
	}

	admin := router.Group("/api/v1/admin")
	admin.Use(middleware.AuthMiddleware(authService))
	admin.Use(middleware.TenantDBMiddleware(authService))
	{
		admin.GET("/cache-stats", middleware.PermissionMiddleware("user:read"), func(c *gin.Context) {
			cacheStats := map[string]interface{}{
				"cached_tenants": config.TenantManager.GetCachedTenantCount(),
				"memory_usage":   "optimized",
			}
			utils.SuccessResponse(c, http.StatusOK, "Cache statistics", cacheStats)
		})

		admin.POST("/clear-cache", middleware.PermissionMiddleware("tenant:create"), func(c *gin.Context) {
			config.TenantManager.ClearCache()
			utils.SuccessResponse(c, http.StatusOK, "Cache cleared successfully", nil)
		})

		RegisterModuleRoutes(admin, moduleService)
		RegisterPermissionRoutes(admin, permissionService)
		RegisterRoleRoutes(admin, roleService)
	}
}
