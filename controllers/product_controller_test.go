package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"kalebecommerce/config"
	"kalebecommerce/utils"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// NOTE: ProductInput is no longer fully used for Create/Update due to multipart form

// --- Helper for creating multipart requests ---

// createMultipartForm creates a multipart/form-data request body with fields and an optional file.
func createMultipartForm(t *testing.T, fields map[string]string, fileFieldName, filename string) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 1. Add form fields
	for key, value := range fields {
		err := writer.WriteField(key, value)
		assert.NoError(t, err)
	}

	// 2. Add file
	if fileFieldName != "" && filename != "" {
		// Create a mock file part
		part, err := writer.CreateFormFile(fileFieldName, filename)
		assert.NoError(t, err)

		// Write dummy content (e.g., "mock image content")
		_, err = io.WriteString(part, "mock image content")
		assert.NoError(t, err)
	}

	err := writer.Close()
	assert.NoError(t, err)

	return body, writer.FormDataContentType()
}

// --- 1. TestCreateProduct (Admin Route) ---

func TestCreateProduct_Success(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter()
	router.POST("/admin/products", mockAdminAuthMiddleware(), CreateProduct(db))

	// Ensure the uploads directory exists for cleanup
	os.MkdirAll(utils.UploadDir, 0755)
	defer os.RemoveAll(utils.UploadDir)

	fields := map[string]string{
		"name":        "Test Laptop",
		"description": "A powerful test machine.",
		"price":       "999.99", // Sent as string
		"stock":       "50",     // Sent as string
		"category":    "Electronics",
	}
	body, contentType := createMultipartForm(t, fields, "image", "test_image.jpg")

	req, _ := http.NewRequest("POST", "/admin/products", body)
	req.Header.Set("Content-Type", contentType)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "product created")

	// Verify product was created in DB and has an image URL
	var createdProduct config.Product
	db.Last(&createdProduct)
	assert.Equal(t, "Test Laptop", createdProduct.Name)
	assert.NotEqual(t, "", createdProduct.ImageURL, "ImageURL should be set after upload")
	assert.True(t, strings.HasSuffix(createdProduct.ImageURL, ".jpg"), "ImageURL should have the correct extension")

	// Verify file exists on disk (cleanup will happen via defer)
	assert.FileExists(t, "."+createdProduct.ImageURL)
}

func TestCreateProduct_ValidationFailure(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter()
	router.POST("/admin/products", mockAdminAuthMiddleware(), CreateProduct(db))

	// Invalid Input: Price is an invalid string (fails strconv.ParseFloat)
	fields := map[string]string{
		"name":        "Invalid Product",
		"description": "Should fail.",
		"price":       "not-a-number",
		"stock":       "10",
	}
	body, contentType := createMultipartForm(t, fields, "", "") // No file needed

	req, _ := http.NewRequest("POST", "/admin/products", body)
	req.Header.Set("Content-Type", contentType)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "validation error")
}

// --- 2. TestUpdateProduct (Admin Route) ---

func TestUpdateProduct_Success_WithImageUpdate(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter()
	// 1. Create a product to update
	productID := uuid.New()
	originalImageURL := filepath.Join(utils.UploadDir, productID.String()+".png")
	// Mock a previous image file
	os.MkdirAll(utils.UploadDir, 0755)
	os.WriteFile(originalImageURL, []byte("old image content"), 0644)
	defer os.RemoveAll(utils.UploadDir)

	db.Create(&config.Product{
		ID: productID.String(), Name: "Old Name", Price: 10.00, Stock: 5, ImageURL: "/" + originalImageURL,
	})

	router.PUT("/admin/products/:id", mockAdminAuthMiddleware(), UpdateProduct(db))

	// Updates payload using multipart form data
	updates := map[string]string{
		"name":  "New Name",
		"price": "20.00",
	}

	// Create request with a *new* image (which should overwrite the old one)
	body, contentType := createMultipartForm(t, updates, "image", "new_test_image.gif")

	url := fmt.Sprintf("/admin/products/%s", productID.String())
	req, _ := http.NewRequest("PUT", url, body)
	req.Header.Set("Content-Type", contentType)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "product updated")

	// Verify product was updated in DB
	var updatedProduct config.Product
	db.First(&updatedProduct, "id = ?", productID)
	assert.Equal(t, "New Name", updatedProduct.Name)
	assert.Equal(t, 20.00, updatedProduct.Price)
	assert.NotEqual(t, "/"+originalImageURL, updatedProduct.ImageURL, "ImageURL should be updated")
	assert.True(t, strings.HasSuffix(updatedProduct.ImageURL, ".gif"), "ImageURL should reflect the new file extension")

	// Check that the new file exists on disk
	assert.FileExists(t, "."+updatedProduct.ImageURL)
}

