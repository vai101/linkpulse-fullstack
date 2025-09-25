package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AnalyticsData struct {
	ShortCode  string `json:"short_code"`
	LongURL    string `json:"long_url"`
	ClickCount int    `json:"click_count"`
}

type PostgresStore struct {
	db        *pgxpool.Pool
	sqsClient *sqs.Client
	queueURL  string
}

func NewPostgresStore(dbConnStr, queueURL string) (*PostgresStore, error) {
	dbpool, err := pgxpool.New(context.Background(), dbConnStr)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}
	if err := dbpool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}
	log.Println("Successfully connected to the database.")

	awsCfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}
	log.Println("Successfully loaded AWS configuration.")

	return &PostgresStore{
		db:        dbpool,
		sqsClient: sqs.NewFromConfig(awsCfg),
		queueURL:  queueURL,
	}, nil
}

func (s *PostgresStore) PublishClickEvent(shortCode string) error {
	_, err := s.sqsClient.SendMessage(context.Background(), &sqs.SendMessageInput{
		QueueUrl:    &s.queueURL,
		MessageBody: &shortCode,
	})
	if err != nil {
		return fmt.Errorf("failed to send message to SQS: %w", err)
	}
	log.Printf("Successfully sent click event for %s to SQS.", shortCode)
	return nil
}

func (s *PostgresStore) Save(id uint64, shortCode, longURL string) error {
	query := "INSERT INTO urls (id, short_code, long_url) VALUES ($1, $2, $3)"
	_, err := s.db.Exec(context.Background(), query, id, shortCode, longURL)
	return err
}

func (s *PostgresStore) Load(shortCode string) (string, error) {
	var longURL string
	query := "SELECT long_url FROM urls WHERE short_code = $1"
	err := s.db.QueryRow(context.Background(), query, shortCode).Scan(&longURL)
	if err != nil {
		return "", err
	}
	return longURL, nil
}

func (s *PostgresStore) Close() {
	s.db.Close()
}

func (s *PostgresStore) GetAnalytics() ([]AnalyticsData, error) {
	query := `
		SELECT u.short_code, u.long_url, COUNT(c.id) as click_count
		FROM urls u
		LEFT JOIN clicks c ON u.id = c.url_id
		GROUP BY u.id, u.short_code, u.long_url
		ORDER BY click_count DESC
	`
	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []AnalyticsData
	for rows.Next() {
		var data AnalyticsData
		if err := rows.Scan(&data.ShortCode, &data.LongURL, &data.ClickCount); err != nil {
			return nil, err
		}
		results = append(results, data)
	}
	return results, nil
}

func (s *PostgresStore) GetLastID() (uint64, error) {
	var lastID uint64
	query := "SELECT COALESCE(MAX(id), 0) FROM urls"
	err := s.db.QueryRow(context.Background(), query).Scan(&lastID)
	if err != nil {
		return 0, err
	}
	return lastID, nil
}
