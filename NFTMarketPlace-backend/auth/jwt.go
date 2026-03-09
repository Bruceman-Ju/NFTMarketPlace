package auth

import (
	"NFTMarketPlace-backend/config"
	"crypto/ecdsa"
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrUnsupportedAlgorithm = errors.New("unsupported JWT algorithm")
	ErrKeyNotConfigured     = errors.New("JWT key not configured for selected algorithm")
)

type JWTService struct {
	algorithm   string
	secret      []byte
	privateKey  interface{} // *rsa.PrivateKey or *ecdsa.PrivateKey
	publicKey   interface{} // *rsa.PublicKey or *ecdsa.PublicKey
	expireHours int
}

func NewJWTService(cfg config.Config) (*JWTService, error) {
	s := &JWTService{
		algorithm:   cfg.JWT.Algorithm,
		expireHours: cfg.JWT.ExpireHours,
	}

	switch cfg.JWT.Algorithm {
	case "HS256":
		if cfg.JWT.Secret == "" {
			return nil, fmt.Errorf("%w: secret is empty", ErrKeyNotConfigured)
		}
		s.secret = []byte(cfg.JWT.Secret)

	case "RS256":
		if cfg.JWT.PrivateKeyPath == "" || cfg.JWT.PublicKeyPath == "" {
			return nil, fmt.Errorf("%w: RSA key paths not set", ErrKeyNotConfigured)
		}
		priv, pub, err := loadRSAKeys(cfg.JWT.PrivateKeyPath, cfg.JWT.PublicKeyPath)
		if err != nil {
			return nil, err
		}
		s.privateKey = priv
		s.publicKey = pub

	case "ES256":
		if cfg.JWT.PrivateKeyPath == "" || cfg.JWT.PublicKeyPath == "" {
			return nil, fmt.Errorf("%w: ECDSA key paths not set", ErrKeyNotConfigured)
		}
		priv, pub, err := loadECDSAKeys(cfg.JWT.PrivateKeyPath, cfg.JWT.PublicKeyPath)
		if err != nil {
			return nil, err
		}
		s.privateKey = priv
		s.publicKey = pub

	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedAlgorithm, cfg.JWT.Algorithm)
	}

	return s, nil
}

// GenerateToken 生成 JWT
func (s *JWTService) GenerateToken(address string) (string, error) {
	claims := jwt.MapClaims{
		"address": address,
		"exp":     time.Now().Add(time.Hour * time.Duration(s.expireHours)).Unix(),
		"iat":     time.Now().Unix(),
	}

	var token *jwt.Token
	switch s.algorithm {
	case "HS256":
		token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		return token.SignedString(s.secret)

	case "RS256":
		token = jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		return token.SignedString(s.privateKey.(*rsa.PrivateKey))

	case "ES256":
		token = jwt.NewWithClaims(jwt.SigningMethodES256, claims)
		return token.SignedString(s.privateKey.(*ecdsa.PrivateKey))

	default:
		return "", ErrUnsupportedAlgorithm
	}
}

// ParseToken 解析并验证 JWT
func (s *JWTService) ParseToken(tokenStr string) (*jwt.Token, error) {
	var keyFunc jwt.Keyfunc

	switch s.algorithm {
	case "HS256":
		keyFunc = func(t *jwt.Token) (interface{}, error) {
			return s.secret, nil
		}
	case "RS256":
		keyFunc = func(t *jwt.Token) (interface{}, error) {
			return s.publicKey.(*rsa.PublicKey), nil
		}
	case "ES256":
		keyFunc = func(t *jwt.Token) (interface{}, error) {
			return s.publicKey.(*ecdsa.PublicKey), nil
		}
	default:
		return nil, ErrUnsupportedAlgorithm
	}

	return jwt.Parse(tokenStr, keyFunc)
}
