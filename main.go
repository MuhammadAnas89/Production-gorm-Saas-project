package main

import (
	"go-multi-tenant/config"
	"go-multi-tenant/routes"
	"go-multi-tenant/services"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load Configuration
	cfg := config.Load()

	// 2. Initialize Master Database
	if err := config.InitMasterDB(cfg); err != nil {
		log.Fatal("❌ Failed to initialize master database:", err)
	}

	// 3. Initialize Redis (Optional but recommended)
	if err := config.InitRedis(cfg); err != nil {
		log.Println("⚠️  Warning: Redis connection failed. Cache will not work.", err)
	}

	// 4. Initialize Tenant Manager
	config.InitTenantManager(cfg)

	// 5. Create Shared Database (Agar exist nahi karta)
	if err := config.TenantManager.CreateSharedDatabase(); err != nil {
		log.Printf("⚠️  Warning: Failed to create shared database: %v", err)
	}

	// 6. SEED MASTER DATA (Super Admin, Plans, Modules)
	// Ye bohot zaroori step hai first run ke liye
	saUser := os.Getenv("SUPERADMIN_USERNAME")
	saEmail := os.Getenv("SUPERADMIN_EMAIL")
	saPass := os.Getenv("SUPERADMIN_PASSWORD")

	log.Println("🌱 Seeding master data...")
	if err := services.SeedMasterData(config.MasterDB, saUser, saEmail, saPass); err != nil {
		log.Printf("⚠️  Warning: Failed to ensure master seed data: %v", err)
	} else {
		log.Println("✅ Master data seeded successfully")
	}

	// 7. Setup Router
	router := gin.Default()

	// CORS Setup (Taaki Frontend connect ho sake)
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Production mein isay specific domain karo (e.g. localhost:3000)
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 8. Register Routes
	routes.SetupRoutes(router)

	// 9. Start Server
	serverPort := cfg.ServerPort
	if serverPort == "" {
		serverPort = ":8080"
	}

	log.Printf("🚀 Server starting on port %s", serverPort)
	if err := router.Run(serverPort); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}
