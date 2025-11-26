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
	roles.Use(middleware.PermissionMiddleware("admin:full"))
	{
		roles.POST("", h.CreateRole)
		roles.GET("", h.ListRoles)
		roles.GET(":id", h.GetRole)
		roles.PUT(":id", h.UpdateRole)
		roles.DELETE(":id", h.DeleteRole)
		roles.POST(":id/permissions", h.AttachPermissions)
		roles.DELETE(":id/permissions/:pid", h.DetachPermission)
	}
}
