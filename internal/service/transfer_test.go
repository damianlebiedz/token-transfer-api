package service_test

import (
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"sync"
	"testing"
	_ "time"
	"token-transfer-api/internal/db"
	"token-transfer-api/internal/models"
	"token-transfer-api/internal/service"

	_ "github.com/stretchr/testify/assert"
)

func setupTest(t *testing.T) {
	// Initialize test DB
	db.Init()

	// Clear existing wallet data
	err := db.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Wallet{}).Error
	require.NoError(t, err)

	// Create an initial test wallet
	require.NoError(t, db.DB.Create(&models.Wallet{
		Address: "A",
		Balance: 10,
	}).Error)
}

func TestTransfer_Success(t *testing.T) {
	setupTest(t)

	newBalance, err := service.Transfer("A", "B", 10)

	require.NoError(t, err)
	require.Equal(t, 0, newBalance)
}

func TestTransfer_InsufficientBalance(t *testing.T) {
	setupTest(t)

	_, err := service.Transfer("A", "B", 20)

	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient balance")
}

func TestTransfer_WalletNotFound(t *testing.T) {
	setupTest(t)

	_, err := service.Transfer("B", "A", 10)

	require.Error(t, err)
	require.Contains(t, err.Error(), "sender wallet not found")
}

func TestTransfer_ConcurrentTransactionHandling(t *testing.T) {
	setupTest(t)

	err := db.DB.Create(&models.Wallet{Address: "B", Balance: 10}).Error
	require.NoError(t, err)

	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		<-start
		_, _ = service.Transfer("B", "A", 1)
	}()

	go func() {
		defer wg.Done()
		<-start
		_, _ = service.Transfer("A", "B", 4)
	}()

	go func() {
		defer wg.Done()
		<-start
		_, _ = service.Transfer("A", "B", 7)
	}()

	close(start)
	wg.Wait()

	var walletA models.Wallet

	err = db.DB.First(&walletA, "address = ?", "A").Error
	require.NoError(t, err)

	// Possible outcomes:
	// -4 accepted, -7 rejected, +1 accepted -> Balance = 7
	// -7 accepted, -4 rejected, +1 accepted -> Balance = 4
	// -4, +1 and -7 accepted -> Balance = 0
	validBalances := map[int]bool{7: true, 4: true, 0: true}
	require.True(t, validBalances[walletA.Balance], "Invalid final balance of wallet A: %d", walletA.Balance)

	t.Logf("Final balance of wallet A: %d", walletA.Balance)
}
