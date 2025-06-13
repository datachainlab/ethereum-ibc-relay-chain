//go:build ethereum_ibc_relay_chain_debug

package txpool

import (
	"os"
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ContentFrom calls `txpool_contentFrom` of the Ethereum RPC
func ContentFrom(ctx context.Context, client *ethclient.Client, address common.Address) (map[string]map[string]*RPCTransaction, error) {
	res := make(map[string]map[string]*RPCTransaction)
	val, ok := os.LookupEnv("RELAYER_SKIP_TXPOOL_CONTENT_FROM")
	if ok && val == "1" {
	} else {
		if err := client.Client().CallContext(ctx, &res, "txpool_contentFrom", address); err != nil {
			return nil, err
		}
	}
	return res, nil
}
