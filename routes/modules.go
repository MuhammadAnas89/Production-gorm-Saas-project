package routes

import (
	"go-multi-tenant/handlers"
	"go-multi-tenant/middleware"
	"go-multi-tenant/services"

	"github.com/gin-gonic/gin"
)

func RegisterModuleRoutes(rg *gin.RouterGroup, svc *services.ModuleService) {
	h := handlers.NewModuleHandler(svc)
	mods := rg.Group("/modules")

	{
		mods.GET("", middleware.PermissionMiddleware("role:read"), h.ListModules)
		mods.GET(":id", middleware.PermissionMiddleware("role:read"), h.GetModule)
		mods.POST("", middleware.PermissionMiddleware("system:config"), h.CreateModule)
		mods.PUT(":id", middleware.PermissionMiddleware("system:config"), h.UpdateModule)
		mods.DELETE(":id", middleware.PermissionMiddleware("system:config"), h.DeleteModule)
	}
}
