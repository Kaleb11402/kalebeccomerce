package main

import (
	"kalebecommerce/cache" // Import the cache package
	"kalebecommerce/config"
	"kalebecommerce/routes"
	"log"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found, reading environment variables")
	}

	cfg := config.GetConfig()
	db, err := config.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to init db: %v", err)
	}
	defer closeDB(db)

	// Initialize the in-memory cache
	cache.InitCache()

	// Pass the cache instance to the router setup function
	r := routes.SetupRouter(db, cfg, cache.Cache)
	port := cfg.Port
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func closeDB(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}
}
