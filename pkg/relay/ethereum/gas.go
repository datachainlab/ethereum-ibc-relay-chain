package ethereum

import (
	"context"
	"fmt"
	"math/big"

	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/client/txpool"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type GasFeeCalculator struct {
	client IChainClient
	config *ChainConfig
}

func NewGasFeeCalculator(client IChainClient, config *ChainConfig) *GasFeeCalculator {
	return &GasFeeCalculator{
		client: client,
		config: config,
	}
}

func (m *GasFeeCalculator) Apply(ctx context.Context, txOpts *bind.TransactOpts) error {
	minFeeCap := common.Big0
	minTipCap := common.Big0
	var oldTx *txpool.RPCTransaction
	if m.config.PriceBump > 0 && txOpts.Nonce != nil {
		var err error
		if oldTx, minFeeCap, minTipCap, err = m.client.GetMinimumRequiredFee(ctx, txOpts.From, txOpts.Nonce.Uint64(), m.config.PriceBump); err != nil {
			return err
		}
	}
	switch m.config.TxType {
	case TxTypeLegacy:
		gasPrice, err := m.client.SuggestGasPrice(ctx)
		if err != nil {
			return fmt.Errorf("failed to suggest gas price: %v", err)
		}
		if oldTx != nil && oldTx.GasPrice != nil && oldTx.GasPrice.ToInt().Cmp(gasPrice) > 0 {
			return fmt.Errorf("old tx's gasPrice(%v) is higher than suggestion(%v)", oldTx.GasPrice.ToInt(), gasPrice)
		}
		if gasPrice.Cmp(minFeeCap) < 0 {
			gasPrice = minFeeCap
		}
		txOpts.GasPrice = gasPrice
		return nil
	case TxTypeDynamic:
		gasTipCap, gasFeeCap, err := m.feeHistory(ctx)
		if err != nil {
			return err
		}
		// GasTipCap = min(LimitPriorityFeePerGas, simulated_eth_maxPriorityFeePerGas * PriorityFeeRate)
		m.config.DynamicTxGasConfig.PriorityFeeRate.Mul(gasTipCap)

		// GasFeeCap = min(LimitFeePerGas, GasTipCap + BaseFee * BaseFeeRate)
		m.config.DynamicTxGasConfig.BaseFeeRate.Mul(gasFeeCap)
		gasFeeCap.Add(gasFeeCap, gasTipCap)

		if oldTx != nil && oldTx.GasFeeCap != nil && oldTx.GasTipCap != nil {
			if oldTx.GasFeeCap.ToInt().Cmp(gasFeeCap) > 0 && oldTx.GasTipCap.ToInt().Cmp(gasTipCap) > 0 {
				return fmt.Errorf("old tx's gasFeeCap(%v) and gasTipCap(%v) are higher than suggestion(%v, %v)", oldTx.GasFeeCap.ToInt(), oldTx.GasTipCap.ToInt(), gasFeeCap, gasTipCap)
			}
		}

		if gasTipCap.Cmp(minTipCap) < 0 {
			gasTipCap = minTipCap
		}
		if l := m.config.DynamicTxGasConfig.GetLimitPriorityFeePerGas(); l.Sign() > 0 && gasTipCap.Cmp(l) > 0 {
			gasTipCap = l
		}

		if gasFeeCap.Cmp(minFeeCap) < 0 {
			gasFeeCap = minFeeCap
		}
		if l := m.config.DynamicTxGasConfig.GetLimitFeePerGas(); l.Sign() > 0 && gasFeeCap.Cmp(l) > 0 {
			gasFeeCap = l
		}

		if gasFeeCap.Cmp(gasTipCap) < 0 {
			return fmt.Errorf("maxFeePerGas (%v) < maxPriorityFeePerGas (%v)", gasFeeCap, gasTipCap)
		}
		txOpts.GasFeeCap = gasFeeCap
		txOpts.GasTipCap = gasTipCap
		return nil
	default:
		return nil
	}
}

func (m *GasFeeCalculator) feeHistory(ctx context.Context) (*big.Int, *big.Int, error) {
	rewardPercentile := float64(m.config.DynamicTxGasConfig.FeeHistoryRewardPercentile)
	maxRetry := m.config.DynamicTxGasConfig.MaxRetryForFeeHistory

	latest, hErr := m.client.HeaderByNumber(ctx, nil)
	if hErr != nil {
		return nil, nil, fmt.Errorf("failed to get latest header: %v", hErr)
	}
	for i := uint32(0); i < maxRetry+1; i++ {
		block := big.NewInt(0).Sub(latest.Number, big.NewInt(int64(i)))
		history, err := m.client.FeeHistory(ctx, 1, block, []float64{rewardPercentile})
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
