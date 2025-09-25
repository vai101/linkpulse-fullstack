package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	dbConnStr := os.Getenv("DATABASE_URL")
	if dbConnStr == "" {
		log.Fatal("DATABASE_URL must be set")
	}

	// Add debug logging to see what the worker is actually getting
	log.Printf("DEBUG WORKER - DATABASE_URL: %s", dbConnStr)
	log.Printf("DEBUG WORKER - Length: %d", len(dbConnStr))

	queueURL := os.Getenv("SQS_QUEUE_URL")
	if queueURL == "" {
		log.Fatal("SQS_QUEUE_URL must be set")
	}

	log.Printf("DEBUG WORKER - SQS_QUEUE_URL: %s", queueURL)

	dbpool, err := pgxpool.New(context.Background(), dbConnStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbpool.Close()
	log.Println("Successfully connected to the database.")

	awsCfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("Unable to load AWS SDK config: %v", err)
	}
	sqsClient := sqs.NewFromConfig(awsCfg)

	log.Println("Worker started. Listening for messages...")

	for {
		receiveOutput, err := sqsClient.ReceiveMessage(context.Background(), &sqs.ReceiveMessageInput{
			QueueUrl:            &queueURL,
			MaxNumberOfMessages: 10,
			WaitTimeSeconds:     20,
		})
		if err != nil {
			log.Printf("Error receiving message: %v", err)
			continue
		}

		if len(receiveOutput.Messages) == 0 {
			continue
		}

		log.Printf("Received %d message(s).", len(receiveOutput.Messages))

		for _, msg := range receiveOutput.Messages {
			shortCode := *msg.Body
			log.Printf("Processing click for short code: %s", shortCode)

			var urlID int64
			err := dbpool.QueryRow(context.Background(), "SELECT id FROM urls WHERE short_code = $1", shortCode).Scan(&urlID)
			if err != nil {
				log.Printf("Error finding url_id for short_code %s: %v", shortCode, err)
				continue
			}

			_, err = dbpool.Exec(context.Background(), "INSERT INTO clicks (url_id) VALUES ($1)", urlID)
			if err != nil {
				log.Printf("Error saving click to database: %v", err)
				continue
			}

			log.Printf("Successfully saved click for short_code %s.", shortCode)

			_, err = sqsClient.DeleteMessage(context.Background(), &sqs.DeleteMessageInput{
				QueueUrl:      &queueURL,
				ReceiptHandle: msg.ReceiptHandle,
			})
			if err != nil {
				log.Printf("Error deleting message: %v", err)
			}
		}
	}
}
