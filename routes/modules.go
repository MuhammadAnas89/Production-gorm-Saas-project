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
	mods.Use(middleware.PermissionMiddleware("admin:full"))
	{
		mods.POST("", h.CreateModule)
		mods.GET("", h.ListModules)
		mods.GET(":id", h.GetModule)
		mods.PUT(":id", h.UpdateModule)
		mods.DELETE(":id", h.DeleteModule)
	}
}
