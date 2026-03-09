package auth

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

func loadRSAKeys(privatePath, publicPath string) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privBytes, err := os.ReadFile(privatePath)
	if err != nil {
		return nil, nil, err
	}
	priv, err := jwt.ParseRSAPrivateKeyFromPEM(privBytes)
	if err != nil {
		return nil, nil, err
	}

	pubBytes, err := os.ReadFile(publicPath)
	if err != nil {
		return nil, nil, err
	}
	pub, err := jwt.ParseRSAPublicKeyFromPEM(pubBytes)
	if err != nil {
		return nil, nil, err
	}

	return priv, pub, nil
}

func loadECDSAKeys(privatePath, publicPath string) (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	privBytes, err := os.ReadFile(privatePath)
	if err != nil {
		return nil, nil, err
	}
	block, _ := pem.Decode(privBytes)
	if block == nil {
		return nil, nil, errors.New("failed to decode PEM block containing private key")
	}
	priv, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, err
	}

	pubBytes, err := os.ReadFile(publicPath)
	if err != nil {
		return nil, nil, err
	}
	block, _ = pem.Decode(pubBytes)
	if block == nil {
		return nil, nil, errors.New("failed to decode PEM block containing public key")
	}
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, nil, err
	}
	pub, ok := pubKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, nil, errors.New("not an ECDSA public key")
	}

	return priv, pub, nil
}
