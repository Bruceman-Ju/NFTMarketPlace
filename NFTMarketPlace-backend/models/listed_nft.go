package models

type ListedNFT struct {
	ID         uint   `gorm:"primaryKey"`
	ListID     string `gorm:"uniqueIndex;not null"` // hex of bytes32
	NFTAddress string `gorm:"not null;index"`
	TokenID    string `gorm:"not null;index:idx_nft"`
	Price      string `gorm:"not null"` // wei as string
	ListedTime int64  `gorm:"not null"`
	Seller     string `gorm:"not null;index"`
	ExpiredAt  int64  `gorm:"not null;index"`
	Status     string `gorm:"default:listed;index"` // listed, sold, canceled, expired
}

func (m *ListedNFT) TableName() string { return "listed_nfts" }
