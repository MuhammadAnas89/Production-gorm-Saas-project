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
	perms.Use(middleware.PermissionMiddleware("admin:full"))
	{
		perms.POST("", h.CreatePermission)
		perms.GET("", h.ListPermissions)
		perms.GET(":id", h.GetPermission)
		perms.PUT(":id", h.UpdatePermission)
		perms.DELETE(":id", h.DeletePermission)
	}
}
