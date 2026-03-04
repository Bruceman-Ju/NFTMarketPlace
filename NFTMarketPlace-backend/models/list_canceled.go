package models

type ListCanceled struct {
	ListID     string `gorm:"uniqueIndex;not null"` // hex of bytes32
	NFTAddress string `gorm:"not null;index"`
	TokenID    string `gorm:"not null;index:idx_nft"`
	CancelTime int64  `gorm:"not null"`
	Operator   string `gorm:"not null;index"`
}
