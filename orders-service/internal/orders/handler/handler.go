package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/pbb/orders-service/internal/orders/model"
	"github.com/pbb/orders-service/internal/orders/service"
)

func OrderHandler(s *service.OrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var order model.Order
		if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
			log.Printf("Fail to marshall input, Invalid request. %v", err)
			http.Error(w, "Invalid JSON request", http.StatusBadRequest)
			return
		}

		//Gen random id to test
		order.ID = generateOrderID()

		order.CreatedAt = time.Now()
		order.UpdatedAt = order.CreatedAt
		order.Status = model.StatusReceived

		// Save the order to DB.
		if err := s.SaveOrder(&order); err != nil {
			log.Printf("Failed to save order: %v", err)
			http.Error(w, "Failed to save order", http.StatusInternalServerError)
			return
		}

		//Send the order to queue
		if err := s.ProcessOrder(order); err != nil {
			log.Printf("Failed to send order with ID %v because : %v", order.ID, err)
			http.Error(w, "Failed to process order", http.StatusConflict)
			return
		}

		response := model.Response{
			OrderID:    order.ID,
			Status:     string(order.Status),
			StatusCode: 200,
			Message:    "Order received and  being processing by service.",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(response)
	}
}

func generateOrderID() string {
	return uuid.New().String() // Generates a new UUID
}

func HomeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := map[string]string{
			"message": "Welcome to the Orders Service",
			"version": "1.0.0",
			"status":  "running",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
