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

	// Initialize master database
	if err := config.InitMasterDB(cfg); err != nil {
		log.Fatal("Failed to initialize master database:", err)
	}

	// Initialize tenant manager
	config.InitTenantManager(cfg)

	// Create shared database
	if err := config.TenantManager.CreateSharedDatabase(); err != nil {
		log.Printf("Warning: Failed to create shared database: %v", err)
	}

	// Ensure minimal master data (super-admin + core permissions/role).
	saUser := os.Getenv("SUPERADMIN_USERNAME")
	saEmail := os.Getenv("SUPERADMIN_EMAIL")
	saPass := os.Getenv("SUPERADMIN_PASSWORD")
	if err := services.SeedMasterData(config.MasterDB, saUser, saEmail, saPass); err != nil {
		log.Printf("Warning: Failed to ensure master seed data: %v", err)
	}

	log.Println("Server initialization completed successfully")

	router := gin.Default()
	routes.SetupRoutes(router)

	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := router.Run(cfg.ServerPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
