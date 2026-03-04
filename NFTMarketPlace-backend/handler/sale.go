package handler

import (
	"net/http"

	"NFTMarketPlace-backend/repository"
	"NFTMarketPlace-backend/utils"

	"github.com/gin-gonic/gin"
)

type SaleHandler struct {
	repo *repository.Repository
}

func NewSaleHandler(repo *repository.Repository) *SaleHandler {
	return &SaleHandler{repo: repo}
}

func (h *SaleHandler) GetUserSales(c *gin.Context) {
	address := c.Param("address")
	if !utils.IsValidEthAddress(address) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid address"})
		return
	}
	page, size := utils.GetPagination(c, 1, 20)
	sales, err := h.repo.GetUserSales(address, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}
	c.JSON(http.StatusOK, sales)
}

func (h *SaleHandler) GetBuyerPurchases(c *gin.Context) {
	address := c.Param("address")
	if !utils.IsValidEthAddress(address) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid address"})
		return
	}
	page, size := utils.GetPagination(c, 1, 20)
	sales, err := h.repo.GetBuyerPurchases(address, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}
	c.JSON(http.StatusOK, sales)
}
