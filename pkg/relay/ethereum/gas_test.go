package ethereum

import (
	"context"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/client"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"math/big"
	"testing"
)

func Test_TxOpts_LegacyTx(t *testing.T) {
	ethClient, err := client.NewETHClient("https://bsc-dataseed1.binance.org/")
	if err != nil {
		t.Fatal(err)
	}
	config := createConfig()
	config.TxType = "legacy"
	builder := NewGasOptionBuilder(ethClient, config)
	txOpts := &bind.TransactOpts{}
	if err = builder.Set(context.Background(), txOpts); err != nil {
		t.Fatal(err)
	}
	if txOpts.GasTipCap != nil {
		t.Error("gasTipCap must be nil")
	}
	if txOpts.GasFeeCap != nil {
		t.Error("gasFeeCap must be nil")
	}
	if txOpts.GasPrice == nil || txOpts.GasPrice.Cmp(big.NewInt(0)) == 0 {
		t.Error("gasPrice must be suggested")
	}
}

func Test_TxOpts_DynamicTx(t *testing.T) {
	ethClient, err := client.NewETHClient("https://ethereum.publicnode.com")
	if err != nil {
		t.Fatal(err)
	}
	config := createConfig()
	builder := NewGasOptionBuilder(ethClient, config)
	txOpts := &bind.TransactOpts{}
	if err = builder.Set(context.Background(), txOpts); err != nil {
		t.Fatal(err)
	}
	if txOpts.GasTipCap == nil || txOpts.GasTipCap.Cmp(big.NewInt(0)) == 0 {
		t.Error("gasTipCap must be suggested")
	}
	if txOpts.GasFeeCap == nil || txOpts.GasFeeCap.Cmp(big.NewInt(0)) == 0 {
		t.Error("gasFeeCap must be suggested")
	}
	if txOpts.GasPrice != nil {
		t.Error("gasPrice must be nil")
	}
}

func Test_TxOpts_AutoTx(t *testing.T) {
	ethClient, err := client.NewETHClient("https://ethereum.publicnode.com")
	if err != nil {
		t.Fatal(err)
	}
	config := createConfig()
	config.TxType = "auto"
	builder := NewGasOptionBuilder(ethClient, config)
	txOpts := &bind.TransactOpts{}
	if err = builder.Set(context.Background(), txOpts); err != nil {
		t.Fatal(err)
	}
	if txOpts.GasTipCap != nil {
		t.Error("gasTipCap must be nil")
	}
	if txOpts.GasFeeCap != nil {
		t.Error("gasFeeCap must be nil")
	}
	if txOpts.GasPrice != nil {
		t.Error("gasPrice must be nil")
	}
}

func createConfig() *ChainConfig {
	return &ChainConfig{
		TxType: "dynamic",
		DynamicTxGasConfig: &DynamicTxGasConfig{
			LimitPriorityFeePerGas: "1ether",
			PriorityFeeRate: &Fraction{
				Numerator:   1,
				Denominator: 1,
			},
			LimitFeePerGas: "1ether",
			// https://github.com/ethereum/go-ethereum/blob/0b471c312a82adf172bf6efdc7e3fdf285c62fba/accounts/abi/bind/base.go#L35
			BaseFeeRate: &Fraction{
				Numerator:   2,
				Denominator: 1,
			},
			//https://github.com/NomicFoundation/hardhat/blob/197118fb9f92034d250e7e7d12f69e28f960d3b1/packages/hardhat-core/src/internal/core/providers/gas-providers.ts#L248
			FeeHistoryRewardPercentile: 50,
			MaxRetryForFeeHistory:      1,
		},
	}
}

func Test_getFeeInfo(t *testing.T) {
	feeHistory := &ethereum.FeeHistory{}
	if _, _, ok := getFeeInfo(feeHistory); ok {
		t.Fatal("must be error")
	}
	feeHistory.Reward = append(feeHistory.Reward, []*big.Int{})
	if _, _, ok := getFeeInfo(feeHistory); ok {
		t.Fatal("must be error")
	}
	feeHistory.Reward[0] = append(feeHistory.Reward[0], big.NewInt(0))
	if _, _, ok := getFeeInfo(feeHistory); ok {
		t.Fatal("must be error")
	}
	feeHistory.Reward[0][0] = big.NewInt(1)
	if _, _, ok := getFeeInfo(feeHistory); ok {
		t.Fatal("must be error")
	}
	feeHistory.BaseFee = append(feeHistory.BaseFee, big.NewInt(2))
	if gasTip, baseFee, ok := getFeeInfo(feeHistory); ok {
		if gasTip.Int64() != int64(1) {
			t.Error("invalid gasTip", gasTip.Int64())
		}
		if baseFee.Int64() != int64(2) {
			t.Error("invalid baseFee", baseFee.Int64())
		}
	} else {
		t.Fatal("unexpected")
	}

}
