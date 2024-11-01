package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"

	"github.com/pbb/orders-service/internal/orders/model"
	"github.com/pbb/orders-service/internal/orders/repository"
	product_repository "github.com/pbb/orders-service/internal/product/repository"
	redis_helper "github.com/pbb/orders-service/pkg/redis"
)

// OrderService handles the business logic for orders.
type OrderService struct {
	redisClient *redis.Client
	rabbitMQCh  *amqp.Channel
	orderRepo   *repository.OrderRepository
	productRepo *product_repository.ProductRepository
}

// NewOrderService initializes a new OrderService with required dependencies.
func NewOrderService(redisClient *redis.Client, rabbitMQCh *amqp.Channel, orderRepo *repository.OrderRepository, productRepo *product_repository.ProductRepository) *OrderService {
	return &OrderService{
		redisClient: redisClient,
		rabbitMQCh:  rabbitMQCh,
		orderRepo:   orderRepo,
		productRepo: productRepo,
	}
}

// SaveOrder saves the order to the database.
func (s *OrderService) SaveOrder(order *model.Order) error {
	return s.orderRepo.SaveOrder(order)
}

// ProcessOrder sends the order to RabbitMQ for processing.
func (s *OrderService) ProcessOrder(order model.Order) error {
	lockKey := "order-lock:" + order.ID

	// Attempt to acquire lock before queuing the order
	locked, err := redis_helper.AcquireLock(lockKey, 5*time.Minute)
	if err != nil {
		return err
	}

	if !locked {
		return fmt.Errorf("order %s is already being processed", order.ID)
	}

	body, _ := json.Marshal(order)
	err = s.rabbitMQCh.Publish(
		"", "orders", false, false, amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	// Release the lock after sending the order to RabbitMQ
	if err == nil {
		redis_helper.ReleaseLock(lockKey)
	}

	return err
}
