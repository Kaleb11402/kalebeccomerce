package controllers

import (
	"kalebecommerce/config"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	models := []interface{}{&config.User{}, &config.Product{}, &config.Order{}, &config.OrderItem{}}

	// Drop all tables
	db.Migrator().DropTable(models...)

	// Re-migrate
	err = db.AutoMigrate(models...)
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}

// setupRouter initializes Gin for testing
func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.Default()
}

// mockAuthMiddleware is a simple middleware to set the user_id context for tests
// Used in Order and other protected controllers.
func mockAuthMiddleware(userID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	}
}

// mockAdminAuthMiddleware is a simple middleware to simulate an authenticated Admin user.
func mockAdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set a role or flag that your actual middleware would set
		c.Set("user_role", "Admin")
		c.Next()
	}
}

// mockConfig returns a test config (used in auth tests)
func mockConfig() *config.Config {
	return &config.Config{JWTSecret: "testsecret"}
}
