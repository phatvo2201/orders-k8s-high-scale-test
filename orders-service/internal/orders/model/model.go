package model

import (
	"time"

	"github.com/lib/pq"
)

type OrderStatus string

const (
	StatusReceived   OrderStatus = "RECEIVED"
	StatusProcessing OrderStatus = "PROCESSING"
	StatusDone       OrderStatus = "DONE"
	StatusFailed     OrderStatus = "FAILED"
	StatusCancelled  OrderStatus = "CANCELLED"
)

type Order struct {
	ID         string          `gorm:"primaryKey" json:"id"`
	UserID     string          `gorm:"not null" json:"user_id"`
	ProductIDs pq.Int32Array   `gorm:"type:integer[]" json:"product_ids"` // Fixed: Use integer[]
	Quantities pq.Int32Array   `gorm:"type:integer[]" json:"quantities"`  // Slice of quantities for each product
	Prices     pq.Float64Array `gorm:"type:numeric[]" json:"prices"`      // Slice of prices for each product
	Status     OrderStatus     `gorm:"not null" json:"status"`
	CreatedAt  time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}
