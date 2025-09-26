package main

import (
	"context"
	"fmt" // Added this missing import
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	queueURL := os.Getenv("SQS_QUEUE_URL")
	if queueURL == "" {
		log.Fatal("SQS_QUEUE_URL must be set")
	}

	// --- START RETRY LOGIC ---
	var dbpool *pgxpool.Pool
	var err error
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		dbConnStr := os.Getenv("DATABASE_URL")
		if dbConnStr == "" {
			log.Fatal("DATABASE_URL must be set")
		}
		dbpool, err = pgxpool.New(context.Background(), dbConnStr)
		if err == nil {
			break // Success!
		}
		log.Printf("Worker failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		log.Fatalf("Worker failed to connect to database after %d attempts: %v", maxRetries, err)
	}
	// --- END RETRY LOGIC ---

	defer dbpool.Close()
	log.Println("Successfully connected to the database.")

	awsCfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("Unable to load AWS SDK config: %v", err)
	}
	sqsClient := sqs.NewFromConfig(awsCfg)

	// --- START of HEALTH CHECK SERVER ---
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Worker is alive and running.")
		})
		port := os.Getenv("PORT")
		if port == "" {
			port = "8081"
		}
		log.Printf("Health check server starting on port %s", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Printf("Health check server failed: %v", err)
		}
	}()
	// --- END of HEALTH CHECK SERVER ---

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
