package database

import (
	"ecom/backend/models"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() (*gorm.DB, error) {
	// Construct the DSN from environment variables
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("MAIN_DB_HOST"),
		os.Getenv("MAIN_DB_USER"),
		os.Getenv("MAIN_DB_PASSWORD"),
		os.Getenv("MAIN_DB_NAME"),
		os.Getenv("MAIN_DB_PORT"),
		os.Getenv("MAIN_DB_SSL_MODE"),
	)

	log.Printf("Connecting to database with DSN: %s", dsn) // Debug log

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	DB = db
	models := models.GetMigrationModels()

	// Apply migrations for each model
	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			log.Fatalf("Failed to migrate model: %v", err)
		}
	}
	log.Println("Database migrations completed successfully!")

	return db, nil
}
