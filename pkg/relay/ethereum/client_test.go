package ethereum_test

import (
	"context"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/relay/ethereum"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/relay/ethereum/signers/hd"
	"math/big"
	"testing"
)

func Test_TxOpts_LegacyTx(t *testing.T) {
	config := createConfig()
	config.RpcAddr = "https://bsc-dataseed1.binance.org/"
	config.TxType = "legacy"
	chain, err := ethereum.NewChain(*config)
	if err != nil {
		t.Fatal(err)
	}
	txOpts, err := chain.TxOpts(context.Background())
	if err != nil {
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
	config := createConfig()
	config.RpcAddr = "https://ethereum.publicnode.com"
	chain, err := ethereum.NewChain(*config)
	if err != nil {
		t.Fatal(err)
	}
	txOpts, err := chain.TxOpts(context.Background())
	if err != nil {
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
	config := createConfig()
	config.RpcAddr = "https://ethereum.publicnode.com"
	config.TxType = "auto"
	chain, err := ethereum.NewChain(*config)
	if err != nil {
		t.Fatal(err)
	}
	txOpts, err := chain.TxOpts(context.Background())
	if err != nil {
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

func createConfig() *ethereum.ChainConfig {
	signer, err := types.NewAnyWithValue(&hd.SignerConfig{
		Mnemonic: "math razor capable expose worth grape metal sunset metal sudden usage scheme",
		Path:     "m/44'/60'/0'/0/0",
	})
	if err != nil {
		panic(err)
	}
	return &ethereum.ChainConfig{
		Signer: signer,
		TxType: "dynamic",
		DynamicTxGasConfig: &ethereum.DynamicTxGasConfig{
			LimitPriorityFeePerGas: "1ether",
			PriorityFeeRate: &ethereum.Fraction{
				Numerator:   1,
				Denominator: 1,
			},
			LimitFeePerGas: "1ether",
			// https://github.com/ethereum/go-ethereum/blob/0b471c312a82adf172bf6efdc7e3fdf285c62fba/accounts/abi/bind/base.go#L35
			BaseFeeRate: &ethereum.Fraction{
				Numerator:   2,
				Denominator: 1,
			},
			//https://github.com/NomicFoundation/hardhat/blob/197118fb9f92034d250e7e7d12f69e28f960d3b1/packages/hardhat-core/src/internal/core/providers/gas-providers.ts#L248
			FeeHistoryRewardPercentile: 50,
			MaxRetryForFeeHistory:      1,
		},
	}
}
