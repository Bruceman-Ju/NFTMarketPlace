package listener

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"NFTMarketPlace-backend/cache"
	"NFTMarketPlace-backend/config"
	"NFTMarketPlace-backend/eth"
	"NFTMarketPlace-backend/repository"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

const (
	maxBlocksPerRequest = 100 // 根据 RPC 服务商限制调整（Infura: 10k, Alchemy: 10k, 本地节点可更大）
	batchSize           = 10  // 每批处理的日志数量
)

var contractAbi = eth.ContractAbi // 假设你在 eth 包暴露了它

type Listener struct {
	ethHttpClient      *eth.Client
	ethWebsocketClient *eth.Client
	dbRepo             *repository.Repository
	cache              *cache.RedisCache
	contractAddr       common.Address
	startBlock         uint64
}

func NewListener(ethHttpClient *eth.Client, ethWebsocketClient *eth.Client, repo *repository.Repository, cache *cache.RedisCache) *Listener {
	return &Listener{
		ethHttpClient:      ethHttpClient,
		ethWebsocketClient: ethWebsocketClient,
		dbRepo:             repo,
		cache:              cache,
		contractAddr:       common.HexToAddress(config.Cfg.Eth.ContractAddress),
		startBlock:         config.Cfg.Eth.StartBlock,
	}
}

func (l *Listener) Start(ctx context.Context) {

	latestBlock := make(chan uint64, 1)
	// 启动监听协程
	go l.startListener(ctx, latestBlock)
	// 启动同步协程
	go l.syncHistory(ctx, latestBlock)

}

func (l *Listener) startListener(ctx context.Context, latestBlock chan uint64) {

	logrus.Info("Starting listener...")
	headers := make(chan *types.Header)
	var sub, subErr = l.ethWebsocketClient.SubscribeLatestBlock(ctx, headers)
	if subErr != nil {
		log.Fatal("Failed to subscribe to new heads:", subErr)
	}
	defer sub.Unsubscribe()

	for {
		select {
		case header := <-headers:
			logrus.Infof("New block: %d", header.Number.Uint64())
			latestBlock <- header.Number.Uint64()
		case err := <-sub.Err():
			log.Printf("Subscription error: %v", err)
			sub, subErr = l.ethWebsocketClient.SubscribeLatestBlock(ctx, headers)
			if subErr != nil {
				log.Fatal("Subscribe err, try to re-subscribe to latest block:", subErr)
			}
			return
		}
	}
}
func (l *Listener) syncHistory(ctx context.Context, latestBlockNum <-chan uint64) {

	// 1. check checkpoint in db
	lastSyncBlock := l.dbRepo.GetLastSyncBlock()
	current := lastSyncBlock.LastProcessedBlock

	step := uint64(maxBlocksPerRequest)

	syncBlockNum, recentBlockNumErr := l.ethHttpClient.GetBlockNumber()
	if recentBlockNumErr != nil {
		logrus.Fatalf("Failed to get recent block number: %v", recentBlockNumErr)
		return
	}

	latestFinalHeader, finalBLockErr := l.ethHttpClient.GetFinalizedBlockNumber()

	if finalBLockErr != nil {
		logrus.Fatalf("Failed to get finalized block number: %v", finalBLockErr)
		return
	}

	for {
		select {
		case <-ctx.Done():
			logrus.Info("History sync stopped by system")
			return
		case latest := <-latestBlockNum:
			logrus.Infof("Received latest block num %d", latest)
			syncBlockNum = latest
		default:
			// 没有新区块就继续执行。
		}

		if current == syncBlockNum {
			timer := time.NewTimer(12 * time.Second)
			logrus.Info("All historical events sync complete, will sleep 12 seconds")
			select {
			case <-ctx.Done():
				logrus.Info("History sync stopped by system")
				return
			case latest := <-latestBlockNum:
				syncBlockNum = latest
				if !timer.Stop() {
					<-timer.C
				}
			case <-timer.C:
				logrus.Info("12 seconds gone without new block subscribed")
				continue
			}
		}

		logrus.Infof("Syncing historical events from block %d to block %d", current, syncBlockNum)

		end := current + step - 1
		if end > syncBlockNum {
			end = syncBlockNum
		}

		logrus.Infof("Fetching logs from %d to %d...", current, end)

		query := l.buildQuery(current, end)
		var logs []types.Log
		var err error
		for attempt := 0; attempt < 3; attempt++ {
			logs, err = l.ethHttpClient.FilterLogs(ctx, query)
			if err == nil {
				break
			}
			logrus.Warnf("Failed to fetch logs [%d-%d], attempt %d: %v", current, end, attempt+1, err)
			time.Sleep(time.Second * time.Duration(2<<attempt)) // 指数退避
		}

		if err != nil {
			logrus.Fatalf("Failed to fetch logs after retries [%d-%d]: %v", current, end, err)
			break
		}

		if current < latestFinalHeader.Number.Uint64() {
			logrus.Infof("Sync with final block %d", current)
			// db
			// 分批处理日志（避免内存堆积）
			l.processLogsInBatches(logs)
		} else if current == latestFinalHeader.Number.Uint64() {
			logrus.Info("Update final block number")
			// 更新最新 final 区块
			latestFinalHeader, finalBLockErr = l.ethHttpClient.GetFinalizedBlockNumber()
		} else {
			logrus.Info("Start sync pending block %d", current)
			// redis
			fmt.Println("Redis done")
		}

		current = end + 1

		// 防止 RPC 限流（可选）
		time.Sleep(100 * time.Millisecond)
	}
}

