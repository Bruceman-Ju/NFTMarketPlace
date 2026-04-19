package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"

	"NFTMarketPlace-backend/cache"
	"NFTMarketPlace-backend/config"
	"NFTMarketPlace-backend/eth"
	"NFTMarketPlace-backend/listener"
	"NFTMarketPlace-backend/repository"
	"NFTMarketPlace-backend/routes"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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
	ethHttpClient, httpErr := eth.NewETHHttpClient(config.Cfg.Eth.RPCURL, config.Cfg.Eth.ContractAddress)
	if httpErr != nil {
		log.Fatal("Failed to connect Ethereum node")
	}
	ethWebsocketClient, websocketErr := eth.NewETHWebsocketClient(config.Cfg.Eth.WebSocketURL, config.Cfg.Eth.ContractAddress)
	if websocketErr != nil {
		log.Fatal("Failed to connect Ethereum node")
	}

	// Repo
	repo := repository.New(db)

	// Start listener in goroutine
	l := listener.NewListener(ethHttpClient, ethWebsocketClient, repo, redisCache)
	l.Start(context.Background())

	// HTTP server
	handler := gin.Default()
	routes.RegisterRoutes(handler, repo)

	service := &http.Server{
		Addr:    ":" + strconv.Itoa(config.Cfg.Server.Port),
		Handler: handler,
	}

	logrus.Infof("Server starting on port %d", config.Cfg.Server.Port)
	if err := service.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logrus.Fatalf("Server failed: %v", err)
	}
}