func TestUpdateProduct_Success_NoImageUpdate(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter()
	// 1. Create a product to update
	productID := uuid.New()
	db.Create(&config.Product{
		ID: productID.String(), Name: "Old Name", Price: 10.00, Stock: 5, ImageURL: "/uploads/products/static.png",
	})
	originalImageURL := "/uploads/products/static.png" // Keep original image URL

	router.PUT("/admin/products/:id", mockAdminAuthMiddleware(), UpdateProduct(db))

	// Updates payload using multipart form data (no file attached)
	updates := map[string]string{
		"name":  "New Name Only",
		"stock": "10",
	}

	body, contentType := createMultipartForm(t, updates, "", "")

	url := fmt.Sprintf("/admin/products/%s", productID.String())
	req, _ := http.NewRequest("PUT", url, body)
	req.Header.Set("Content-Type", contentType)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify product was updated in DB
	var updatedProduct config.Product
	db.First(&updatedProduct, "id = ?", productID)
	assert.Equal(t, "New Name Only", updatedProduct.Name)
	assert.Equal(t, 10.00, updatedProduct.Price) // Price should be unchanged
	assert.Equal(t, 10, updatedProduct.Stock)
	assert.Equal(t, originalImageURL, updatedProduct.ImageURL, "ImageURL should remain the same")
}

func TestUpdateProduct_NotFound(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter()
	router.PUT("/admin/products/:id", mockAdminAuthMiddleware(), UpdateProduct(db))

	nonExistentID := uuid.New().String()
	updates := map[string]string{"name": "Non-existent Update"}

	body, contentType := createMultipartForm(t, updates, "", "")

	url := fmt.Sprintf("/admin/products/%s", nonExistentID)
	req, _ := http.NewRequest("PUT", url, body)
	req.Header.Set("Content-Type", contentType)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "product not found")
}

// --- 3. TestDeleteProduct (Admin Route) ---

func TestDeleteProduct_Success_WithImageDeletion(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter()
	// 1. Create a product and a mock file to delete
	productID := uuid.New()
	imagePath := filepath.Join(utils.UploadDir, productID.String()+".jpg")
	os.MkdirAll(utils.UploadDir, 0755)
	os.WriteFile(imagePath, []byte("image to be deleted"), 0644)
	defer os.RemoveAll(utils.UploadDir)

	db.Create(&config.Product{
		ID: productID.String(), Name: "Delete Me", Price: 1.00, Stock: 1, ImageURL: "/" + imagePath,
	})

	assert.FileExists(t, imagePath, "Precondition: Image file must exist before deletion test.")

	router.DELETE("/admin/products/:id", mockAdminAuthMiddleware(), DeleteProduct(db))

	url := fmt.Sprintf("/admin/products/%s", productID.String())
	req, _ := http.NewRequest("DELETE", url, nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "product deleted")

	// Verify product was deleted (record not found)
	var deletedProduct config.Product
	err := db.First(&deletedProduct, "id = ?", productID).Error
	assert.Equal(t, gorm.ErrRecordNotFound, err)

	// Verify image file was deleted
	assert.NoFileExists(t, imagePath, "Image file should be deleted from disk.")
}

func TestDeleteProduct_Success_WithoutImage(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter()
	// 1. Create a product with no image
	productID := uuid.New()
	db.Create(&config.Product{
		ID: productID.String(), Name: "No Image", Price: 1.00, Stock: 1, ImageURL: "",
	})

	router.DELETE("/admin/products/:id", mockAdminAuthMiddleware(), DeleteProduct(db))

	url := fmt.Sprintf("/admin/products/%s", productID.String())
	req, _ := http.NewRequest("DELETE", url, nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "product deleted")
}

