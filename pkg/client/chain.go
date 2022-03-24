package client

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

func InitChain(endpoint string) (*ethclient.Client, error) {
	rpcClient, err := rpc.DialContext(context.Background(), endpoint)
	if err != nil {
		return nil, fmt.Errorf("dial eth client: %w", err)
	}
	return ethclient.NewClient(rpcClient), nil
}
