package auth

import (
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// VerifyEIP191Signature 验证 EIP-191 签名（兼容 MetaMask）
func VerifyEIP191Signature(signatureHex, message, address string) (bool, error) {
	if !common.IsHexAddress(address) {
		return false, fmt.Errorf("invalid address")
	}
	expectedAddr := common.HexToAddress(address)

	sig, err := hex.DecodeString(signatureHex[2:]) // remove 0x
	if err != nil {
		return false, err
	}

	// EIP-191: \x19Ethereum Signed Message:\n + len(message) + message
	prefixedMsg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	hash := crypto.Keccak256Hash([]byte(prefixedMsg))

	pubKey, err := crypto.SigToPub(hash.Bytes(), sig)
	if err != nil {
		return false, err
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	return recoveredAddr == expectedAddr, nil
}
