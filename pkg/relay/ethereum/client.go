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
	txOpts := &bind.TransactOpts{
		From:   chain.signer.Address(),
		Signer: chain.signer.Sign,
	}
	if chain.config.EnableLegacyTx {
		gasPrice, err := chain.Client().SuggestGasPrice(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to suggest gas price: %v", err)
		}
		txOpts.GasPrice = gasPrice
		return txOpts, nil
	}

	gasTipCap, gasFeeCap, err := chain.feeHistory(ctx, float64(chain.Config().RewardPercentile))
	if err != nil {
		return nil, err
	}
	// GasTipCap = min(LimitPriorityFeePerGas, Suggested * PriorityFeeRate)
	chain.config.PriorityFeeRate.Mul(gasTipCap)
	if l := chain.config.GetLimitPriorityFeePerGas(); l.Sign() > 0 && gasTipCap.Cmp(l) > 0 {
		gasTipCap = l
	}
	// GasFeeCap = min(LimitFeePerGas, GasTipCap + BaseFee * BaseFeeRate)
	chain.config.BaseFeeRate.Mul(gasFeeCap)
	gasFeeCap.Add(gasFeeCap, gasTipCap)
	if l := chain.config.GetLimitFeePerGas(); l.Sign() > 0 && gasFeeCap.Cmp(l) > 0 {
		gasFeeCap = l
	}

	if gasFeeCap.Cmp(gasTipCap) < 0 {
		return nil, fmt.Errorf("maxFeePerGas (%v) < maxPriorityFeePerGas (%v)", gasFeeCap, gasTipCap)
	}
	txOpts.GasFeeCap = gasFeeCap
	txOpts.GasTipCap = gasTipCap
	return txOpts, nil
}

func (chain *Chain) feeHistory(ctx context.Context, rewardPercentile float64) (*big.Int, *big.Int, error) {
	history, err := chain.Client().FeeHistory(ctx, 1, nil, []float64{rewardPercentile})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get feeHistory: %v", err)
	}
	if len(history.Reward) == 0 {
		return nil, nil, fmt.Errorf("no reward found")
	}
	if len(history.Reward[0]) == 0 {
		return nil, nil, fmt.Errorf("no reward found")
	}
	if len(history.BaseFee) < 2 {
		return nil, nil, fmt.Errorf("insufficient base fee")
	}
	gasTipCap := history.Reward[0][0]
	gasFeeCap := history.BaseFee[1]
	return gasTipCap, gasFeeCap, nil
}
