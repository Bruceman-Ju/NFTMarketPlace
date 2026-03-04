package handler

import (
	"net/http"

	"NFTMarketPlace-backend/repository"
	"NFTMarketPlace-backend/utils"

	"github.com/gin-gonic/gin"
)

type ListingHandler struct {
	repo *repository.Repository
}

func NewListingHandler(repo *repository.Repository) *ListingHandler {
	return &ListingHandler{repo: repo}
}

// GetAllListings godoc
// @Summary Get all active listings
// @Tags listings
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param size query int false "Page size" default(20)
// @Success 200 {array} models.ListedNFT
// @Router /listings [get]
func (h *ListingHandler) GetAllListings(c *gin.Context) {
	page, size := utils.GetPagination(c, 1, 20)
	listings, err := h.repo.GetActiveListings(page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}
	c.JSON(http.StatusOK, listings)
}

func (h *ListingHandler) GetUserListings(c *gin.Context) {
	address := c.Param("address")
	if !utils.IsValidEthAddress(address) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid address"})
		return
	}
	page, size := utils.GetPagination(c, 1, 20)
	listings, err := h.repo.GetUserListings(address, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}
	c.JSON(http.StatusOK, listings)
}
