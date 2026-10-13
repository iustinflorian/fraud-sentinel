package main

import (
	"fraud-sentinel/pkg/db"
	"fraud-sentinel/pkg/queue"
	"log"
	"net/http"
	"os"
)

func main() {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017" // Fallback for local `go run` execution
	}

	rabbitURI := os.Getenv("RABBITMQ_URI")
	if rabbitURI == "" {
		rabbitURI = "amqp://guest:guest@localhost:5672/" // Fallback for local `go run` execution
	}

	database, err := db.ConnectMongo(mongoURI, "fraud-sentinel")
	if err != nil {
		log.Fatalf("error connecting to MongoDB: %v", err)
	}

	repo := db.NewTransactionRepository(database)
	log.Println("Connection with MongoDB established!")

	mq, err := queue.NewRabbitMQ(rabbitURI)
	if err != nil {
		log.Fatalf("error connecting to RabbitMQ: %v", err)
	}

	log.Println("Connection with RabbitMQ established!")
	defer mq.Conn.Close()
	defer mq.Channel.Close()

	txHandler := NewTransactionHandler(repo, mq)

	http.HandleFunc("/api/v1/transactions", txHandler.CreateTransactionHandler)
	http.HandleFunc("/api/v1/transactions/status", txHandler.GetTransactionStatusHandler)

	log.Println("API Server runs on http://localhost:8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("error starting server: %v", err)
	}
}
