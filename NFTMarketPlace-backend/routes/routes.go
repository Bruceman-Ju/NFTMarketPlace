package routes

import (
	"NFTMarketPlace-backend/auth"
	"NFTMarketPlace-backend/config"
	"NFTMarketPlace-backend/docs"
	"NFTMarketPlace-backend/handler"
	"NFTMarketPlace-backend/middleware"
	"NFTMarketPlace-backend/repository"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterRoutes(r *gin.Engine, repo *repository.Repository) {

	allowOrigins := cors.Config{
		AllowOrigins:     config.Cfg.Server.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	r.Use(cors.New(allowOrigins))

	jwtService, err := auth.NewJWTService(config.Cfg)
	if err != nil {
		log.Fatalf("Failed to init JWT service: %v", err)
	}

	authHandler := handler.NewAuthHandler(jwtService)

	r.GET("/api/v1/auth/nonce", authHandler.GetNonce)
	r.POST("/api/v1/auth/login", authHandler.Login)

	listingHandler := handler.NewListingHandler(repo)
	saleHandler := handler.NewSaleHandler(repo)

	docs.SwaggerInfo.BasePath = "/api/v1"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	protected := r.Group("/api/v1")
	protected.Use(middleware.JWTAuth(jwtService))
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
