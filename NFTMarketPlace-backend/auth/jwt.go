package auth

import (
	"time"

	"NFTMarketPlace-backend/config"
	"github.com/golang-jwt/jwt/v4"
)

func GenerateToken(address string) (string, error) {
	claims := jwt.MapClaims{
		"address": address,
		"exp":     time.Now().Add(time.Hour * time.Duration(config.Cfg.JWT.ExpireHours)).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Cfg.JWT.Secret))
}
