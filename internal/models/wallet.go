package models

// Wallet model with Address as primary key and integer Balance
type Wallet struct {
	Address string `gorm:"primaryKey"`
	Balance int
}
