package config

import (
	"os"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	Port        string
}

func GetConfig() *Config {
	return &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		Port:        os.Getenv("PORT"),
	}
}

func InitDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")
	err = db.AutoMigrate(&User{}, &Product{}, &Order{}, &OrderItem{})
	return db, err
}

// Models
type User struct {
	ID        string `gorm:"primaryKey" json:"id" json:"id"`
	Username  string `gorm:"uniqueIndex;not null" json:"username"`
	Email     string `gorm:"uniqueIndex;not null" json:"email"`
	Password  string `gorm:"not null" json:"-"`
	Role      string `gorm:"default:'User'" json:"role"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Product struct {
	ID          string     `gorm:"primaryKey" json:"id" json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	ImageURL    string     `json:"image_url"`
	Price       float64    `json:"price"`
	Stock       int        `json:"stock"`
	Category    string     `json:"category"`
	UserID      *uuid.UUID `json:"user_id"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Order struct {
	ID         string      `gorm:"primaryKey" json:"id" json:"id"`
	UserID     uuid.UUID   `json:"user_id"`
	TotalPrice float64     `json:"total_price"`
	Status     string      `json:"status"`
	Items      []OrderItem `gorm:"foreignKey:OrderID" json:"items"`
	CreatedAt  time.Time
}

type OrderItem struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	OrderID   string    `json:"order_id"`
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
	UnitPrice float64   `json:"unit_price"`
}
