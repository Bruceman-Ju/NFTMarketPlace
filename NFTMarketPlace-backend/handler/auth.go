package handler

import (
	"NFTMarketPlace-backend/cache"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"NFTMarketPlace-backend/auth"
	"NFTMarketPlace-backend/models"
	"NFTMarketPlace-backend/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	jwtService *auth.JWTService
}

func NewAuthHandler(jwtService *auth.JWTService) *AuthHandler {
	return &AuthHandler{jwtService: jwtService}
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

	token, err := h.jwtService.GenerateToken(req.Address)
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
	address := c.Query("address")

	nonce, _ := generateNonce()

	// 存储到 Redis，5 分钟过期
	cache.NewRedis().Set(fmt.Sprintf("nonce:%s", address), nonce, 5*time.Minute)

	c.JSON(http.StatusOK, gin.H{"nonce": nonce})
}

func generateNonce() (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
	randomPart := hex.EncodeToString(randomBytes)
	return fmt.Sprintf("%s-%s", timestamp, randomPart), nil
}
