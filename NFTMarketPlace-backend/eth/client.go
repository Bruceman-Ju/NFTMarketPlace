package eth

import (
	"context"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
)

type Client struct {
	client *ethclient.Client
	addr   common.Address
}

func NewETHHttpClient(rpcURL string, contractAddr string) (*Client, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}
	return &Client{
		client: client,
		addr:   common.HexToAddress(contractAddr),
	}, nil
}

func NewETHWebsocketClient(rpcURL string, contractAddr string) (*Client, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}
	return &Client{
		client: client,
		addr:   common.HexToAddress(contractAddr),
	}, nil
}

func (c *Client) GetBlockNumber() (uint64, error) {
	return c.client.BlockNumber(context.Background())
}

func (c *Client) FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error) {
	return c.client.FilterLogs(ctx, query)
}

func (c *Client) SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (event.Subscription, error) {
	return c.client.SubscribeFilterLogs(ctx, query, ch)
}
