package ethereum

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
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

	if err := NewGasFeeCalculator(chain.client, &chain.config).Apply(ctx, txOpts); err != nil {
		return nil, err
	}

	if useLatestNonce {
		if nonce, err := chain.client.NonceAt(ctx, addr, nil); err != nil {
			return nil, err
		} else {
			txOpts.Nonce = new(big.Int).SetUint64(nonce)
		}
	}

	return txOpts, nil
}
