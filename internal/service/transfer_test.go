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
	db.Init()

	// Clear existing wallet data
	err := db.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Wallet{}).Error
	require.NoError(t, err)
}

func TestTransfer_Success(t *testing.T) {
	setupTest(t)

	require.NoError(t, db.DB.Create(&models.Wallet{Address: "A", Balance: 10}).Error)

	newBalance, err := service.Transfer("A", "B", 10)

	require.NoError(t, err)
	require.Equal(t, 0, newBalance)
}

func TestTransfer_InsufficientBalance(t *testing.T) {
	setupTest(t)

	require.NoError(t, db.DB.Create(&models.Wallet{Address: "A", Balance: 10}).Error)

	_, err := service.Transfer("A", "B", 20)

	require.Error(t, err)
	require.Contains(t, err.Error(), "insufficient balance")
}

func TestTransfer_WalletNotFound(t *testing.T) {
	setupTest(t)

	require.NoError(t, db.DB.Create(&models.Wallet{Address: "A", Balance: 10}).Error)

	_, err := service.Transfer("B", "A", 10)

	require.Error(t, err)
	require.Contains(t, err.Error(), "sender wallet not found")
}

func TestTransfer_ConcurrentTransactionHandling(t *testing.T) {
	setupTest(t)

	require.NoError(t, db.DB.Create(&models.Wallet{Address: "A", Balance: 10}).Error)

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

	t.Logf("Final balance of wallet A: %d, should be 7, 4 or 0", walletA.Balance)
}

func TestTransfer_ConcurrentReceiverCreation(t *testing.T) {
	setupTest(t)

	require.NoError(t, db.DB.Create(&models.Wallet{Address: "A", Balance: 10}).Error)

	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		<-start
		_, _ = service.Transfer("A", "C", 5)
	}()

	go func() {
		defer wg.Done()
		<-start
		_, _ = service.Transfer("A", "C", 5)
	}()

	close(start)
	wg.Wait()

	var walletC models.Wallet
	err := db.DB.First(&walletC, "address = ?", "C").Error
	require.NoError(t, err)
	require.Equal(t, 10, walletC.Balance)

	var walletA models.Wallet
	err = db.DB.First(&walletA, "address = ?", "A").Error
	require.NoError(t, err)
	require.Equal(t, 0, walletA.Balance)

	t.Logf("Final balances: C=%d, A=%d; expected: C=10, A=0", walletC.Balance, walletA.Balance)

}

func TestTransfer_ConcurrentDeadlock(t *testing.T) {
	setupTest(t)

	require.NoError(t, db.DB.Create(&models.Wallet{Address: "A", Balance: 100}).Error)
	require.NoError(t, db.DB.Create(&models.Wallet{Address: "B", Balance: 100}).Error)

	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		<-start
		_, err := service.Transfer("A", "B", 30)
		require.NoError(t, err)
	}()

	go func() {
		defer wg.Done()
		<-start
		_, err := service.Transfer("B", "A", 50)
		require.NoError(t, err)
	}()

	close(start)
	wg.Wait()

	var walletA, walletB models.Wallet
	require.NoError(t, db.DB.First(&walletA, "address = ?", "A").Error)
	require.NoError(t, db.DB.First(&walletB, "address = ?", "B").Error)

	total := walletA.Balance + walletB.Balance

	require.Equal(t, 200, total)

	t.Logf("Final balances: A=%d, B=%d; expected A+B=200", walletA.Balance, walletB.Balance)
}

func TestTransfer_Foo(t *testing.T) {
	setupTest(t)

	require.NoError(t, db.DB.Create(&models.Wallet{Address: "A", Balance: 1000}).Error)
	require.NoError(t, db.DB.Create(&models.Wallet{Address: "B", Balance: 1000}).Error)
	require.NoError(t, db.DB.Create(&models.Wallet{Address: "C", Balance: 0}).Error)

	var wg sync.WaitGroup
	wg.Add(2000)

	for i := 0; i < 1000; i++ {
		go func() {
			defer wg.Done()
			_, err := service.Transfer("A", "C", 1)
			require.NoError(t, err)
		}()

		go func() {
			defer wg.Done()
			_, err := service.Transfer("B", "C", 1)
			require.NoError(t, err)
		}()
	}

	wg.Wait()

	var walletC models.Wallet
	err := db.DB.First(&walletC, "address = ?", "C").Error
	require.NoError(t, err)

	require.Equal(t, 2000, walletC.Balance)

	t.Logf("expected: 2000 , actual: %d", walletC.Balance)
}
