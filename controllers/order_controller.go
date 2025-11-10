package controllers

// Order controller content from previous scaffoldpackage controllers

import (
	"fmt"
	"kalebecommerce/config"
	"kalebecommerce/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PlaceOrder - user places an order with product IDs & quantities
func PlaceOrder(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req []struct {
			ProductID string `json:"productId" binding:"required,uuid"`
			Quantity  int    `json:"quantity" binding:"required,min=1"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.JSON(c, http.StatusBadRequest, false, "validation error", nil, err.Error())
			return
		}

		userID := c.GetString("user_id")
		uid, _ := uuid.Parse(userID)

		err := db.Transaction(func(tx *gorm.DB) error {
			var total float64
			order := config.Order{ID: uuid.New().String(), UserID: uid, Status: "pending"}
			if err := tx.Create(&order).Error; err != nil {
				return err
			}

			for _, item := range req {
				pid, _ := uuid.Parse(item.ProductID)
				var p config.Product
				if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&p, "id = ?", pid).Error; err != nil {
					return err
				}

				if p.Stock < item.Quantity {
					return fmt.Errorf("insufficient stock for %s", p.Name)
				}
				p.Stock -= item.Quantity
				if err := tx.Save(&p).Error; err != nil {
					return err
				}

				oi := config.OrderItem{
					ID:        uuid.New().String(),
					OrderID:   order.ID,
					ProductID: pid,
					Quantity:  item.Quantity,
					UnitPrice: p.Price,
				}
				if err := tx.Create(&oi).Error; err != nil {
					return err
				}
				total += float64(item.Quantity) * p.Price
			}

			order.TotalPrice = total
			return tx.Save(&order).Error
		})

		if err != nil {
			if strings.Contains(err.Error(), "insufficient stock") {
				utils.JSON(c, http.StatusBadRequest, false, "insufficient stock", nil, err.Error())
				return
			}
			utils.JSON(c, http.StatusInternalServerError, false, "failed to place order", nil, err.Error())
			return
		}
		utils.JSON(c, http.StatusCreated, true, "order placed successfully", nil, nil)
	}
}

// ListOrders - view user orders
func ListOrders(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		uid, _ := uuid.Parse(userID)

		var orders []config.Order
		if err := db.Preload("Items").Where("user_id = ?", uid).Find(&orders).Error; err != nil {
			utils.JSON(c, http.StatusInternalServerError, false, "failed to fetch orders", nil, err.Error())
			return
		}

		utils.JSON(c, http.StatusOK, true, "orders retrieved", orders, nil)
	}
}
