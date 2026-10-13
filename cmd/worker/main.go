package main

import (
	"context"
	"encoding/json"
	"fraud-sentinel/pkg/db"
	"fraud-sentinel/pkg/models"
	"fraud-sentinel/pkg/processor"
	"fraud-sentinel/pkg/queue"
	"log"
	"os"
	"time"
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

	deliveries, err := mq.ConsumeTransactions()
	if err != nil {
		log.Fatalf("failed to consume transactions: %v", err)
	}

	log.Println("[Worker] Worker service running... Waiting for transactions from RabbitMQ.")

	for msg := range deliveries {
		var tx models.Transaction

		if err := json.Unmarshal(msg.Body, &tx); err != nil {
			log.Printf("[Worker] Corrupted message JSON: %v. Dropping message.", err)
			msg.Nack(false, false)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		fraudResult := processor.EvaluateFraudRisk(&tx)

		status := models.StatusApproved
		if fraudResult.Flagged {
			status = models.StatusFlagged
		}

		err = repo.UpdateFraudAssessment(ctx, tx.ID, status, fraudResult)
		if err != nil {
			log.Printf("[Worker] Failed to update status for transaction %s: %v", tx.ID, err)

			msg.Nack(false, true)
			cancel()
			continue
		}

		msg.Ack(false)
		log.Printf("[Worker] Transaction %s processed: Status=%s, RiskScore=%d", tx.ID, status, fraudResult.RiskScore)
		cancel()
	}
}
