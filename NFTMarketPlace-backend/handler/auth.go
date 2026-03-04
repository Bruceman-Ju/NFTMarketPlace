package handler

import (
	"net/http"

	"NFTMarketPlace-backend/auth"
	"NFTMarketPlace-backend/models"
	"NFTMarketPlace-backend/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// Login godoc
// @Summary User login via wallet signature
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login request"
// @Success 200 {object} models.AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !utils.IsValidEthAddress(req.Address) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid address"})
		return
	}

	// 验证签名：message = nonce
	valid, err := auth.VerifyEIP191Signature(req.Signature, req.Nonce, req.Address)
	if err != nil || !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}

	token, err := auth.GenerateToken(req.Address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{Token: token})
}

// GetNonce godoc
// @Summary Get login nonce
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string
// @Router /auth/nonce [get]
func (h *AuthHandler) GetNonce(c *gin.Context) {
	// 实际项目中应从 DB 读取用户专属 nonce（防重放）
	// 此处简化为固定值或随机值
	c.JSON(http.StatusOK, gin.H{"nonce": "Sign this message to login to NFT Marketplace"})
}
