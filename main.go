package main

import (
	"go-multi-tenant/config"
	"go-multi-tenant/routes"
	"go-multi-tenant/services"
	"go-multi-tenant/utils"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	cfg := config.Load()
	utils.InitJWT(cfg.JWTSecret)

	if err := config.InitMasterDB(cfg); err != nil {
		log.Fatal("Failed to initialize master database:", err)
	}

	if err := config.InitRedis(cfg); err != nil {
		log.Println("Warning: Redis connection failed. Cache will not work.", err)
	}
	config.InitTenantManager(cfg)
	if err := config.TenantManager.CreateSharedDatabase(); err != nil {
		log.Printf("Warning: Failed to create shared database: %v", err)
	}

	saUser := os.Getenv("SUPERADMIN_USERNAME")
	saEmail := os.Getenv("SUPERADMIN_EMAIL")
	saPass := os.Getenv("SUPERADMIN_PASSWORD")

	log.Println("Seeding master data...")
	if err := services.SeedMasterData(config.MasterDB, saUser, saEmail, saPass); err != nil {
		log.Printf("Warning: Failed to ensure master seed data: %v", err)
	} else {
		log.Println("Master data seeded successfully")
	}

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	routes.SetupRoutes(router)

	serverPort := cfg.ServerPort
	if serverPort == "" {
		serverPort = ":8080"
	}

	log.Printf("Server starting on port %s", serverPort)
	if err := router.Run(serverPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
