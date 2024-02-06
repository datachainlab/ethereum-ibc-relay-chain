package pending

import (
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/relay/ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
	"math/big"
	"testing"
)

func TestSuccessCopyTxData(t *testing.T) {
	logic := Logic{}
	opts := &gasFees{
		GasTipCapInc: big.NewInt(10),
		MaxGasTipCap: big.NewInt(1000),
		GasFeeCapInc: big.NewInt(100),
		MaxGasFeeCap: big.NewInt(10000),
		GasPriceInc:  big.NewInt(200),
		MaxGasPrice:  big.NewInt(20000),
	}

	// Legacy Tx
	src := types.NewTx(&types.LegacyTx{
		GasPrice: big.NewInt(19800),
		Gas:      0,
	})
	txData, err := logic.copyTxData(src, opts)
	if err != nil {
		t.Error(err)
	}
	dst := types.NewTx(txData)
	if dst.GasPrice().Uint64() != uint64(20000) {
		t.Error("invalid LegacyTx gasPrice")
	}

	// AccessList Tx
	src = types.NewTx(&types.AccessListTx{
		GasPrice: big.NewInt(19800),
		Gas:      0,
	})
	txData, err = logic.copyTxData(src, opts)
	if err != nil {
		t.Error(err)
	}
	dst = types.NewTx(txData)
	if dst.GasPrice().Uint64() != uint64(20000) {
		t.Error("invalid AccessListTx gasPrice")
	}

	// Dynamic Tx
	src = types.NewTx(&types.DynamicFeeTx{
		GasTipCap: big.NewInt(990),
		GasFeeCap: big.NewInt(9900),
		Gas:       0,
	})
	txData, err = logic.copyTxData(src, opts)
	if err != nil {
		t.Error(err)
	}
	dst = types.NewTx(txData)
	if dst.GasTipCap().Uint64() != uint64(1000) {
		t.Error("invalid DynamicFeeTx gasPrice")
	}
	if dst.GasFeeCap().Uint64() != uint64(10000) {
		t.Error("invalid DynamicFeeTx GasFeeCap")
	}

	// Blob Tx
	src = types.NewTx(&types.BlobTx{
		GasTipCap: uint256.NewInt(990),
		GasFeeCap: uint256.NewInt(9900),
		Gas:       0,
	})
	txData, err = logic.copyTxData(src, opts)
	if err != nil {
		t.Error(err)
	}
	dst = types.NewTx(txData)
	if dst.GasTipCap().Uint64() != uint64(1000) {
		t.Error("invalid BlobTx gasPrice")
	}
	if dst.GasFeeCap().Uint64() != uint64(10000) {
		t.Error("invalid BlobTx GasFeeCap")

	}
}

func TestErrorCopyTxData(t *testing.T) {
	logic := Logic{}
	opts := &gasFees{
		GasTipCapInc: big.NewInt(10),
		MaxGasTipCap: big.NewInt(1000),
		GasFeeCapInc: big.NewInt(100),
		MaxGasFeeCap: big.NewInt(10000),
		GasPriceInc:  big.NewInt(200),
		MaxGasPrice:  big.NewInt(20000),
	}

	// Legacy Tx
	src := types.NewTx(&types.LegacyTx{
		GasPrice: big.NewInt(19801),
		Gas:      0,
	})
	_, err := logic.copyTxData(src, opts)
	if err == nil {
		t.Error("unexpected success for LegacyTx")
	}

	// AccessList Tx
	src = types.NewTx(&types.AccessListTx{
		GasPrice: big.NewInt(19801),
		Gas:      0,
	})
	_, err = logic.copyTxData(src, opts)
	if err == nil {
		t.Error("unexpected success for AccessListTx")
	}

	// Dynamic Tx
	src = types.NewTx(&types.DynamicFeeTx{
		GasTipCap: big.NewInt(991),
		GasFeeCap: big.NewInt(1000),
		Gas:       0,
	})
	_, err = logic.copyTxData(src, opts)
	if err == nil {
		t.Error("unexpected success for DynamicTx GasTipCap")
	}
	src = types.NewTx(&types.DynamicFeeTx{
		GasTipCap: big.NewInt(1),
		GasFeeCap: big.NewInt(9901),
		Gas:       0,
	})
	_, err = logic.copyTxData(src, opts)
	if err == nil {
		t.Error("unexpected success for DynamicTx GasFeeCap")
	}

	// Blob Tx
	src = types.NewTx(&types.BlobTx{
		GasTipCap: uint256.NewInt(991),
		GasFeeCap: uint256.NewInt(9900),
		Gas:       0,
	})
	_, err = logic.copyTxData(src, opts)
	if err == nil {
		t.Error("unexpected success for DynamicTx GasTipCap")
	}
	src = types.NewTx(&types.BlobTx{
		GasTipCap: uint256.NewInt(990),
		GasFeeCap: uint256.NewInt(9901),
		Gas:       0,
	})
	_, err = logic.copyTxData(src, opts)
	if err == nil {
		t.Error("unexpected success for DynamicTx GasFeeCap")
	}
}

func TestParseGasFee(t *testing.T) {
	cfg := &ethereum.ReplaceTxConfig{
		GasTipCapInc: "1wei",
		MaxGasTipCap: "2ether",
		GasFeeCapInc: "3gwei",
		MaxGasFeeCap: "10wei",
		GasPriceInc:  "20ether",
		MaxGasPrice:  "30gwei",
	}
	gas, err := parseGasFee(cfg)
	if err != nil {
		t.Fatal("must success")
	}
	if gas.GasTipCapInc.Uint64() != uint64(1) {
		t.Errorf("invalid GasTipCapInc: %v", gas.GasTipCapInc)
	}
	if gas.MaxGasTipCap.String() != "2000000000000000000" {
		t.Errorf("invalid MaxGasTipCap: %v", gas.MaxGasTipCap)
	}
	if gas.GasFeeCapInc.Uint64() != uint64(3000000000) {
		t.Errorf("invalid GasFeeCapInc: %v", gas.GasFeeCapInc)
	}
	if gas.MaxGasFeeCap.Uint64() != uint64(10) {
		t.Errorf("invalid MaxGasFeeCap: %v", gas.MaxGasFeeCap)
	}
	if gas.GasPriceInc.String() != "20000000000000000000" {
		t.Errorf("invalid GasPriceInc: %v", gas.GasPriceInc)
	}
	if gas.MaxGasPrice.Uint64() != uint64(30000000000) {
		t.Errorf("invalid MaxGasPrice: %v", gas.MaxGasPrice)
	}
}
