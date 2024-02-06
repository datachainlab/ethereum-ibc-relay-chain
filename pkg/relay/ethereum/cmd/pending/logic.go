package pending

import (
	"context"
	"fmt"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/relay/ethereum"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
	"math/big"
	"time"
)

type pendingTransactions map[uint64]*types.Transaction

func (p pendingTransactions) GetMinimumNonceTransaction() *types.Transaction {
	minNonce := uint64(0)
	var minValue *types.Transaction = nil
	first := true
	for k, v := range p {
		if first || minNonce > k {
			minNonce = k
			minValue = v
			first = false
		}
	}
	return minValue
}

type txPoolContent struct {
	Pending pendingTransactions `json:"pending"`
}

type gasFees struct {
	GasPriceInc  *big.Int
	MaxGasPrice  *big.Int
	GasTipCapInc *big.Int
	MaxGasTipCap *big.Int
	GasFeeCapInc *big.Int
	MaxGasFeeCap *big.Int
}

type Logic struct {
	ethChain *ethereum.Chain
}

func NewLogic(ethChain *ethereum.Chain) *Logic {
	return &Logic{
		ethChain: ethChain,
	}
}

func (m *Logic) ShowPendingTx(ctx context.Context) (*types.Transaction, error) {
	txs, err := m.listPendingTx(ctx)
	if err != nil {
		return nil, err
	}
	tx := txs.GetMinimumNonceTransaction()
	if tx == nil {
		return tx, fmt.Errorf("no pending transaction was found")
	}
	return tx, nil
}

func (m *Logic) ReplacePendingTx(ctx context.Context, txHash common.Hash) error {
	replaceConfig := m.ethChain.Config().ReplaceTxConfig
	timer := time.NewTimer(time.Duration(replaceConfig.CheckInterval) * time.Second)
	defer timer.Stop()

	logger := ethereum.GetModuleLogger()
	start := time.Now()
	for {
		select {
		case <-timer.C:
			tx, isPending, err := m.ethChain.Client().Client.TransactionByHash(ctx, txHash)
			if err != nil {
				return err
			}
			if !isPending {
				logger.Info("tx is not pending", "txHash", tx.Hash())
				return nil
			}
			if time.Now().After(start.Add(time.Duration(replaceConfig.PendingDurationToReplace) * time.Second)) {
				logger.Info("try to replace pending transaction", "txHash", tx.Hash())
				return m.replacePendingTx(ctx, tx)
			}
			logger.Info("tx is still pending", "txHash", tx.Hash())
			timer.Reset(time.Duration(replaceConfig.CheckInterval) * time.Second)
		}
	}
}

func (m *Logic) listPendingTx(ctx context.Context) (pendingTransactions, error) {
	fromAddress := m.ethChain.CallOpts(ctx, 0).From
	var value *txPoolContent
	if err := m.ethChain.Client().Client.Client().Call(&value, "txpool_contentFrom", fromAddress); err != nil {
		return nil, err
	}
	if value == nil {
		return nil, nil
	}
	return value.Pending, nil
}

func (m *Logic) replacePendingTx(ctx context.Context, tx *types.Transaction) error {
	client := m.ethChain.Client()
	cfg := m.ethChain.Config().ReplaceTxConfig
	if cfg != nil {
		return fmt.Errorf("\"replace_tx_config\" in chain config is required to replace tx")
	}

	gasToReplace, err := parseGasInfo(cfg)
	if err != nil {
		return err
	}

	txData, err := m.copyTxData(tx, gasToReplace)
	if err != nil {
		return err
	}
	txOpts, err := m.ethChain.TxOpts(ctx)
	if err != nil {
		return err
	}
	newTx, err := txOpts.Signer(txOpts.From, types.NewTx(txData))
	if err != nil {
		return err
	}
	if err = client.Client.SendTransaction(ctx, newTx); err != nil {
		return err
	}

	logger := ethereum.GetModuleLogger()
	if receipt, revertReason, err := client.WaitForReceiptAndGet(ctx, newTx.Hash(), m.ethChain.Config().EnableDebugTrace); err != nil {
		return fmt.Errorf("replace tx error: txHash=%s, err=%v", newTx.Hash(), err)
	} else if receipt.Status == types.ReceiptStatusFailed {
		return fmt.Errorf("replace tx failed: txHash=%s, revertReason=%s", newTx.Hash(), revertReason)
	}
	logger.Info("replace tx success", "txHash", newTx.Hash())
	return nil
}

