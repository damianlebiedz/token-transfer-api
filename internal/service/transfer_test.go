package service_test

import (
	"gorm.io/gorm"
	"sync"
	"testing"
	_ "time"

	"github.com/damianlebiedz/token-transfer-api/internal/db"
	"github.com/damianlebiedz/token-transfer-api/internal/models"
	"github.com/damianlebiedz/token-transfer-api/internal/service"
	_ "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) {
	db.Init()

	err := db.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Wallet{}).Error
	require.NoError(t, err)

	err = db.DB.Create(&models.Wallet{
		Address: "A",
		Balance: 10,
	}).Error
	require.NoError(t, err)
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
