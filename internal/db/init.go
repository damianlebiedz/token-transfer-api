package db

import (
	"fmt"
	"time"
	"token-transfer-api/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
)

var DB *gorm.DB

func Init() {
	// Initialize PostgreSQL connection using env variables
	user := os.Getenv("POSTGRES_USER")
	pass := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("POSTGRES_HOST")
	db := os.Getenv("POSTGRES_DB")

	if user == "" || pass == "" || host == "" || db == "" {
		log.Fatal("Missing one or more DB connection variables in .env file")
	}

	fmt.Printf("Connecting to DB on %s using user %s", host, user)

	databaseURL := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", user, pass, host, db)

	// Wait 3 seconds to ensure the database container is up before connecting
	time.Sleep(3 * time.Second)

	var err error
	DB, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Cannot connect to database: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Cannot get generic database object: %v", err)
	}

	// Connection pool configuration
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Connected to PostgreSQL with GORM")

	// Automatically migrate the schema for the Wallet model to the database
	err = DB.AutoMigrate(&models.Wallet{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// If not running in the test environment, initialize the database with a default wallet if it doesn't exist
	if os.Getenv("INIT_ENV") != "test" {
		initDefaultWallet("0x0000000000000000000000000000000000000000", 1000000)
	}
}

func initDefaultWallet(Address string, Balance int) {
	var count int64
	DB.Model(&models.Wallet{}).Count(&count)
	log.Printf("Number of wallets in DB: %d\n", count)
	if count == 0 {
		DB.Create(&models.Wallet{Address: Address, Balance: Balance})
		log.Printf("Wallet %s initialized with balance %d", Address, Balance)
	} else {
		log.Printf("Default wallet already initialized")
	}
}