// --- 4. TestGetProduct (Public Route) ---
// (No change needed as logic is unaffected)

func TestGetProduct_Success(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter()
	// 1. Create a product to fetch
	productID := uuid.New()
	db.Create(&config.Product{
		ID: productID.String(), Name: "Fetch Test", Price: 50.00, Stock: 10,
	})

	router.GET("/products/:id", GetProduct(db))

	url := fmt.Sprintf("/products/%s", productID.String())
	req, _ := http.NewRequest("GET", url, nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "product retrieved")
	assert.Contains(t, w.Body.String(), "Fetch Test")
}

func TestGetProduct_NotFound(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter()
	router.GET("/products/:id", GetProduct(db))

	nonExistentID := uuid.New().String()
	url := fmt.Sprintf("/products/%s", nonExistentID)
	req, _ := http.NewRequest("GET", url, nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "product not found")
}

// --- 5. TestListOrSearchProducts (Public Route) ---
// (No change needed as logic is unaffected)

func TestListOrSearchProducts_Success(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter()
	cacheL := cache.New(cache.NoExpiration, cache.NoExpiration)
	router.GET("/products", ListOrSearchProducts(db, cacheL))

	// 1. Create multiple products for testing pagination/search
	db.Create(&config.Product{ID: uuid.New().String(), Name: "Apple iPad", Price: 500, Stock: 10})
	db.Create(&config.Product{ID: uuid.New().String(), Name: "Samsung Galaxy", Price: 700, Stock: 20})
	db.Create(&config.Product{ID: uuid.New().String(), Name: "Apple Watch", Price: 250, Stock: 5})

	req, _ := http.NewRequest("GET", "/products", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "products listed")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// FIX: Access the content under the "object" key, as defined by BaseResponse
	object, ok := response["object"].(map[string]interface{})
	assert.True(t, ok, "Response 'object' field missing or not a map")

	// Check total count is under the "object" map (JSON unmarshals numbers to float64)
	assert.Equal(t, float64(3), object["totalProducts"])
	assert.Contains(t, w.Body.String(), "Apple iPad")
}

func TestListOrSearchProducts_Search(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter()
	cacheL := cache.New(cache.NoExpiration, cache.NoExpiration)
	router.GET("/products", ListOrSearchProducts(db, cacheL))

	// 1. Create products
	db.Create(&config.Product{ID: uuid.New().String(), Name: "Blue Shirt", Price: 10, Stock: 1})
	db.Create(&config.Product{ID: uuid.New().String(), Name: "Red Dress", Price: 20, Stock: 1})
	db.Create(&config.Product{ID: uuid.New().String(), Name: "Blue Jeans", Price: 30, Stock: 1})

	// Search for "blue" (case-insensitive)
	req, _ := http.NewRequest("GET", "/products?search=blue", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "products listed")

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	object := response["object"].(map[string]interface{})
	assert.Equal(t, float64(2), object["totalProducts"]) // Blue Shirt, Blue Jeans
	assert.Contains(t, w.Body.String(), "Blue Shirt")
	assert.NotContains(t, w.Body.String(), "Red Dress")
}

func TestListOrSearchProducts_Pagination(t *testing.T) {
	db := setupTestDB(t)
	router := setupRouter()
	cacheL := cache.New(cache.NoExpiration, cache.NoExpiration)
	router.GET("/products", ListOrSearchProducts(db, cacheL))

	// 1. Create 5 products
	for i := 1; i <= 5; i++ {
		db.Create(&config.Product{
			ID: uuid.New().String(), Name: fmt.Sprintf("Item %d", i), Price: float64(i), Stock: 1,
		})
	}

	// Request page 2, limit 2
	req, _ := http.NewRequest("GET", "/products?page=2&limit=2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"currentPage":2`)
	assert.Contains(t, w.Body.String(), `"pageSize":2`)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// FIX: Access the content under the "object" key
	object, ok := response["object"].(map[string]interface{})
	assert.True(t, ok, "Response 'object' field missing or not a map")

	// Check total count
	assert.Equal(t, float64(5), object["totalProducts"])

	// Access products array from the 'object' map
	products, ok := object["products"].([]interface{})
	assert.True(t, ok, "Products field missing or not an array in object")

	assert.Len(t, products, 2)
}
