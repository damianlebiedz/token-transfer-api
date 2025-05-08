package service_test

import (
	"github.com/damianlebiedz/token-transfer-api/internal/db"
	"github.com/damianlebiedz/token-transfer-api/internal/models"
	"github.com/damianlebiedz/token-transfer-api/internal/service"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"sync"
	"testing"
	_ "time"

	_ "github.com/stretchr/testify/assert"
)

func setupTest(t *testing.T) {
	// Initialize test DB
	db.Init()

	// Clear existing wallet data
	err := db.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Wallet{}).Error
	require.NoError(t, err)

	// Create an initial test wallet
	err = db.DB.Create(&models.Wallet{
		Address: "A",
		Balance: 10,
	}).Error
	require.NoError(t, err)
}

// Test successful transfer and check if balance is updated correctly
func TestTransfer_Success(t *testing.T) {
	setupTest(t)

	newBalance, err := service.Transfer("A", "B", 10)

	require.NoError(t, err)
	require.Equal(t, 0, newBalance)
}

// Test transfer with insufficient balance
func TestTransfer_InsufficientBalance(t *testing.T) {
	setupTest(t)

	_, err := service.Transfer("A", "B", 20)

	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient balance")
}

// Test transfer when sender wallet does not exist
func TestTransfer_WalletNotFound(t *testing.T) {
	setupTest(t)

	_, err := service.Transfer("B", "A", 10)

	require.Error(t, err)
	require.Contains(t, err.Error(), "sender wallet not found")
}

// Test handling of race condition during concurrent transfers
func TestTransfer_RaceCondition(t *testing.T) {
	setupTest(t)

	err := db.DB.Create(&models.Wallet{
		Address: "B",
		Balance: 10,
	}).Error
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		_, _ = service.Transfer("B", "A", 1)
	}()

	go func() {
		defer wg.Done()
		_, _ = service.Transfer("A", "B", 4)
	}()

	go func() {
		defer wg.Done()
		_, _ = service.Transfer("A", "B", 7)
	}()

	wg.Wait()
}
