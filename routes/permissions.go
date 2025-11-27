package routes

import (
	"go-multi-tenant/handlers"
	"go-multi-tenant/middleware"
	"go-multi-tenant/services"

	"github.com/gin-gonic/gin"
)

func RegisterPermissionRoutes(rg *gin.RouterGroup, svc *services.PermissionService) {
	h := handlers.NewPermissionHandler(svc)
	perms := rg.Group("/permissions")

	{
		perms.GET("", middleware.PermissionMiddleware("role:read"), h.ListPermissions)
		perms.GET(":id", middleware.PermissionMiddleware("role:read"), h.GetPermission)
		perms.POST("", middleware.PermissionMiddleware("permission:manage"), h.CreatePermission)
		perms.PUT(":id", middleware.PermissionMiddleware("permission:manage"), h.UpdatePermission)
		perms.DELETE(":id", middleware.PermissionMiddleware("permission:manage"), h.DeletePermission)
	}
}
