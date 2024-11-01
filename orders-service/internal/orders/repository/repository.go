package repository

import (
	"log"

	"gorm.io/gorm"

	"github.com/pbb/orders-service/internal/orders/model"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) SaveOrder(order *model.Order) error {
	if err := r.db.Create(order).Error; err != nil {
		log.Printf("Failed to save order: %v", err)
		return err
	}
	return nil
}

func (r *OrderRepository) UpdateOrderStatus(orderID string, status model.OrderStatus) error {
	if err := r.db.Model(&model.Order{}).
		Where("id = ?", orderID).
		Update("status", status).Error; err != nil {
		log.Printf("Failed to update order status: %v", err)
		return err
	}
	return nil
}
