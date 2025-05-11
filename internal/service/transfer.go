package service

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"token-transfer-api/internal/db"
	"token-transfer-api/internal/models"
)

// Transfer the tokens between wallets
func Transfer(from string, to string, amount int) (int, error) {
	if amount <= 0 {
		return 0, errors.New("transfer amount must be greater than 0")
	}

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var sender models.Wallet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&sender, "address = ?", from).Error; err != nil {
			return fmt.Errorf("sender wallet not found: %w", err)
		}
		if sender.Balance < amount {
			return fmt.Errorf("sender has insufficient balance: required %d, available %d", amount, sender.Balance)
		}

		var receiver models.Wallet
		if err := tx.First(&receiver, "address = ?", to).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// If the receiver doesn't exist, initialize a new wallet with 0 balance
				receiver = models.Wallet{
					Address: to,
					Balance: 0,
				}
				if err := tx.Create(&receiver).Error; err != nil {
					return fmt.Errorf("failed to create receiver wallet: %w", err)
				} else {
					log.Printf("initialized wallet: %s with balance: 0", to)
				}
			} else {
				return fmt.Errorf("failed to get receiver wallet: %w", err)
			}
		}

		// Perform the transfer
		sender.Balance -= amount
		receiver.Balance += amount

		// Update the balances
		if err := tx.Save(&sender).Error; err != nil {
			return fmt.Errorf("failed to update sender: %w", err)
		}
		if err := tx.Save(&receiver).Error; err != nil {
			return fmt.Errorf("failed to update receiver: %w", err)
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	// Load updated sender wallet
	var updatedSender models.Wallet
	if err := db.DB.First(&updatedSender, "address = ?", from).Error; err != nil {
		return 0, fmt.Errorf("failed to load updated sender: %w", err)
	}

	return updatedSender.Balance, nil
}
