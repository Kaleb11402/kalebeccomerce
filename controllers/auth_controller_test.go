package controllers

import (
	"bytes"
	"encoding/json"
	"kalebecommerce/config"
	"kalebecommerce/utils"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRegister_Success(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter()
	// Assuming Register is defined as func Register(db *gorm.DB) gin.HandlerFunc
	router.POST("/register", Register(db))

	body := RegisterInput{
		Username: "kaleb",
		Email:    "kaleb@example.com",
		Password: "Strong@123",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "user created")
}

// TestRegister_WeakPassword ensures weak password returns an error
func TestRegister_WeakPassword(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter()
	// Assuming Register is defined as func Register(db *gorm.DB) gin.HandlerFunc
	router.POST("/register", Register(db))

	body := RegisterInput{
		Username: "weakuser",
		Email:    "weak@example.com",
		Password: "weak", // Fails validation due to 'min' tag
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	// FIX: Changed the assertion to check for the actual response message
	assert.Contains(t, w.Body.String(), "validation error")
}

// TestRegister_DuplicateUser checks for duplicate email or username
func TestRegister_DuplicateUser(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter()
	// Assuming Register is defined as func Register(db *gorm.DB) gin.HandlerFunc
	router.POST("/register", Register(db))

	user := config.User{
		ID:       uuid.New().String(),
		Username: "kaleb",
		Email:    "kaleb@example.com",
		Password: "hashedpass",
	}
	db.Create(&user)

	body := RegisterInput{
		Username: "kaleb",
		Email:    "kaleb@example.com",
		Password: "Strong@123",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "user already exists")
}

// TestLogin_Success tests successful login
func TestLogin_Success(t *testing.T) {
	db := setupTestDB(t)
	cfg := mockConfig()
	router := setupRouter()
	// Assuming Login is defined as func Login(db *gorm.DB, cfg *config.Config) gin.HandlerFunc
	router.POST("/login", Login(db, cfg))

	hash, _ := utils.HashPassword("Strong@123")
	user := config.User{
		ID:       uuid.New().String(),
		Username: "kaleb",
		Email:    "kaleb@example.com",
		Password: hash,
		Role:     "User",
	}
	db.Create(&user)

	body := LoginInput{
		Email:    "kaleb@example.com",
		Password: "Strong@123",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "login successful")
}

// TestLogin_InvalidPassword tests login with wrong password
func TestLogin_InvalidPassword(t *testing.T) {
	db := setupTestDB(t)
	cfg := mockConfig()
	router := setupRouter()
	// Assuming Login is defined as func Login(db *gorm.DB, cfg *config.Config) gin.HandlerFunc
	router.POST("/login", Login(db, cfg))

	hash, _ := utils.HashPassword("Strong@123")
	user := config.User{
		ID:       uuid.New().String(),
		Username: "kaleb",
		Email:    "kaleb@example.com",
		Password: hash,
	}
	db.Create(&user)

	body := LoginInput{
		Email:    "kaleb@example.com",
		Password: "WrongPass@123",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid credentials")
}

// TestLogin_UserNotFound tests login with non-existing email
func TestLogin_UserNotFound(t *testing.T) {
	db := setupTestDB(t)
	cfg := mockConfig()
	router := setupRouter()
	// Assuming Login is defined as func Login(db *gorm.DB, cfg *config.Config) gin.HandlerFunc
	router.POST("/login", Login(db, cfg))

	body := LoginInput{
		Email:    "unknown@example.com",
		Password: "Strong@123",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid credentials")
}
