package service

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
	"token-transfer-api/internal/models"
)

// Transfer the tokens between wallets
func Transfer(db *gorm.DB, from string, to string, amount int) (int, error) {
	if amount <= 0 {
		return 0, errors.New("transfer amount must be greater than 0")
	}

	var updatedBalance int

	// Determine the order of addresses for locking in alphabetical order to avoid deadlocks
	err := db.Transaction(func(tx *gorm.DB) error {
		var firstAddr, secondAddr string
		if from < to {
			firstAddr, secondAddr = from, to
		} else {
			firstAddr, secondAddr = to, from
		}

		var firstWallet, secondWallet models.Wallet

		// Lock first wallet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&firstWallet, "address = ?", firstAddr).Error; err != nil {
			if firstAddr == from {
				return fmt.Errorf("sender wallet not found: %w", err)
			} else {
				return fmt.Errorf("receiver wallet not found: %w", err)
			}
		}

		// Lock second wallet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&secondWallet, "address = ?", secondAddr).Error; err != nil {
			// If the receiver doesn't exist, initialize a new wallet with 0 balance
			if errors.Is(err, gorm.ErrRecordNotFound) && secondAddr == to {
				secondWallet = models.Wallet{
					Address: to,
					Balance: 0,
				}
				if err := tx.Create(&secondWallet).Error; err != nil {
					if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "UNIQUE constraint failed") {
						// Someone else created it - load again
						if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&secondWallet, "address = ?", to).Error; err != nil {
							return fmt.Errorf("failed to re-load receiver after conflict: %w", err)
						}
					} else {
						return fmt.Errorf("failed to create receiver wallet: %w", err)
					}
				}
			} else {
				if secondAddr == from {
					return fmt.Errorf("sender wallet not found: %w", err)
				} else {
					return fmt.Errorf("receiver wallet not found: %w", err)
				}
			}
		}

		var sender, receiver *models.Wallet
		if from == firstWallet.Address {
			sender, receiver = &firstWallet, &secondWallet
		} else {
			sender, receiver = &secondWallet, &firstWallet
		}

		if sender.Balance < amount {
			return fmt.Errorf("sender has insufficient balance: required %d, available %d", amount, sender.Balance)
		}

		// Perform the transfer
		sender.Balance -= amount
		receiver.Balance += amount

		if err := tx.Save(sender).Error; err != nil {
			return fmt.Errorf("failed to update sender: %w", err)
		}
		if err := tx.Save(receiver).Error; err != nil {
			return fmt.Errorf("failed to update receiver: %w", err)
		}

		updatedBalance = sender.Balance

		return nil
	})

	if err != nil {
		return 0, err
	}

	return updatedBalance, nil
}
