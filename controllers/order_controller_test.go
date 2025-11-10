package controllers

import (
	"bytes"
	"encoding/json"
	"kalebecommerce/config"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// OrderItemRequest struct to match the request body for PlaceOrder
type OrderItemRequest struct {
	ProductID string `json:"productId" binding:"required,uuid"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

// TestPlaceOrder_Success tests successful order placement
func TestPlaceOrder_Success(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter()
	testUserID := uuid.New().String()
	testProductID := uuid.New()

	// 1. Create a User (not strictly needed by controller but good practice)
	user := config.User{ID: testUserID, Username: "testuser", Email: "test@example.com"}
	db.Create(&user)

	// 2. Create a Product with stock
	product := config.Product{ID: testProductID.String(), Name: "Test Product", Price: 100.00, Stock: 5}
	db.Create(&product)

	router.POST("/orders", mockAuthMiddleware(testUserID), PlaceOrder(db))

	requestBody := []OrderItemRequest{
		{ProductID: testProductID.String(), Quantity: 2}, // Order 2 units
	}
	jsonBody, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "order placed successfully")

	// Verify database state: Order created, Stock updated
	var order config.Order
	db.Last(&order)
	assert.Equal(t, float64(200.00), order.TotalPrice) // 2 * 100.00

	var updatedProduct config.Product
	db.First(&updatedProduct, "id = ?", testProductID)
	assert.Equal(t, 3, updatedProduct.Stock) // Initial 5 - 2 ordered = 3
}

// TestPlaceOrder_InsufficientStock tests stock check failure
func TestPlaceOrder_InsufficientStock(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter()
	testUserID := uuid.New().String()
	testProductID := uuid.New()

	// 1. Create Product with LOW stock
	product := config.Product{ID: testProductID.String(), Name: "Low Stock Item", Price: 50.00, Stock: 1}
	db.Create(&product)

	router.POST("/orders", mockAuthMiddleware(testUserID), PlaceOrder(db))

	requestBody := []OrderItemRequest{
		{ProductID: testProductID.String(), Quantity: 5}, // Request 5 units, but only 1 in stock
	}
	jsonBody, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "insufficient stock")

	// Verify database state: Stock should NOT be updated (transaction rollback)
	var updatedProduct config.Product
	db.First(&updatedProduct, "id = ?", testProductID)
	assert.Equal(t, 1, updatedProduct.Stock) // Should remain 1
}

// TestPlaceOrder_ValidationFailure tests invalid request body
func TestPlaceOrder_ValidationFailure(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter()
	testUserID := uuid.New().String()

	router.POST("/orders", mockAuthMiddleware(testUserID), PlaceOrder(db))

	// Invalid Request: Quantity 0 (min=1 validation fails)
	requestBody := []map[string]interface{}{
		{"productId": uuid.New().String(), "quantity": 0},
	}
	jsonBody, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "validation error")
}
