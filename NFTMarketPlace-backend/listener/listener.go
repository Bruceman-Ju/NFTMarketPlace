package listener

import (
	"context"
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
	concurrentWorkers   = 3   // 并行 worker 数（谨慎使用，并行可能打满 RPC）
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
	lastBlock := l.dbRepo.GetLastSyncBlock()
	logrus.Infof("Last synced block: %d", lastBlock)

	current, _ := l.ethHttpClient.GetBlockNumber()
	logrus.Infof("Current block: %d", current)

	// Sync historical
	from := l.startBlock
	if lastBlock > from {
		from = lastBlock + 1
	}
	if from <= current {
		logrus.Infof("Syncing historical events from %d to %d", from, current)
		l.syncRange(ctx, from, current)
	}

	// Watch new
	logrus.Info("Starting real-time event watcher...")
	l.watchNew(ctx, current+1)
}

func (l *Listener) syncRange(ctx context.Context, from, to uint64) {
	logrus.Infof("Syncing historical events from block %d to %d", from, to)

	var current uint64 = from
	step := uint64(maxBlocksPerRequest)

	for current <= to {
		end := current + step - 1
		if end > to {
			end = to
		}

		logrus.Infof("Fetching logs from %d to %d...", current, end)
		query := l.buildQuery(current, end)

		// 带重试的 Fetch
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
		}

		// 分批处理日志（避免内存堆积）
		l.processLogsInBatches(logs)

		// 更新最后同步块（每段成功后更新，实现断点续传）
		l.dbRepo.UpdateSyncBlock(end)

		current = end + 1

		// 防止 RPC 限流（可选）
		time.Sleep(100 * time.Millisecond)
	}

	logrus.Info("Historical sync completed.")
}

// 分批处理日志（控制内存和 DB 压力）
func (l *Listener) processLogsInBatches(logs []types.Log) {
	for i := 0; i < len(logs); i += batchSize {
		end := i + batchSize
		if end > len(logs) {
			end = len(logs)
		}
		batch := logs[i:end]

		// 可选：并行处理 batch（注意 DB 并发安全）
		l.processLogs(batch)

		// 可选：打印进度
		if i%1000 == 0 {
			logrus.Debugf("Processed %d/%d logs in current segment", i, len(logs))
		}
	}
}

func (l *Listener) watchNew(ctx context.Context, from uint64) {
	query := l.buildQuery(from, 0) // 0 = latest
	ch := make(chan types.Log)
	sub, err := l.ethWebsocketClient.SubscribeFilterLogs(ctx, query, ch)
	if err != nil {
		logrus.Fatalf("Subscribe failed: %v", err)
	}
	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return
		case err := <-sub.Err():
			logrus.Errorf("Subscription error: %v", err)
			time.Sleep(5 * time.Second)
			// reconnect logic omitted for brevity
		case log := <-ch:
			l.processLog(log)
			l.dbRepo.UpdateSyncBlock(log.BlockNumber)
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
	for _, log := range logs {
		l.processLog(log)
	}
}

func (l *Listener) processLog(log types.Log) {
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
		l.cache.AddListing(nft)

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
