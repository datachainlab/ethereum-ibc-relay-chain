package client

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/client/txpool"
)

func (cl *ETHClient) GetMinimumRequiredFee(ctx context.Context, address common.Address, nonce uint64, priceBump uint64) (*txpool.RPCTransaction, *big.Int, *big.Int, error) {
	return txpool.GetMinimumRequiredFee(ctx, cl.Client, address, nonce, priceBump);
}
