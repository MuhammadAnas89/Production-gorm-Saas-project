package routes

import (
	"go-multi-tenant/config"
	"go-multi-tenant/handlers"
	"go-multi-tenant/middleware"
	"go-multi-tenant/repositories"
	"go-multi-tenant/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {

	tenantRepo := repositories.NewTenantRepository(config.MasterDB)
	moduleService := services.NewModuleService()
	permissionService := services.NewPermissionService()

	authService := services.NewAuthService(tenantRepo)
	tenantService := services.NewTenantService(tenantRepo)
	userService := services.NewUserService()
	catalogService := services.NewCatalogService()
	inventoryService := services.NewInventoryService()
	roleService := services.NewRoleService()
	purchaseService := services.NewPurchaseService()
	purchaseHandler := handlers.NewPurchaseHandler(purchaseService)

	authHandler := handlers.NewAuthHandler(authService)
	tenantHandler := handlers.NewTenantHandler(tenantService)
	userHandler := handlers.NewUserHandler(userService)
	catalogHandler := handlers.NewCatalogHandler(catalogService)
	inventoryHandler := handlers.NewInventoryHandler(inventoryService)
	roleHandler := handlers.NewRoleHandler(roleService)
	moduleHandler := handlers.NewModuleHandler(moduleService)
	permHandler := handlers.NewPermissionHandler(permissionService)

	api := router.Group("/api/v1")

	api.POST("/login", authHandler.Login)

	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware())
	protected.Use(middleware.TenantDBMiddleware())

	users := protected.Group("/users")
	{
		users.POST("", middleware.PermissionMiddleware("user:create"), userHandler.CreateUser)
		users.GET("", middleware.PermissionMiddleware("user:read"), userHandler.ListUsers)
		users.GET("/:id", middleware.PermissionMiddleware("user:read"), userHandler.GetUser)
		users.PUT("/:id", middleware.PermissionMiddleware("user:update"), userHandler.UpdateUser)
		users.DELETE("/:id", middleware.PermissionMiddleware("user:delete"), userHandler.DeleteUser)
	}

	products := protected.Group("/products")
	{

		products.POST("", middleware.PermissionMiddleware("product:create"), catalogHandler.CreateProduct)
		products.GET("", middleware.PermissionMiddleware("product:read"), catalogHandler.ListProducts)
	}

	// === CATEGORIES ===
	categories := protected.Group("/categories")
	{

		categories.POST("", middleware.PermissionMiddleware("category:create"), catalogHandler.CreateCategory)
		categories.GET("", middleware.PermissionMiddleware("category:read"), catalogHandler.ListCategories)
	}

	// === INVENTORY & STOCK ===
	inventory := protected.Group("/inventory")
	{
		inventory.PUT("/stock", middleware.PermissionMiddleware("inventory:update"), inventoryHandler.UpdateStock)
		inventory.GET("/alerts", middleware.PermissionMiddleware("inventory:read"), inventoryHandler.GetLowStockAlerts)

	}

	roles := protected.Group("/roles")
	{

		roles.POST("", middleware.PermissionMiddleware("role:manage"), roleHandler.CreateRole)
		roles.GET("", middleware.PermissionMiddleware("user:read"), roleHandler.ListRoles) // User read wala bhi dekh sake
		roles.GET("/:id", middleware.PermissionMiddleware("role:manage"), roleHandler.GetRole)
		roles.PUT("/:id/permissions", middleware.PermissionMiddleware("role:manage"), roleHandler.UpdatePermissions)
	}

	protected.POST("/tenants", middleware.PermissionMiddleware("tenant:create"), tenantHandler.CreateTenant)
	protected.GET("/tenants", middleware.PermissionMiddleware("tenant:manage"), tenantHandler.ListTenants)

	modules := protected.Group("/modules")
	{

		modules.POST("", middleware.PermissionMiddleware("system:manage"), moduleHandler.Create)
		modules.GET("", middleware.PermissionMiddleware("system:manage"), moduleHandler.List)
		modules.PUT("", middleware.PermissionMiddleware("system:manage"), moduleHandler.Update)
		modules.DELETE("/:id", middleware.PermissionMiddleware("system:manage"), moduleHandler.Delete)
	}

	perms := protected.Group("/permissions")
	{

		perms.POST("", middleware.PermissionMiddleware("system:manage"), permHandler.Create)
		perms.GET("", middleware.PermissionMiddleware("system:manage"), permHandler.List)
		perms.DELETE("/:id", middleware.PermissionMiddleware("system:manage"), permHandler.Delete)
	}

	purchase := protected.Group("/purchase-orders")
	{

		purchase.POST("", middleware.PermissionMiddleware("purchase:create"), purchaseHandler.Create)
		purchase.PUT("/:id", middleware.PermissionMiddleware("purchase:update"), purchaseHandler.UpdateRequest)
		purchase.POST("/:id/action", middleware.PermissionMiddleware("purchase:action"), purchaseHandler.PurchaserAction)
		purchase.POST("/:id/receive", middleware.PermissionMiddleware("purchase:receive"), purchaseHandler.Receive)
		purchase.GET("", middleware.PermissionMiddleware("purchase:view"), purchaseHandler.List)
	}

}
