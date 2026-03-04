package main

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"NFTMarketPlace-backend/cache"
	"NFTMarketPlace-backend/config"
	"NFTMarketPlace-backend/eth"
	"NFTMarketPlace-backend/handler"
	"NFTMarketPlace-backend/listener"
	"NFTMarketPlace-backend/repository"
	"NFTMarketPlace-backend/routes"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title NFT Marketplace API
// @version 1.0
// @description Backend API for NFT marketplace
// @host localhost:8080
// @BasePath /api/v1
// @schemes http https
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	config.Init()

	// DB
	db, err := gorm.Open(mysql.Open(config.Cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect DB")
	}
	//db.AutoMigrate(&models.ListedNFT{}, &models.Sale{}, &models.SyncState{})

	// Redis
	redisCache := cache.NewRedis()

	// Eth client
	ethHttpClient, err := eth.NewETHHttpClient(config.Cfg.Eth.RPCURL, config.Cfg.Eth.ContractAddress)
	if err != nil {
		log.Fatal("Failed to connect Ethereum node")
	}
	ethWebsocketClient, err := eth.NewETHWebsocketClient(config.Cfg.Eth.WebSocketURL, config.Cfg.Eth.ContractAddress)
	if err != nil {
		log.Fatal("Failed to connect Ethereum node")
	}

	// Repo
	repo := repository.New(db)

	// Start listener in goroutine
	go func() {
		l := listener.NewListener(ethHttpClient, ethWebsocketClient, repo, redisCache)
		l.Start(context.Background())
	}()

	// HTTP server
	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	listingHandler := handler.NewListingHandler(repo)
	saleHandler := handler.NewSaleHandler(repo)
	routes.RegisterRoutes(r, listingHandler, saleHandler)

	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(config.Cfg.Server.Port),
		Handler: r,
	}

	logrus.Infof("Server starting on port %d", config.Cfg.Server.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logrus.Fatalf("Server failed: %v", err)
	}
}
