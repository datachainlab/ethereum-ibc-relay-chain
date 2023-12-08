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
	gasTipCap.Mul(gasTipCap, new(big.Int).SetUint64(chain.config.PriorityFeeRate.Numerator))
	gasTipCap.Div(gasTipCap, new(big.Int).SetUint64(chain.config.PriorityFeeRate.Denominator))
	if limit := new(big.Int).SetUint64(chain.config.LimitPriorityFeePerGas); gasTipCap.Cmp(limit) > 0 {
		gasTipCap.Set(limit)
	}

	// GasFeeCap = min(LimitFeePerGas, GasTipCap + BaseFee * BaseFeeRate)
	head, err := chain.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the latest header: %v", err)
	}
	gasFeeCap := new(big.Int).Mul(head.BaseFee, new(big.Int).SetUint64(chain.config.BaseFeeRate.Numerator))
	gasFeeCap.Div(gasFeeCap, new(big.Int).SetUint64(chain.config.BaseFeeRate.Denominator))
	gasFeeCap.Add(gasFeeCap, gasTipCap)
	if limit := new(big.Int).SetUint64(chain.config.LimitFeePerGas); gasFeeCap.Cmp(limit) > 0 {
		gasFeeCap.Set(limit)
	}

	return &bind.TransactOpts{
		From:      chain.signer.Address(),
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Signer:    chain.signer.Sign,
	}, nil
}
