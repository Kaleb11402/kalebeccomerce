package middleware

import (
	"kalebecommerce/config"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthRequired middleware ensures that a valid JWT token is provided
// before allowing access to protected routes.
func AuthRequired(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		auth := c.GetHeader("Authorization")
		if auth == "" {
			// Reject if token is missing
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "missing token"})
			return
		}

		// Remove "Bearer " prefix if present
		tokenStr := strings.TrimPrefix(auth, "Bearer ")

		// Parse and validate the JWT token
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			// Return the secret key for validation
			return []byte(cfg.JWTSecret), nil
		})

		// If token is invalid or parsing failed, reject the request
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
			return
		}

		// Extract claims (data) from the token
		claims := token.Claims.(jwt.MapClaims)

		// Store user_id and role in the Gin context for downstream handlers
		c.Set("user_id", claims["user_id"])
		c.Set("role", claims["role"])

		// Continue to the next handler
		c.Next()
	}
}

// AdminOnly middleware restricts access to only users with the "Admin" role.
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user role from the context (set in AuthRequired)
		role, _ := c.Get("role")

		// If the user is not an Admin, block access
		if role != "Admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "admin only"})
			return
		}

		// Continue to the next handler if user is Admin
		c.Next()
	}
}
