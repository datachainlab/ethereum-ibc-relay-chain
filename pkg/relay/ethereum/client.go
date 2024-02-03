package ethereum

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
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
	switch chain.config.TxType {
	case TxTypeLegacy:
		gasPrice, err := chain.Client().SuggestGasPrice(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to suggest gas price: %v", err)
		}
		txOpts.GasPrice = gasPrice
		return txOpts, nil
	case TxTypeDynamic:
		gasTipCap, gasFeeCap, err := chain.feeHistory(ctx)
		if err != nil {
			return nil, err
		}
		// GasTipCap = min(LimitPriorityFeePerGas, simulated_eth_maxPriorityFeePerGas * PriorityFeeRate)
		chain.config.DynamicTxGasConfig.PriorityFeeRate.Mul(gasTipCap)
		if l := chain.config.GetLimitPriorityFeePerGas(); l.Sign() > 0 && gasTipCap.Cmp(l) > 0 {
			gasTipCap = l
		}
		// GasFeeCap = min(LimitFeePerGas, GasTipCap + BaseFee * BaseFeeRate)
		chain.config.DynamicTxGasConfig.BaseFeeRate.Mul(gasFeeCap)
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
	default:
		return txOpts, nil
	}
}

func (chain *Chain) feeHistory(ctx context.Context) (*big.Int, *big.Int, error) {
	rewardPercentile := float64(chain.config.DynamicTxGasConfig.FeeHistoryRewardPercentile)
	maxRetry := chain.config.DynamicTxGasConfig.MaxRetryForFeeHistory

	latest, hErr := chain.Client().HeaderByNumber(ctx, nil)
	if hErr != nil {
		return nil, nil, fmt.Errorf("failed to get latest header: %v", hErr)
	}
	for i := uint32(0); i < maxRetry+1; i++ {
		block := big.NewInt(0).Sub(latest.Number, big.NewInt(int64(i)))
		history, err := chain.Client().FeeHistory(ctx, 1, block, []float64{rewardPercentile})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get feeHistory: %v", err)
		}
		if gasTipCap, baseFee, ok := getFeeInfo(history); ok {
			return gasTipCap, baseFee, nil
		}
	}
	return nil, nil, fmt.Errorf("no fee was found: latest=%v, maxRetry=%d", latest, maxRetry)
}

func getFeeInfo(v *ethereum.FeeHistory) (*big.Int, *big.Int, bool) {
	if len(v.Reward) == 0 || len(v.Reward[0]) == 0 || v.Reward[0][0].Cmp(big.NewInt(0)) == 0 {
		return nil, nil, false
	}
	gasTipCap := v.Reward[0][0]

	if len(v.BaseFee) < 1 {
		return nil, nil, false
	}
	// history.BaseFee[0] is baseFee (same as chain.Client().HeaderByNumber(ctx, nil).BaseFee)
	// history.BaseFee[1] is nextBaseFee
	baseFee := v.BaseFee[0]
	return gasTipCap, baseFee, true
}
