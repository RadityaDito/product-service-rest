package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Config holds database configuration
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewConnection creates a new database connection with pooling
func NewConnection() *sqlx.DB {
	// Read configuration from environment variables
	config := Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "productuser"),
		Password: getEnv("DB_PASSWORD", "productpass"),
		DBName:   getEnv("DB_NAME", "productdb"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	// Create connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

	// Attempt connection with retries
	var db *sqlx.DB
	var err error
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		db, err = sqlx.Open("postgres", connStr)
		if err != nil {
			log.Printf("Error opening database: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// Configure connection pool - SAME AS GRPC FOR FAIR COMPARISON
		db.SetMaxOpenConns(25)                  // Max open connections
		db.SetMaxIdleConns(5)                   // Max idle connections
		db.SetConnMaxLifetime(30 * time.Minute) // Connection lifetime

		// Test the connection
		err = db.Ping()
		if err == nil {
			log.Println("Successfully connected to the database")
			return db
		}

		log.Printf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(5 * time.Second)
	}

	log.Fatalf("Could not connect to database after %d attempts", maxRetries)
	return nil
}

// getEnv retrieves environment variable with a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// InitSchema creates the products table if it doesn't exist
func InitSchema(db *sqlx.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS products (
		id UUID PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		price DECIMAL(10,2) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_product_name ON products(name);
	CREATE INDEX IF NOT EXISTS idx_product_price ON products(price);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("error creating schema: %v", err)
	}

	return nil
}
