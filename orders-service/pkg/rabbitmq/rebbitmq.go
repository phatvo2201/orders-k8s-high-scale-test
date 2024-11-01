package rabbitmq_helper

import (
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
)

var (
	RabbitMQConn *amqp.Connection
	RabbitMQCh   *amqp.Channel
)

func init() {
	var err error
	maxRetrieNum := 5
	retryInterval := 15 * time.Second

	for i := 1; i <= maxRetrieNum; i++ {
		rabbitMQURL := os.Getenv("RABBITMQ_URL")

		RabbitMQConn, err = amqp.Dial(rabbitMQURL)
		if err == nil {
			break
		}
		log.Printf("Rabbitmq connect fail , Try to retry ... (%d/%d)", i, maxRetrieNum)
		time.Sleep(retryInterval)
	}

	if err != nil {
		log.Fatalf("Rabbitmq connect fail , %d try to retry but err : %v", maxRetrieNum, err)
	}

	RabbitMQCh, err = RabbitMQConn.Channel()
	if err != nil {
		log.Fatalf("Failed to open Rabbitmq channel: %v", err)
	}

	DeclareOrderQueue("orders")
	log.Println("RabbitMQ connected and queue declared successfully")
}

func DeclareOrderQueue(queueName string) {
	_, err := RabbitMQCh.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare RabbitMQ queue: %v", err)
	}
	log.Printf("Queue '%s' declared successfully", queueName)
}
