package txpool

import (
	"context"
	"math/big"
	"slices"

	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/client"
	"github.com/ethereum/go-ethereum/common"
)

// PendingTransactions returns pending txs sent from `address` sorted by nonce.
func PendingTransactions(ctx context.Context, cl client.IETHClient, address common.Address) ([]*client.RPCTransaction, error) {
	txs, err := cl.ContentFrom(ctx, address)
	if err != nil {
		return nil, err
	}

	pendingTxMap, found := txs["pending"]
	if !found {
		return nil, nil
	}

	var pendingTxs []*client.RPCTransaction
	for _, pendingTx := range pendingTxMap {
		pendingTxs = append(pendingTxs, pendingTx)
	}

	slices.SortFunc(pendingTxs, func(a, b *client.RPCTransaction) int {
		if a.Nonce < b.Nonce {
			return -1
		} else if a.Nonce > b.Nonce {
			return 1
		} else {
			return 0
		}
	})

	return pendingTxs, nil
}

func inclByPercent(n *big.Int, percent uint64) {
	n.Mul(n, big.NewInt(int64(100+percent)))
	n.Div(n, big.NewInt(100))
}

// GetMinimumRequiredFee returns the minimum fee required to successfully send a transaction
func GetMinimumRequiredFee(ctx context.Context, cl client.IETHClient, address common.Address, nonce uint64, priceBump uint64) (*client.RPCTransaction, *big.Int, *big.Int, error) {
	pendingTxs, err := PendingTransactions(ctx, cl, address)
	if err != nil {
		return nil, nil, nil, err
	} else if len(pendingTxs) == 0 {
		return nil, common.Big0, common.Big0, nil
	}

	var targetTx *client.RPCTransaction
	for _, pendingTx := range pendingTxs {
		if uint64(pendingTx.Nonce) == nonce {
			targetTx = pendingTx
			break
		}
	}
	if targetTx == nil {
		return nil, common.Big0, common.Big0, nil
	}
	gasFeeCap := new(big.Int).Set(targetTx.GasFeeCap.ToInt())
	gasTipCap := new(big.Int).Set(targetTx.GasTipCap.ToInt())

	inclByPercent(gasFeeCap, priceBump)
	inclByPercent(gasTipCap, priceBump)

	return targetTx, gasFeeCap, gasTipCap, nil
}
