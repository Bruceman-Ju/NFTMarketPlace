package utils

import (
	"net"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

func GetPagination(c *gin.Context, defaultPage, defaultSize int) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", strconv.Itoa(defaultPage)))
	size, _ := strconv.Atoi(c.DefaultQuery("size", strconv.Itoa(defaultSize)))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return page, size
}

func IsValidEthAddress(addr string) bool {
	if !common.IsHexAddress(addr) {
		return false
	}
	_, err := net.ResolveIPAddr("ip4", common.HexToAddress(addr).Hex())
	return err == nil // just a simple check; better to use regex or common.IsHexAddress only
}