func (m *Logic) copyTxData(src *types.Transaction, gas *gasFees) (types.TxData, error) {
	switch src.Type() {
	case types.AccessListTxType:
		gasPrice := add(src.GasPrice(), gas.GasPriceInc)
		if gasPrice.Cmp(gas.MaxGasPrice) > 0 {
			return nil, fmt.Errorf("gasPrice > max : AccessListTx value=%v,max=%v", gasPrice, gas.MaxGasPrice)
		}
		return &types.AccessListTx{
			Nonce:    src.Nonce(),
			GasPrice: gasPrice,
			Gas:      src.Gas(),
			To:       src.To(),
			Value:    src.Value(),
			Data:     src.Data(),
		}, nil
	case types.DynamicFeeTxType:
		gasTipCap := add(src.GasTipCap(), gas.GasTipCapInc)
		if gasTipCap.Cmp(gas.MaxGasTipCap) > 0 {
			return nil, fmt.Errorf("gasTipCap > max : DynamicFeeTx value=%v,max=%v", gasTipCap, gas.MaxGasTipCap)
		}
		gasFeeCap := add(src.GasFeeCap(), gas.GasFeeCapInc)
		if gasFeeCap.Cmp(gas.MaxGasFeeCap) > 0 {
			return nil, fmt.Errorf("gasFeeCap > max : DynamicFeeTx value=%v,max=%v", gasFeeCap, gas.MaxGasFeeCap)
		}
		return &types.DynamicFeeTx{
			ChainID:    src.ChainId(),
			Nonce:      src.Nonce(),
			GasTipCap:  gasTipCap,
			GasFeeCap:  gasFeeCap,
			Gas:        src.Gas(),
			To:         src.To(),
			Value:      src.Value(),
			Data:       src.Data(),
			AccessList: src.AccessList(),
		}, nil
	case types.BlobTxType:
		gasTipCap := add(src.GasTipCap(), gas.GasTipCapInc)
		if gasTipCap.Cmp(gas.MaxGasTipCap) > 0 {
			return nil, fmt.Errorf("gasTipCap > max : BlobTx value=%v,max=%v", gasTipCap, gas.MaxGasTipCap)
		}
		gasFeeCap := add(src.GasFeeCap(), gas.GasFeeCapInc)
		if gasFeeCap.Cmp(gas.MaxGasFeeCap) > 0 {
			return nil, fmt.Errorf("gasFeeCap > max : BlobTx value=%v,max=%v", gasFeeCap, gas.MaxGasFeeCap)
		}
		return &types.BlobTx{
			ChainID:    uint256.MustFromBig(src.ChainId()),
			Nonce:      src.Nonce(),
			GasTipCap:  uint256.MustFromBig(gasTipCap),
			GasFeeCap:  uint256.MustFromBig(gasFeeCap),
			Gas:        src.Gas(),
			To:         src.To(),
			Value:      uint256.MustFromBig(src.Value()),
			Data:       src.Data(),
			AccessList: src.AccessList(),
			BlobFeeCap: uint256.MustFromBig(src.BlobGasFeeCap()),
			BlobHashes: src.BlobHashes(),
		}, nil

	default:
		gasPrice := add(src.GasPrice(), gas.GasPriceInc)
		if gasPrice.Cmp(gas.MaxGasPrice) > 0 {
			return nil, fmt.Errorf("gasPrice > max : LegacyTx value=%v,max=%v", gasPrice, gas.MaxGasPrice)
		}
		return &types.LegacyTx{
			Nonce:    src.Nonce(),
			GasPrice: gasPrice,
			Gas:      src.Gas(),
			To:       src.To(),
			Value:    src.Value(),
			Data:     src.Data(),
		}, nil
	}
}

func add(x *big.Int, y *big.Int) *big.Int {
	return new(big.Int).Add(x, y)
}

func parseGasInfo(cfg *ethereum.ReplaceTxConfig) (*gasFees, error) {
	gasPriceInc, err := utils.ParseEtherAmount(cfg.GasPriceInc)
	if err != nil {
		return nil, err
	}
	maxGasPrice, err := utils.ParseEtherAmount(cfg.MaxGasPrice)
	if err != nil {
		return nil, err
	}
	gasTipCapInc, err := utils.ParseEtherAmount(cfg.GasTipCapInc)
	if err != nil {
		return nil, err
	}
	maxGasTipCap, err := utils.ParseEtherAmount(cfg.MaxGasTipCap)
	if err != nil {
		return nil, err
	}
	gasFeeCapInc, err := utils.ParseEtherAmount(cfg.GasFeeCapInc)
	if err != nil {
		return nil, err
	}
	maxGasFeeCap, err := utils.ParseEtherAmount(cfg.MaxGasFeeCap)
	if err != nil {
		return nil, err
	}

	return &gasFees{
		GasPriceInc:  gasPriceInc,
		MaxGasPrice:  maxGasPrice,
		GasTipCapInc: gasTipCapInc,
		MaxGasTipCap: maxGasTipCap,
		GasFeeCapInc: gasFeeCapInc,
		MaxGasFeeCap: maxGasFeeCap,
	}, nil
}
