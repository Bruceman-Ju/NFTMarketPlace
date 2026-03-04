package models

type Sale struct {
	ID         uint   `gorm:"primaryKey"`
	ListID     string `gorm:"not null"`
	NFTAddress string `gorm:"not null"`
	TokenID    string `gorm:"not null"`
	Price      string `gorm:"not null"`
	SoldTime   int64  `gorm:"not null"`
	Seller     string `gorm:"not null"`
	Buyer      string `gorm:"not null;index"`
}

func (m *Sale) TableName() string { return "nft_sales" }
