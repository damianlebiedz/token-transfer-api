package db

import (
	"fmt"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
	"testing"
)

var TestDB *gorm.DB

func InitTest(t *testing.T) (*gorm.DB, func()) {
	t.Helper()

	err := godotenv.Load(".env_test")
	if err != nil {
		log.Fatalf("Error loading .env_test file")
	}

	user := os.Getenv("POSTGRES_USER")
	pass := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	db := os.Getenv("POSTGRES_DB")

	if user == "" || pass == "" || host == "" || port == "" || db == "" {
		log.Fatal("Missing one or more DB connection variables in .env_test")
	}

	fmt.Printf("Connecting to DB with user %s, host %s...\n", user, host)

	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, db)

	TestDB, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test DB: %v", err)
	}

	return TestDB, func() {
		sqlDB, _ := TestDB.DB()
		err := sqlDB.Close()
		if err != nil {
			return
		}
	}
}
