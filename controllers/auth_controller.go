package controllers

import (
	"kalebecommerce/config"
	"kalebecommerce/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// --- Structs ---

type RegisterInput struct {
	Username string `json:"username" binding:"required,alphanum,min=3,max=30"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// --- Register Handler ---

func Register(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input RegisterInput
		if err := c.ShouldBindJSON(&input); err != nil {
			utils.JSON(c, http.StatusBadRequest, false, "validation error", nil, err.Error())
			return
		}

		// 1. Password Strength Check (using utils)
		if !utils.IsStrongPassword(input.Password) {
			utils.JSON(c, http.StatusBadRequest, false, "weak password", nil,
				"password must include uppercase, lowercase, number, and special character")
			return
		}

		// 2. Uniqueness Check
		var count int64
		// Use db.Where().Take() or db.Where().First() instead of db.Count()
		// for slight performance gain if you only need existence confirmation.
		db.Model(&config.User{}).Where("email = ? OR username = ?", input.Email, input.Username).Count(&count)
		if count > 0 {
			utils.JSON(c, http.StatusBadRequest, false, "user already exists", nil, nil)
			return
		}

		// 3. Hashing (using utils)
		hash, err := utils.HashPassword(input.Password)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, false, "failed to hash password", nil, err.Error())
			return
		}

		// 4. Create User
		user := config.User{
			ID:       uuid.New().String(),
			Username: input.Username,
			Email:    input.Email,
			Password: hash,
			Role:     "User",
		}
		if err := db.Create(&user).Error; err != nil {
			utils.JSON(c, http.StatusInternalServerError, false, "failed to create user", nil, err.Error())
			return
		}

		utils.JSON(c, http.StatusCreated, true, "user created",
			gin.H{"id": user.ID, "username": user.Username, "email": user.Email}, nil)
	}
}

// --- Login Handler ---

func Login(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input LoginInput
		if err := c.ShouldBindJSON(&input); err != nil {
			utils.JSON(c, http.StatusBadRequest, false, "validation error", nil, err.Error())
			return
		}

		var user config.User
		if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
			// Do NOT reveal if the error is "record not found" or "invalid password".
			// Return generic 'invalid credentials' for security.
			utils.JSON(c, http.StatusUnauthorized, false, "invalid credentials", nil, nil)
			return
		}

		// 1. Password Comparison (using utils)
		if !utils.CheckPasswordHash(input.Password, user.Password) {
			utils.JSON(c, http.StatusUnauthorized, false, "invalid credentials", nil, nil)
			return
		}

		// 2. Token Generation
		claims := jwt.MapClaims{
			"user_id":  user.ID,
			"username": user.Username,
			"role":     user.Role,
			"exp":      time.Now().Add(72 * time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString([]byte(cfg.JWTSecret))
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, false, "token generation failed", nil, err.Error())
			return
		}

		utils.JSON(c, http.StatusOK, true, "login successful", gin.H{"token": signed}, nil)
	}
}
