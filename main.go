package main

import (
	"go-multi-tenant/config"
	"go-multi-tenant/routes"
	"go-multi-tenant/services"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	if err := config.InitMasterDB(cfg); err != nil {
		log.Fatal("Failed to initialize master database:", err)
	}
	if err := config.InitRedis(cfg); err != nil {

		log.Printf("Warning: Redis Connection Failed: %v", err)
	}

	config.InitTenantManager(cfg)

	if err := config.TenantManager.CreateSharedDatabase(); err != nil {
		log.Printf("Warning: Failed to create shared database: %v", err)
	}

	saUser := os.Getenv("SUPERADMIN_USERNAME")
	saEmail := os.Getenv("SUPERADMIN_EMAIL")
	saPass := os.Getenv("SUPERADMIN_PASSWORD")
	if err := services.SeedMasterData(config.MasterDB, saUser, saEmail, saPass); err != nil {
		log.Printf("Warning: Failed to ensure master seed data: %v", err)
	}

	services.StartCronJob()
	log.Println("Server initialization completed successfully")
	router := gin.Default()
	routes.SetupRoutes(router)

	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := router.Run(cfg.ServerPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
