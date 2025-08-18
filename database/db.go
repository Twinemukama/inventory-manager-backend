package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Twinemukama/go-inventory-manager/models" // Adjust the import path as necessary
	"github.com/joho/godotenv"
)

var DB *gorm.DB

func InitDB() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found")
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Africa/Kampala",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("Database connection established.")

	// Auto migrate models
	err = DB.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Item{},
		&models.Transaction{},
	)
	if err != nil {
		log.Fatal("Failed to auto-migrate models:", err)
	}

	fmt.Println("Database migrated successfully.")
}
