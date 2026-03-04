package repository

import (
	"NFTMarketPlace-backend/models"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Listing
func (r *Repository) CreateListing(nft *models.ListedNFT) error {
	return r.db.Create(nft).Error
}

func (r *Repository) MarkListingAsSold(listID string) error {
	return r.db.Model(&models.ListedNFT{}).
		Where("list_id = ?", listID).
		Update("status", "sold").Error
}

func (r *Repository) MarkListingAsInactive(listID string, status string) error {
	return r.db.Model(&models.ListedNFT{}).
		Where("list_id = ?", listID).
		Update("status", status).Error
}

func (r *Repository) GetActiveListings(page, size int) ([]models.ListedNFT, error) {
	var list []models.ListedNFT
	offset := (page - 1) * size
	err := r.db.Where("status = ?", "listed").
		Order("listed_time DESC").
		Limit(size).Offset(offset).
		Find(&list).Error
	return list, err
}

func (r *Repository) GetUserListings(address string, page, size int) ([]models.ListedNFT, error) {
	var list []models.ListedNFT
	offset := (page - 1) * size
	err := r.db.Where("seller = ? AND status = ?", address, "listed").
		Order("listed_time DESC").
		Limit(size).Offset(offset).
		Find(&list).Error
	return list, err
}

// Sale
func (r *Repository) CreateSale(sale *models.Sale) error {
	return r.db.Create(sale).Error
}

func (r *Repository) GetUserSales(address string, page, size int) ([]models.Sale, error) {
	var sales []models.Sale
	offset := (page - 1) * size
	err := r.db.Where("seller = ?", address).
		Order("sold_time DESC").
		Limit(size).Offset(offset).
		Find(&sales).Error
	return sales, err
}

func (r *Repository) GetBuyerPurchases(address string, page, size int) ([]models.Sale, error) {
	var sales []models.Sale
	offset := (page - 1) * size
	err := r.db.Where("buyer = ?", address).
		Order("sold_time DESC").
		Limit(size).Offset(offset).
		Find(&sales).Error
	return sales, err
}

// Sync
func (r *Repository) GetLastSyncBlock() uint64 {
	var state models.SyncState
	r.db.FirstOrCreate(&state, models.SyncState{ID: 1})
	return state.LastProcessedBlock
}

func (r *Repository) UpdateSyncBlock(block uint64) {
	r.db.Model(&models.SyncState{}).Where("id = 1").Update("last_processed_block", block)
}
