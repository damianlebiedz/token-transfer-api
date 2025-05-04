package db

import (
	"fmt"
	"github.com/damianlebiedz/token-transfer-api/internal/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
)

var DB *gorm.DB

func Init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	user := os.Getenv("POSTGRES_USER")
	pass := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	db := os.Getenv("POSTGRES_DB")

	if user == "" || pass == "" || host == "" || port == "" || db == "" {
		log.Fatal("Missing one or more DB connection variables in .env")
	}

	fmt.Printf("Connecting to DB with user %s, host %s...\n", user, host)

	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, db)

	DB, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Cannot connect to database: %v", err)
	}

	log.Println("Connected to PostgreSQL with GORM")

	err = DB.AutoMigrate(&models.Wallet{})
	if err != nil {
		return
	}

	if os.Getenv("INIT_ENV") != "test" {
		var count int64
		DB.Model(&models.Wallet{}).Count(&count)
		if count == 0 {
			DB.Create(&models.Wallet{Address: "0x0000000000000000000000000000000000000000", Balance: 1000000})
		}
	}
}
