//go:build !ethereum_ibc_relay_chain_debug

package txpool

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ContentFrom calls `txpool_contentFrom` of the Ethereum RPC
func ContentFrom(ctx context.Context, client *ethclient.Client, address common.Address) (map[string]map[string]*RPCTransaction, error) {
	var res map[string]map[string]*RPCTransaction
	if err := client.Client().CallContext(ctx, &res, "txpool_contentFrom", address); err != nil {
		return nil, err
	}
	return res, nil
}
