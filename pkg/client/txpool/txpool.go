package txpool

import (
	"context"
	"math/big"
	"slices"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// RPCTransaction represents a transaction that will serialize to the RPC representation of a transaction
type RPCTransaction struct {
	BlockHash           *common.Hash      `json:"blockHash"`
	BlockNumber         *hexutil.Big      `json:"blockNumber"`
	From                common.Address    `json:"from"`
	Gas                 hexutil.Uint64    `json:"gas"`
	GasPrice            *hexutil.Big      `json:"gasPrice"`
	GasFeeCap           *hexutil.Big      `json:"maxFeePerGas,omitempty"`
	GasTipCap           *hexutil.Big      `json:"maxPriorityFeePerGas,omitempty"`
	MaxFeePerBlobGas    *hexutil.Big      `json:"maxFeePerBlobGas,omitempty"`
	Hash                common.Hash       `json:"hash"`
	Input               hexutil.Bytes     `json:"input"`
	Nonce               hexutil.Uint64    `json:"nonce"`
	To                  *common.Address   `json:"to"`
	TransactionIndex    *hexutil.Uint64   `json:"transactionIndex"`
	Value               *hexutil.Big      `json:"value"`
	Type                hexutil.Uint64    `json:"type"`
	Accesses            *types.AccessList `json:"accessList,omitempty"`
	ChainID             *hexutil.Big      `json:"chainId,omitempty"`
	BlobVersionedHashes []common.Hash     `json:"blobVersionedHashes,omitempty"`
	V                   *hexutil.Big      `json:"v"`
	R                   *hexutil.Big      `json:"r"`
	S                   *hexutil.Big      `json:"s"`
	YParity             *hexutil.Uint64   `json:"yParity,omitempty"`
}

// PendingTransactions returns pending txs sent from `address` sorted by nonce.
func PendingTransactions(ctx context.Context, client *ethclient.Client, address common.Address) ([]*RPCTransaction, error) {
	txs, err := ContentFrom(ctx, client, address)
	if err != nil {
		return nil, err
	}

	pendingTxMap, found := txs["pending"]
	if !found {
		return nil, nil
	}

	var pendingTxs []*RPCTransaction
	for _, pendingTx := range pendingTxMap {
		pendingTxs = append(pendingTxs, pendingTx)
	}

	slices.SortFunc(pendingTxs, func(a, b *RPCTransaction) int {
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
func GetMinimumRequiredFee(ctx context.Context, client *ethclient.Client, address common.Address, nonce uint64, priceBump uint64) (*RPCTransaction, *big.Int, *big.Int, error) {
	pendingTxs, err := PendingTransactions(ctx, client, address)
	if err != nil {
		return nil, nil, nil, err
	} else if len(pendingTxs) == 0 {
		return nil, common.Big0, common.Big0, nil
	}

	var targetTx *RPCTransaction
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
