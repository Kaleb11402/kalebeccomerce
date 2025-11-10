package controllers

import (
	"fmt"

	"kalebecommerce/config"
	"kalebecommerce/utils"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"
)

// CreateProduct (Admin) - Now accepts multipart/form-data
func CreateProduct(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Since we are handling file uploads, we read data from the form
		var in struct {
			Name        string `form:"name" binding:"required"`
			Description string `form:"description" binding:"required"`
			Price       string `form:"price" binding:"required"` // Read as string, convert later
			Stock       string `form:"stock" binding:"required"` // Read as string, convert later
			Category    string `form:"category"`
		}

		// Use c.ShouldBind to handle form data binding
		if err := c.ShouldBind(&in); err != nil {
			utils.JSON(c, http.StatusBadRequest, false, "validation error in form fields", nil, err.Error())
			return
		}

		// Convert string fields to their correct types
		price, err := strconv.ParseFloat(in.Price, 64)
		if err != nil || price <= 0 {
			utils.JSON(c, http.StatusBadRequest, false, "validation error", nil, "price must be a valid number greater than 0")
			return
		}
		stock, err := strconv.Atoi(in.Stock)
		if err != nil || stock < 0 {
			utils.JSON(c, http.StatusBadRequest, false, "validation error", nil, "stock must be a valid non-negative integer")
			return
		}

		// Handle file upload
		file, err := c.FormFile("image")
		var imageURL string
		if err == nil {
			// File was provided, save it
			productID := uuid.New().String()
			imageURL, err = utils.SaveUploadedFile(file, productID)
			if err != nil {
				utils.JSON(c, http.StatusInternalServerError, false, "failed to save image", nil, err.Error())
				return
			}
		} else if err != http.ErrMissingFile {
			// Handle other file related errors (like size limit)
			utils.JSON(c, http.StatusBadRequest, false, "file error", nil, err.Error())
			return
		}
		// If err == http.ErrMissingFile, we proceed with an empty imageURL

		p := config.Product{
			ID:          uuid.New().String(),
			Name:        in.Name,
			Description: in.Description,
			Price:       price,
			Stock:       stock,
			Category:    in.Category,
			ImageURL:    imageURL, // Store the path
		}

		if err := db.Create(&p).Error; err != nil {
			utils.JSON(c, http.StatusInternalServerError, false, "failed to create product", nil, err.Error())
			return
		}
		utils.JSON(c, http.StatusCreated, true, "product created", p, nil)
	}
}

// UpdateProduct (Admin) - Now handles image update via form data
func UpdateProduct(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		pid, err := uuid.Parse(id)
		if err != nil {
			utils.JSON(c, http.StatusBadRequest, false, "invalid product id", nil, nil)
			return
		}

		var p config.Product
		if err := db.First(&p, "id = ?", pid).Error; err != nil {
			utils.JSON(c, http.StatusNotFound, false, "product not found", nil, nil)
			return
		}

		// Handle file upload (optional image replacement)
		file, err := c.FormFile("image")
		var imageURL string
		if err == nil {
			// File was provided, save it
			// Use existing product ID to keep filename consistent (and potentially overwrite old image)
			imageURL, err = utils.SaveUploadedFile(file, p.ID)
			if err != nil {
				utils.JSON(c, http.StatusInternalServerError, false, "failed to save image", nil, err.Error())
				return
			}
			// Update the ImageURL in the product object
			p.ImageURL = imageURL
		} else if err != http.ErrMissingFile {
			// Handle other file related errors (like size limit)
			utils.JSON(c, http.StatusBadRequest, false, "file error", nil, err.Error())
			return
		}

		// Bind all other fields from the form
		var in struct {
			Name        string `form:"name"`
			Description string `form:"description"`
			Price       string `form:"price"`
			Stock       string `form:"stock"`
			Category    string `form:"category"`
		}

		if err := c.ShouldBind(&in); err != nil {
			// Note: Gin's form binding requires explicit checks for partial updates on empty fields.
			// For simplicity and robust parsing, we'll manually check for non-empty string updates.
		}

		updates := make(map[string]interface{})

		if in.Name != "" {
			updates["name"] = in.Name
		}
		if in.Description != "" {
			updates["description"] = in.Description
		}
		if in.Category != "" {
			updates["category"] = in.Category
		}

		// Handle Price update
		if in.Price != "" {
			price, err := strconv.ParseFloat(in.Price, 64)
			if err == nil && price > 0 {
				updates["price"] = price
			}
		}

		// Handle Stock update
		if in.Stock != "" {
			stock, err := strconv.Atoi(in.Stock)
			if err == nil && stock >= 0 {
				updates["stock"] = stock
			}
		}

		// Add image update if a file was provided
		if imageURL != "" {
			updates["image_url"] = imageURL
		}

		if len(updates) == 0 && err == http.ErrMissingFile {
			utils.JSON(c, http.StatusOK, true, "no fields to update", p, nil)
			return
		}

		if err := db.Model(&p).Updates(updates).Error; err != nil {
			utils.JSON(c, http.StatusInternalServerError, false, "update failed", nil, err.Error())
			return
		}

		// Reload the product to ensure the response is up-to-date
		db.First(&p, "id = ?", pid)
		utils.JSON(c, http.StatusOK, true, "product updated", p, nil)
	}
}

// DeleteProduct (Admin)
func DeleteProduct(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		pid, err := uuid.Parse(id)
		if err != nil {
			utils.JSON(c, http.StatusBadRequest, false, "invalid product id", nil, nil)
			return
		}

		// OPTIONAL: Delete the associated image file from disk
		var p config.Product
		// FIX: Separate the short declaration from the boolean expression
		if err := db.First(&p, "id = ?", pid).Error; err == nil && p.ImageURL != "" {
			// Prepend "." to handle relative path from root
			if err := os.Remove("." + p.ImageURL); err != nil {
				fmt.Printf("Warning: Failed to delete file %s: %v\n", p.ImageURL, err)
				// Continue deletion even if file deletion fails
			}
		}

		if err := db.Delete(&config.Product{}, "id = ?", pid).Error; err != nil {
			utils.JSON(c, http.StatusInternalServerError, false, "delete failed", nil, err.Error())
			return
		}
		utils.JSON(c, http.StatusOK, true, "product deleted", nil, nil)
	}
}

// GetProduct (Public)
func GetProduct(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		pid, err := uuid.Parse(id)
		if err != nil {
			utils.JSON(c, http.StatusBadRequest, false, "invalid product id", nil, nil)
			return
		}

		var product config.Product
		if err := db.First(&product, "id = ?", pid).Error; err != nil {
			utils.JSON(c, http.StatusNotFound, false, "product not found", nil, nil)
			return
		}
		utils.JSON(c, http.StatusOK, true, "product retrieved", product, nil)
	}
}

// ListProducts (Public)
func ListOrSearchProducts(db *gorm.DB, productCache *cache.Cache) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		search := strings.ToLower(c.DefaultQuery("search", ""))
		offset := (page - 1) * limit

		var products []config.Product
		query := db.Model(&config.Product{})
		if search != "" {
			query = query.Where("LOWER(name) LIKE ?", "%"+search+"%")
		}

		var total int64
		query.Count(&total)
		query.Offset(offset).Limit(limit).Find(&products)

		utils.JSON(c, http.StatusOK, true, "products listed",
			gin.H{
				"currentPage":   page,
				"pageSize":      limit,
				"totalProducts": total,
				"products":      products,
			}, nil)
	}
}
