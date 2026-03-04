package models

type SyncState struct {
	ID                 uint   `gorm:"primaryKey;default:1"`
	LastProcessedBlock uint64 `gorm:"not null"`
}

func (m *SyncState) TableName() string { return "sync_state" }
