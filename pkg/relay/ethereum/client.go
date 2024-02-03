package ethereum

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

func (chain *Chain) CallOpts(ctx context.Context, height int64) *bind.CallOpts {
	opts := &bind.CallOpts{
		From:    chain.signer.Address(),
		Context: ctx,
	}
	if height > 0 {
		opts.BlockNumber = big.NewInt(height)
	}
	return opts
}

func (chain *Chain) TxOpts(ctx context.Context) (*bind.TransactOpts, error) {
	txOpts := &bind.TransactOpts{
		From:   chain.signer.Address(),
		Signer: chain.signer.Sign,
	}
	if err := NewGasOptionBuilder(chain.client, &chain.config).Set(ctx, txOpts); err != nil {
		return nil, err
	}
	return txOpts, nil
}
