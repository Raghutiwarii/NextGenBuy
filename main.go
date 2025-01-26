package main

import (
	"ecom/backend/models"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Get environment variables
	serverPort := os.Getenv("SERVER_PORT")
	dbHost := os.Getenv("MAIN_DB_HOST")
	dbName := os.Getenv("MAIN_DB_NAME")
	dbUser := os.Getenv("MAIN_DB_USER")
	dbPassword := os.Getenv("MAIN_DB_PASSWORD")
	dbPort := os.Getenv("MAIN_DB_PORT")
	dbSSLMode := os.Getenv("MAIN_DB_SSL_MODE")

	// Database connection string
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)

	// Connect to the database using GORM
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Set log level
	})
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	// Perform migrations
	migrateModels(db)

	// Set up a simple HTTP server
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Service is up and running!"))
	})

	banner := `
	                                          
	  ,------.  ,-----.  ,-----.  ,--.   ,--. 
	  |  .---' '  .--./ '  .-.  ' |   '.'   | 
	  |  '--,  |  |     |  | |  | |  |'.'|  | 
	  |  '--.' '  '--'\ '  '-'  ' |  |   |  | 
	  '------'  '-----'  '-----'  '--'   '--' 
	`

	log.Println(banner)

	// Start the server
	log.Printf("Server is running on port %s", serverPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", serverPort), nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

// migrateModels performs database migrations
func migrateModels(db *gorm.DB) {
	// Get all models for migration
	models := models.GetMigrationModels()

	// Apply migrations for each model
	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			log.Fatalf("Failed to migrate model: %v", err)
		}
	}
	log.Println("Database migrations completed successfully!")
}
