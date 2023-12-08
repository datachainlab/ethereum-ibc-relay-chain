package ethereum

import (
	"context"
	"fmt"
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
	// GasTipCap = min(LimitPriorityFeePerGas, Suggested * PriorityFeeRate)
	gasTipCap, err := chain.client.SuggestGasTipCap(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to call eth_maxPriorityFeePerGas: %v", err)
	}
	chain.config.PriorityFeeRate.Mul(gasTipCap)
	if l := chain.config.GetLimitPriorityFeePerGas(); l.Sign() > 0 && gasTipCap.Cmp(l) > 0 {
		gasTipCap = l
	}

	// GasFeeCap = min(LimitFeePerGas, GasTipCap + BaseFee * BaseFeeRate)
	head, err := chain.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the latest header: %v", err)
	}
	gasFeeCap := head.BaseFee
	chain.config.BaseFeeRate.Mul(gasFeeCap)
	gasFeeCap.Add(gasFeeCap, gasTipCap)
	if l := chain.config.GetLimitFeePerGas(); l.Sign() > 0 && gasFeeCap.Cmp(l) > 0 {
		gasFeeCap = l
	}

	return &bind.TransactOpts{
		From:      chain.signer.Address(),
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Signer:    chain.signer.Sign,
	}, nil
}
