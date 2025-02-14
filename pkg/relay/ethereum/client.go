package ethereum

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/client"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/client/txpool"
)

func (chain *Chain) CallOpts(ctx context.Context, height int64) *bind.CallOpts {
	opts := &bind.CallOpts{
		From:    chain.ethereumSigner.Address(),
		Context: ctx,
	}
	if height > 0 {
		opts.BlockNumber = big.NewInt(height)
	}
	return opts
}

func (chain *Chain) TxOpts(ctx context.Context, useLatestNonce bool) (*bind.TransactOpts, error) {
	addr := chain.ethereumSigner.Address()

	txOpts := &bind.TransactOpts{
		From:   addr,
		Signer: chain.ethereumSigner.Sign,
	}

	if useLatestNonce {
		if nonce, err := chain.client.NonceAt(ctx, addr, nil); err != nil {
			return nil, err
		} else {
			txOpts.Nonce = new(big.Int).SetUint64(nonce)
		}
	}

	if err := NewGasFeeCalculator(chain.client, &chain.config).Apply(ctx, txOpts); err != nil {
		return nil, err
	}

	return txOpts, nil
}

// wrapping interface of client.ETHClient struct
type IChainClient interface {
	ethereum.ChainReader
	ethereum.GasPricer
	ethereum.FeeHistoryReader

	GetMinimumRequiredFee(ctx context.Context, address common.Address, nonce uint64, priceBump uint64) (*txpool.RPCTransaction, *big.Int, *big.Int, error)
}

type ChainClient struct {
	*client.ETHClient
}

func (cl *ChainClient) GetMinimumRequiredFee(ctx context.Context, address common.Address, nonce uint64, priceBump uint64) (*txpool.RPCTransaction, *big.Int, *big.Int, error) {
	return txpool.GetMinimumRequiredFee(ctx, cl.ETHClient.Client, address, nonce, priceBump)
}
