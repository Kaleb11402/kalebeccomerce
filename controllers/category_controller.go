package controllers

import (
	"kalebecommerce/config"
	"kalebecommerce/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RegisterCategory struct {
	name string `json:"name" binding:"required,alphanum,min=3,max=30"`
}

func CreateCategory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input RegisterCategory
		if err := c.ShouldBindJSON(&input); err != nil {
			utils.JSON(c, http.StatusBadRequest, false, "validation error", nil, err.Error())
			return
		}

		// 4. Create User
		category := config.Category{
			ID:   uuid.New().String(),
			Name: input.name,
		}
		if err := db.Create(&category).Error; err != nil {
			utils.JSON(c, http.StatusInternalServerError, false, "failed to create cateogry", nil, err.Error())
			return
		}

		utils.JSON(c, http.StatusCreated, true, "user created",
			gin.H{"id": category.ID, "categoryname": category.Name}, nil)
	}
}
