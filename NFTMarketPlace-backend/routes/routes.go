package routes

import (
	"NFTMarketPlace-backend/handler"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, listingHandler *handler.ListingHandler, saleHandler *handler.SaleHandler) {
	authHandler := handler.NewAuthHandler()

	r.POST("/api/v1/auth/login", authHandler.Login)
	r.GET("/api/v1/auth/nonce", authHandler.GetNonce)

	protected := r.Group("/api/v1")
	//protected.Use(middleware.JWTAuth())
	{
		protected.GET("/user/me", func(c *gin.Context) {
			addr := c.MustGet("address").(string)
			c.JSON(http.StatusOK, gin.H{"address": addr})
		})
		protected.GET("/listings", listingHandler.GetAllListings)
		protected.GET("/listings/user/:address", listingHandler.GetUserListings)
		protected.GET("/sales/seller/:address", saleHandler.GetUserSales)
		protected.GET("/sales/buyer/:address", saleHandler.GetBuyerPurchases)
	}
}
