package routes

import (
	"go-multi-tenant/handlers"
	"go-multi-tenant/middleware"
	"go-multi-tenant/services"

	"github.com/gin-gonic/gin"
)

func RegisterRoleRoutes(rg *gin.RouterGroup, svc *services.RoleService) {
	h := handlers.NewRoleHandler(svc)
	roles := rg.Group("/roles")

	{
		roles.POST("", middleware.PermissionMiddleware("role:create"), h.CreateRole)
		roles.GET("", middleware.PermissionMiddleware("role:read"), h.ListRoles)
		roles.GET(":id", middleware.PermissionMiddleware("role:read"), h.GetRole)
		roles.PUT(":id", middleware.PermissionMiddleware("role:update"), h.UpdateRole)
		roles.DELETE(":id", middleware.PermissionMiddleware("role:delete"), h.DeleteRole)
		roles.POST(":id/permissions", middleware.PermissionMiddleware("role:update"), h.AttachPermissions)
		roles.DELETE(":id/permissions/:pid", middleware.PermissionMiddleware("role:update"), h.DetachPermission)
	}
}
