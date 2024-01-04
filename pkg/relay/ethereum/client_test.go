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
	signer, err := types.NewAnyWithValue(&hd.SignerConfig{
		Mnemonic: "math razor capable expose worth grape metal sunset metal sudden usage scheme",
		Path:     "m/44'/60'/0'/0/0",
	})
	if err != nil {
		t.Fatal(err)
	}
	config := ethereum.ChainConfig{
		RpcAddr:        "https://bsc-dataseed1.binance.org/",
		Signer:         signer,
		EnableLegacyTx: true,
	}
	ctx := context.Background()
	chain, err := ethereum.NewChain(config)
	if err != nil {
		t.Fatal(err)
	}
	txOpts, err := chain.TxOpts(ctx)
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

func Test_TxOpts_DynamicTx_BSC_Error(t *testing.T) {
	signer, err := types.NewAnyWithValue(&hd.SignerConfig{
		Mnemonic: "math razor capable expose worth grape metal sunset metal sudden usage scheme",
		Path:     "m/44'/60'/0'/0/0",
	})
	if err != nil {
		t.Fatal(err)
	}
	config := ethereum.ChainConfig{
		RpcAddr:            "https://bsc-dataseed1.binance.org/",
		Signer:             signer,
		EnableLegacyTx:     false,
		DynamicTxGasConfig: &ethereum.DynamicTxGasConfig{},
	}
	ctx := context.Background()
	chain, err := ethereum.NewChain(config)
	if err != nil {
		t.Fatal(err)
	}
	_, err = chain.TxOpts(ctx)
	if err == nil || err.Error() != "suggested baseFeePerGas is zero" {
		t.Fatal("unexpected result")
	}
}

func Test_TxOpts_DynamicTx(t *testing.T) {
	signer, err := types.NewAnyWithValue(&hd.SignerConfig{
		Mnemonic: "math razor capable expose worth grape metal sunset metal sudden usage scheme",
		Path:     "m/44'/60'/0'/0/0",
	})
	if err != nil {
		t.Fatal(err)
	}
	config := ethereum.ChainConfig{
		RpcAddr:        "https://ethereum.publicnode.com",
		Signer:         signer,
		EnableLegacyTx: false,
		DynamicTxGasConfig: &ethereum.DynamicTxGasConfig{
			LimitPriorityFeePerGas: "1ether",
			PriorityFeeRate: &ethereum.Fraction{
				Numerator:   1,
				Denominator: 1,
			},
			LimitFeePerGas: "1ether",
			BaseFeeRate: &ethereum.Fraction{
				Numerator:   1,
				Denominator: 1,
			},
		},
	}
	ctx := context.Background()
	chain, err := ethereum.NewChain(config)
	if err != nil {
		t.Fatal(err)
	}
	txOpts, err := chain.TxOpts(ctx)
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
