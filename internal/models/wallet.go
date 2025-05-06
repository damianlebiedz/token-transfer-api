package models

type Wallet struct {
	Address string `gorm:"primaryKey"`
	Balance int
}
