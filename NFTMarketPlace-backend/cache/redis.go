package cache

import (
	"context"
	"encoding/json"
	"time"

	"NFTMarketPlace-backend/config"
	"NFTMarketPlace-backend/models"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedis() *RedisCache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Cfg.Redis.Addr,
		Password: config.Cfg.Redis.Password,
		DB:       config.Cfg.Redis.DB,
	})
	return &RedisCache{
		client: rdb,
		ctx:    context.Background(),
	}
}

func (r *RedisCache) AddListing(nft *models.ListedNFT) {
	// Cache individual
	key := "listing:" + nft.ListID
	data, _ := json.Marshal(nft)
	r.client.Set(r.ctx, key, data, 24*time.Hour)

	// Add to user set
	userKey := "user:listings:" + nft.Seller
	r.client.ZAdd(r.ctx, userKey, redis.Z{
		Score:  float64(nft.ListedTime),
		Member: nft.ListID,
	})

	// Add to global active set
	globalKey := "global:listings:active"
	r.client.ZAdd(r.ctx, globalKey, redis.Z{
		Score:  float64(nft.ListedTime),
		Member: nft.ListID,
	})
}

func (r *RedisCache) RemoveListing(listID string) {
	r.client.Del(r.ctx, "listing:"+listID)
	// Note: ZREM requires scanning all user keys – not efficient.
	// In practice, rely on DB as source of truth, use Redis as L1 cache with TTL.
}

func (r *RedisCache) GetListing(listID string) (*models.ListedNFT, error) {
	val, err := r.client.Get(r.ctx, "listing:"+listID).Result()
	if err != nil {
		return nil, err
	}
	var nft models.ListedNFT
	json.Unmarshal([]byte(val), &nft)
	return &nft, nil
}
