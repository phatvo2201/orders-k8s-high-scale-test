package worker

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/lib/pq"
	"github.com/streadway/amqp"

	"github.com/pbb/orders-service/internal/orders/model"
	"github.com/pbb/orders-service/internal/orders/repository"
	product_repository "github.com/pbb/orders-service/internal/product/repository"
	redis_helper "github.com/pbb/orders-service/pkg/redis"
)

type OrderWorker struct {
	orderRepo   *repository.OrderRepository
	productRepo *product_repository.ProductRepository
}

func NewOrderWorker(orderRepo *repository.OrderRepository, productRepo *product_repository.ProductRepository) *OrderWorker {
	return &OrderWorker{orderRepo: orderRepo, productRepo: productRepo}
}

func (w *OrderWorker) StartWorker(id int, jobs <-chan amqp.Delivery, wg *sync.WaitGroup, done <-chan struct{}) {
	defer wg.Done()

	for {
		select {
		case msg, ok := <-jobs:
			if !ok {
				log.Printf("Worker %v is stopping.", id)
				return
			}

			var order model.Order
			if err := json.Unmarshal(msg.Body, &order); err != nil {
				log.Printf("Worker %v: Invalid input: %v", id, err)
				// Order failed due to invalid input
				order.Status = model.StatusFailed
				w.updateOrderFailure(order) // Call helper function to update failure
				msg.Nack(false, false)      // Reject without requeue, bad message format
				continue
			}

			log.Printf("Worker %d: Processing order: %+v", id, order)
			order.Status = model.StatusProcessing

			// Acquire Redis lock to ensure only one worker processes this order
			lockKey := "order-lock:" + order.ID
			locked, err := redis_helper.AcquireLock(lockKey, 5*time.Minute)
			if err != nil {
				log.Printf("Worker %d: Failed to acquire lock for order %s: %v", id, order.ID, err)
				order.Status = model.StatusFailed
				w.updateOrderFailure(order)
				msg.Nack(false, true)
				continue
			}

			if !locked {
				log.Printf("Order %s is already being processed by another worker", order.ID)
				msg.Ack(false)
				continue
			}

			productIDs := ConvertInt32ArrayToSlice(order.ProductIDs)
			quantities := ConvertInt32ArrayToSlice(order.Quantities)

			err = w.productRepo.CheckAndUpdateProductQuantityWithTrans(productIDs, quantities)
			if err != nil {
				log.Printf("Worker %d: Failed to check/update product quantities for order %s: %v", id, order.ID, err)
				order.Status = model.StatusFailed
				w.updateOrderFailure(order)
				// msg.Nack(false, true)
				redis_helper.ReleaseLock(lockKey)
				continue
			}

			// Update order status to "done"
			order.Status = model.StatusDone
			if err := w.orderRepo.UpdateOrderStatus(order.ID, order.Status); err != nil {
				log.Printf("Worker %d: Failed to update status for order %s: %v", id, order.ID, err)
				// Order failed due to status update failure
				order.Status = model.StatusFailed
				w.updateOrderFailure(order) // Call helper function to update failure
				msg.Nack(false, true)       // Requeue the message
			} else {
				log.Printf("Worker %d: Order %s completed successfully", id, order.ID)
				msg.Ack(false)
			}

			// Release the Redis lock after processing
			if err := redis_helper.ReleaseLock(lockKey); err != nil {
				log.Printf("Worker %d: Failed to release lock for order %s: %v", id, order.ID, err)
			}

		case <-done:
			log.Printf("Worker %d: received shutdown signal, stopping worker.", id)
			return
		}
	}
}

// Helper function to update the order status to "failed" in case of errors
func (w *OrderWorker) updateOrderFailure(order model.Order) {
	order.Status = model.StatusFailed
	if err := w.orderRepo.UpdateOrderStatus(order.ID, order.Status); err != nil {
		log.Printf("Failed to update order %s to failed status: %v", order.ID, err)
	}
}

func (w *OrderWorker) Start(ch *amqp.Channel, queueName string, workerCount int, done <-chan struct{}) {
	msgs, err := ch.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to start RabbitMQ consumer: %v", err)
	}

	jobsQueue := make(chan amqp.Delivery, workerCount)

	var wg sync.WaitGroup

	for i := 1; i <= workerCount; i++ {
		wg.Add(1)
		go w.StartWorker(i, jobsQueue, &wg, done)
	}

	go func() {
		for msg := range msgs {
			select {
			case jobsQueue <- msg:
			case <-done:
				log.Println("Shutting down workers.")
				return
			}
		}
	}()

	wg.Wait()

	close(jobsQueue)
}

func ConvertInt32ArrayToSlice(arr pq.Int32Array) []int {
	slice := make([]int, len(arr))
	for i, v := range arr {
		slice[i] = int(v)
	}
	return slice
}
