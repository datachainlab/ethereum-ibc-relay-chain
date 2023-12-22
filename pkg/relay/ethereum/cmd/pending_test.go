package cmd

import (
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/relay/ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
	"math/big"
	"testing"
)

func TestSuccessCopyTxData(t *testing.T) {
	model := pendingModel{}
	opts := ethereum.ReplaceConfig{
		GasTipCapInc: common.Bytes2Hex(big.NewInt(10).Bytes()),
		MaxGasTipCap: common.Bytes2Hex(big.NewInt(1000).Bytes()),
		GasFeeCapInc: common.Bytes2Hex(big.NewInt(100).Bytes()),
		MaxGasFeeCap: common.Bytes2Hex(big.NewInt(10000).Bytes()),
		GasPriceInc:  common.Bytes2Hex(big.NewInt(200).Bytes()),
		MaxGasPrice:  common.Bytes2Hex(big.NewInt(20000).Bytes()),
	}

	// Legacy Tx
	src := types.NewTx(&types.LegacyTx{
		GasPrice: big.NewInt(19800),
		Gas:      0,
	})
	txData, err := model.copyTxData(src, opts)
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
	txData, err = model.copyTxData(src, opts)
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
	txData, err = model.copyTxData(src, opts)
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
	txData, err = model.copyTxData(src, opts)
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
	model := pendingModel{}
	opts := ethereum.ReplaceConfig{
		GasTipCapInc: common.Bytes2Hex(big.NewInt(10).Bytes()),
		MaxGasTipCap: common.Bytes2Hex(big.NewInt(1000).Bytes()),
		GasFeeCapInc: common.Bytes2Hex(big.NewInt(100).Bytes()),
		MaxGasFeeCap: common.Bytes2Hex(big.NewInt(10000).Bytes()),
		GasPriceInc:  common.Bytes2Hex(big.NewInt(200).Bytes()),
		MaxGasPrice:  common.Bytes2Hex(big.NewInt(20000).Bytes()),
	}

	// Legacy Tx
	src := types.NewTx(&types.LegacyTx{
		GasPrice: big.NewInt(19801),
		Gas:      0,
	})
	_, err := model.copyTxData(src, opts)
	if err == nil {
		t.Error("unexpected success for LegacyTx")
	}

	// AccessList Tx
	src = types.NewTx(&types.AccessListTx{
		GasPrice: big.NewInt(19801),
		Gas:      0,
	})
	_, err = model.copyTxData(src, opts)
	if err == nil {
		t.Error("unexpected success for AccessListTx")
	}

	// Dynamic Tx
	src = types.NewTx(&types.DynamicFeeTx{
		GasTipCap: big.NewInt(991),
		GasFeeCap: big.NewInt(1000),
		Gas:       0,
	})
	_, err = model.copyTxData(src, opts)
	if err == nil {
		t.Error("unexpected success for DynamicTx GasTipCap")
	}
	src = types.NewTx(&types.DynamicFeeTx{
		GasTipCap: big.NewInt(1),
		GasFeeCap: big.NewInt(9901),
		Gas:       0,
	})
	_, err = model.copyTxData(src, opts)
	if err == nil {
		t.Error("unexpected success for DynamicTx GasFeeCap")
	}

	// Blob Tx
	src = types.NewTx(&types.BlobTx{
		GasTipCap: uint256.NewInt(991),
		GasFeeCap: uint256.NewInt(9900),
		Gas:       0,
	})
	_, err = model.copyTxData(src, opts)
	if err == nil {
		t.Error("unexpected success for DynamicTx GasTipCap")
	}
	src = types.NewTx(&types.BlobTx{
		GasTipCap: uint256.NewInt(990),
		GasFeeCap: uint256.NewInt(9901),
		Gas:       0,
	})
	_, err = model.copyTxData(src, opts)
	if err == nil {
		t.Error("unexpected success for DynamicTx GasFeeCap")
	}
}