// 分批处理日志
func (l *Listener) processLogsInBatches(logs []types.Log) {
	logsCapacity := len(logs)
	for i := 0; i < logsCapacity; i += batchSize {
		end := i + batchSize
		if end > logsCapacity {
			end = logsCapacity
		}
		batch := logs[i:end]
		fmt.Println("Start is ", i, "end is ", end)

		l.processLogs(batch)

		if i%1000 == 0 {
			logrus.Debugf("Processed %d/%d logs in current segment", i, logsCapacity)
		}
	}
}

func (l *Listener) buildQuery(from, to uint64) ethereum.FilterQuery {
	var toBlock *big.Int
	if to > 0 {
		toBlock = big.NewInt(int64(to))
	}
	return ethereum.FilterQuery{
		Addresses: []common.Address{l.contractAddr},
		FromBlock: big.NewInt(int64(from)),
		ToBlock:   toBlock,
	}
}

func (l *Listener) processLogs(logs []types.Log) {
	for _, logItem := range logs {
		l.processLog(logItem)
	}
}

func (l *Listener) processLog(log types.Log) {
	l.dbRepo.UpdateSyncBlock(log.BlockNumber)
	switch log.Topics[0].Hex() {
	case contractAbi.Events["NFTListed"].ID.Hex():

		nft, err := eth.ParseNFTListed(log)
		if err != nil {
			logrus.Errorf("Parse listed error: %v", err)
			return
		}

		errListNFT := l.dbRepo.CreateListing(nft)
		if errListNFT != nil {
			logrus.Errorf("Create listing error: %v", errListNFT)
			return
		}

		l.cache.AddListing(nft, 30*time.Minute)

	case contractAbi.Events["NFTSold"].ID.Hex():
		sale, listID, err := eth.ParseNFTSold(log)
		if err != nil {
			logrus.Errorf("Parse sold error: %v", err)
			return
		}

		errSale := l.dbRepo.CreateSale(sale)
		if errSale != nil {
			logrus.Errorf("Create sale error: %v", errSale)
			return
		}
		errListSold := l.dbRepo.MarkListingAsSold(listID)

		if errListSold != nil {
			logrus.Errorf("Mark listing as sold error: %v", errListSold)
			return
		}

		l.cache.RemoveListing(listID)

	case contractAbi.Events["NFTCanceled"].ID.Hex(), contractAbi.Events["NFTExpired"].ID.Hex():
		listCanceled, err := eth.ParseNFTCanceled(log)
		if err != nil {
			logrus.Errorf("Parse canceled error: %v", err)
			return
		}

		errNFTCancel := l.dbRepo.MarkListingAsInactive(listCanceled.ListID, "canceled") // or "expired"
		if errNFTCancel != nil {
			logrus.Errorf("Mark listing as inactive error: %v", errNFTCancel)
			return
		}

		l.cache.RemoveListing(listCanceled.ListID)
	}
}
