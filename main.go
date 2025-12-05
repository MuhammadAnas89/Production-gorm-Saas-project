package main

import (
	"go-multi-tenant/config"
	"go-multi-tenant/routes"
	"go-multi-tenant/services"
	"go-multi-tenant/utils" // ✅ Add this import
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	// 1. Load Config
	cfg := config.Load()

	// 2. Init Utils (JWT Secret) ✅ MOST IMPORTANT MISSING PART
	// Agar ye nahi karoge to token generate/validate nahi hoga
	utils.InitJWT(cfg.JWTSecret)

	// 3. Init Master DB
	if err := config.InitMasterDB(cfg); err != nil {
		log.Fatal("❌ Failed to initialize master database:", err)
	}

	// 4. Init Redis
	if err := config.InitRedis(cfg); err != nil {
		log.Println("⚠️  Warning: Redis connection failed. Cache will not work.", err)
	}

	// 5. Init Tenant Manager
	config.InitTenantManager(cfg)

	// 6. Create Shared Database (Physical DB creation if not exists)
	if err := config.TenantManager.CreateSharedDatabase(); err != nil {
		log.Printf("⚠️  Warning: Failed to create shared database: %v", err)
	}

	// 7. Seed Master Data (Super Admin & Plans)
	saUser := os.Getenv("SUPERADMIN_USERNAME")
	saEmail := os.Getenv("SUPERADMIN_EMAIL")
	saPass := os.Getenv("SUPERADMIN_PASSWORD")

	log.Println("Seeding master data...")
	if err := services.SeedMasterData(config.MasterDB, saUser, saEmail, saPass); err != nil {
		log.Printf("Warning: Failed to ensure master seed data: %v", err)
	} else {
		log.Println("✅ Master data seeded successfully")
	}

	// 8. Setup Router
	router := gin.Default()

	// CORS Setup
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Production mein isay specific domain kar dena
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 9. Register Routes
	routes.SetupRoutes(router)

	// 10. Start Server
	serverPort := cfg.ServerPort
	if serverPort == "" {
		serverPort = ":8080"
	}

	log.Printf("🚀 Server starting on port %s", serverPort)
	if err := router.Run(serverPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
