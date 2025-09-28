package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/vai101/linkpulse/shortener"
	"github.com/vai101/linkpulse/storage"
)

type App struct {
	store     *storage.PostgresStore
	counter   uint64
	counterMu sync.Mutex
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}

	dbConnStr := os.Getenv("DATABASE_URL")
	if dbConnStr == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}
	queueURL := os.Getenv("SQS_QUEUE_URL")
	if queueURL == "" {
		log.Fatal("SQS_QUEUE_URL environment variable is not set")
	}

	var store *storage.PostgresStore
	var err error
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		// This is the corrected line with both arguments
		store, err = storage.NewPostgresStore(dbConnStr, queueURL)
		if err == nil {
			break // Success!
		}
		log.Printf("Failed to connect to store (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to connect to store after %d attempts: %v", maxRetries, err)
	}

	defer store.Close()

	lastID, err := store.GetLastID()
	if err != nil {
		log.Fatalf("Failed to get last ID from database: %v", err)
	}

	app := &App{
		store:   store,
		counter: lastID,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", app.rootHandler)
	mux.HandleFunc("/api/analytics", app.handleGetAnalytics)

	fmt.Printf("Starting server on :8080... (last known ID is %d)\n", app.counter)
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func (app *App) handleGetAnalytics(w http.ResponseWriter, r *http.Request) {
	// Change this from GET to POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	analytics, err := app.store.GetAnalytics()
	if err != nil {
		log.Printf("Error fetching analytics: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	json.NewEncoder(w).Encode(analytics)
}
func (app *App) rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.URL.Path == "/" {
		if r.Method == http.MethodPost {
			app.handleShorten(w, r)
		} else {
			http.Error(w, "Please POST to this endpoint to shorten a URL.", http.StatusMethodNotAllowed)
		}
		return
	}
	app.handleRedirect(w, r)
}
func (app *App) handleShorten(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	app.counterMu.Lock()
	app.counter++
	id := app.counter
	app.counterMu.Unlock()
	shortCode := shortener.Base62Encode(id)
	if err := app.store.Save(id, shortCode, req.URL); err != nil {
		log.Printf("Failed to save URL: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get the base URL from an environment variable
	apiBaseUrl := os.Getenv("API_BASE_URL")
	if apiBaseUrl == "" {
		apiBaseUrl = "http://localhost:8080" // A fallback for local dev
	}

	fullShortURL := fmt.Sprintf("%s/%s", apiBaseUrl, shortCode)

	res := struct {
		ShortURL string `json:"short_url"`
	}{ShortURL: fullShortURL}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
func (app *App) handleRedirect(w http.ResponseWriter, r *http.Request) {
	shortCode := r.URL.Path[1:]
	longURL, err := app.store.Load(shortCode)
	if err != nil {
		log.Printf("Failed to load URL for short code %s: %v", shortCode, err)
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	go func() {
		if err := app.store.PublishClickEvent(shortCode); err != nil {
			log.Printf("Error publishing click event: %v", err)
		}
	}()
	http.Redirect(w, r, longURL, http.StatusFound)
}
