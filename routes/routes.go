package routes

import (
	"kalebecommerce/config"
	"kalebecommerce/controllers"
	"kalebecommerce/middleware"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"
)

// SetupRouter sets up all API routes, middleware, and rate limiting.
func SetupRouter(db *gorm.DB, cfg *config.Config, productCache *cache.Cache) *gin.Engine {
	r := gin.Default()

	// 🧩 Global rate limiter: 5 requests every 10 seconds per IP
	r.Use(middleware.RateLimitMiddleware(cache.New(10*time.Second, 20*time.Second), 5, 10*time.Second))

	api := r.Group("/api")

	// 🔐 Authentication routes (login/register)
	api.POST("/auth/register", controllers.Register(db))
	api.POST("/auth/login", controllers.Login(db, cfg))

	// 🛍 Public product routes (with cache)
	api.GET("/products", controllers.ListOrSearchProducts(db, productCache))
	api.GET("/products/:id", controllers.GetProduct(db))

	// 👤 User routes (require login)
	auth := api.Group("").Use(middleware.AuthRequired(cfg))
	auth.POST("/orders", controllers.PlaceOrder(db))
	auth.GET("/orders", controllers.ListOrders(db))

	// 🧑‍💼 Admin routes (require admin role)
	admin := api.Group("").Use(middleware.AuthRequired(cfg), middleware.AdminOnly())
	admin.POST("/products", controllers.CreateProduct(db))
	admin.PUT("/products/:id", controllers.UpdateProduct(db))
	admin.DELETE("/products/:id", controllers.DeleteProduct(db))

	return r
}
