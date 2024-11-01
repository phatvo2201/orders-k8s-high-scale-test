package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	db_pg "github.com/pbb/orders-service/internal/orders/db"
	"github.com/pbb/orders-service/internal/orders/handler"
	"github.com/pbb/orders-service/internal/orders/model"
	"github.com/pbb/orders-service/internal/orders/repository"
	"github.com/pbb/orders-service/internal/orders/service"
	"github.com/pbb/orders-service/internal/orders/worker"
	product_db "github.com/pbb/orders-service/internal/product/db"
	product "github.com/pbb/orders-service/internal/product/model"
	product_repository "github.com/pbb/orders-service/internal/product/repository"
	rabbitmq_helper "github.com/pbb/orders-service/pkg/rabbitmq"
	redis_helper "github.com/pbb/orders-service/pkg/redis"
)

func main() {

	//change to use env
	postgresURL := os.Getenv("POSTGRES_URL")
	db := db_pg.InitDB(postgresURL)

	db_pg.DB = db

	if err := db.AutoMigrate(&model.Order{}); err != nil {
		log.Fatalf("Failed to migrate orders table : %v", err)
	}

	//init db for product
	dns2 := os.Getenv("POSTGRES2_URL")
	db2 := product_db.InitDB(dns2)

	if err := db2.AutoMigrate(&product.Product{}); err != nil {
		log.Fatalf("Failed to auto-migrate products table: %v", err)
	}

	//init redis and rabbitmq
	redisClient := redis_helper.RedisClient
	rabbitMQCh := rabbitmq_helper.RabbitMQCh
	defer func() {
		if err := rabbitMQCh.Close(); err != nil {
			log.Println("Failed to close RabbitMQ chanel", err)
		}
		log.Println("RabbitMQ channel closed.")
	}()

	orderRepo := repository.NewOrderRepository(db)
	productRepo := product_repository.NewProductRepository(db2)

	orderService := service.NewOrderService(redisClient, rabbitMQCh, orderRepo, productRepo)

	// Initialize and start the OrderWorker
	worker := worker.NewOrderWorker(orderRepo, productRepo)
	workerDone := make(chan struct{})
	numWorker := 1000

	go worker.Start(rabbitMQCh, "orders", numWorker, workerDone)

	http.HandleFunc("/order", handler.OrderHandler(orderService))
	http.HandleFunc("/home", handler.HomeHandler())
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{Addr: ":8000"}
	go func() {
		log.Println("Starting server on port 8000")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan

	log.Println("Server is shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server force  shutdown because: %v", err)
	}

	log.Println("Server down.")
}
