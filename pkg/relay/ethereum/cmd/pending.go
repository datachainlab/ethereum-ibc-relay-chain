package cmd

import (
	"context"
	"fmt"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/relay/ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
	"github.com/hyperledger-labs/yui-relayer/config"
	"github.com/spf13/cobra"
	"math/big"
	"time"
)

func pendingCmd(ctx *config.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pending",
		Short: "Manage ethereum pending transactions",
	}

	cmd.AddCommand(
		showPendingTxCmd(ctx),
		replacePendingTxCmd(ctx),
	)

	return cmd
}

func showPendingTxCmd(ctx *config.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show [chain-id]",
		Aliases: []string{"list"},
		Short:   "Show minimum nonce pending transactions sent by relayer",
		RunE: func(cmd *cobra.Command, args []string) error {
			chain, err := ctx.Config.GetChain(args[0])
			if err != nil {
				return err
			}
			ethChain := chain.Chain.(*ethereum.Chain)
			model := pendingModel{ethChain: ethChain}
			tx, err := model.ShowPendingTx()
			if err != nil {
				return err
			}
			json, err := tx.MarshalJSON()
			if err != nil {
				return err
			}
			fmt.Println(string(json))
			return nil
		},
	}
	return cmd
}

func replacePendingTxCmd(ctx *config.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "replace [chain-id]",
		Aliases: []string{"replace"},
		Short:   "Replace minimum nonce pending transaction sent by relayer",
		RunE: func(cmd *cobra.Command, args []string) error {
			chain, err := ctx.Config.GetChain(args[0])
			if err != nil {
				return err
			}
			ethChain := chain.Chain.(*ethereum.Chain)
			model := pendingModel{ethChain: ethChain}
			tx, err := model.ShowPendingTx()
			if err != nil {
				return err
			}
			ethereum.GetModuleLogger().Info("Pending transaction found", "txHash", tx.Hash())

			return model.ReplacePendingTx(tx.Hash())
		},
	}
	return cmd
}

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

type gasInfo struct {
	GasPriceInc  *big.Int
	MaxGasPrice  *big.Int
	GasTipCapInc *big.Int
	MaxGasTipCap *big.Int
	GasFeeCapInc *big.Int
	MaxGasFeeCap *big.Int
}

// Business Logic for pending command
type pendingModel struct {
	ethChain *ethereum.Chain
}

func (m *pendingModel) ShowPendingTx() (*types.Transaction, error) {
	txs, err := m.listPendingTx()
	if err != nil {
		return nil, err
	}
	tx := txs.GetMinimumNonceTransaction()
	if tx == nil {
		return tx, fmt.Errorf("no pending transaction was found")
	}
	return tx, nil
}

func (m *pendingModel) ReplacePendingTx(txHash common.Hash) error {
	replaceConfig := m.ethChain.Config().ReplaceConfig
	timer := time.NewTimer(time.Duration(replaceConfig.CheckInterval) * time.Second)
	defer timer.Stop()

	logger := ethereum.GetModuleLogger()
	start := time.Now()
	for {
		select {
		case <-timer.C:
			tx, isPending, err := m.ethChain.Client().Client.TransactionByHash(context.Background(), txHash)
			if err != nil {
				return err
			}
			if !isPending {
				logger.Info("tx is not pending", "txHash", tx.Hash())
				return nil
			}
			if time.Now().After(start.Add(time.Duration(replaceConfig.PendingDurationToReplace) * time.Second)) {
				logger.Info("try to replace pending transaction", "txHash", tx.Hash())
				return m.replacePendingTx(tx)
			}
			logger.Info("tx is still pending", "txHash", tx.Hash())
			timer.Reset(time.Duration(replaceConfig.CheckInterval) * time.Second)
		}
	}
}

func (m *pendingModel) listPendingTx() (pendingTransactions, error) {
	fromAddress := m.ethChain.CallOpts(context.Background(), 0).From
	var value *txPoolContent
	if err := m.ethChain.Client().Client.Client().Call(&value, "txpool_contentFrom", fromAddress); err != nil {
		return nil, err
	}
	if value == nil {
		return nil, nil
	}
	return value.Pending, nil
}

func (m *pendingModel) replacePendingTx(tx *types.Transaction) error {
	client := m.ethChain.Client()
	cfg := *m.ethChain.Config().ReplaceConfig

	// string -> big.Int
	gasPriceInc, err := m.strToBig(cfg.GasPriceInc)
	if err != nil {
		return err
	}
	maxGasPrice, err := m.strToBig(cfg.MaxGasPrice)
	if err != nil {
		return err
	}
	gasTipCapInc, err := m.strToBig(cfg.GasTipCapInc)
	if err != nil {
		return err
	}
	maxGasTipCap, err := m.strToBig(cfg.MaxGasTipCap)
	if err != nil {
		return err
	}
	gasFeeCapInc, err := m.strToBig(cfg.GasFeeCapInc)
	if err != nil {
		return err
	}
	maxGasFeeCap, err := m.strToBig(cfg.MaxGasFeeCap)
	if err != nil {
		return err
	}

	txData, err := m.copyTxData(tx, gasInfo{
		GasPriceInc:  gasPriceInc,
		MaxGasPrice:  maxGasPrice,
		GasTipCapInc: gasTipCapInc,
		MaxGasTipCap: maxGasTipCap,
		GasFeeCapInc: gasFeeCapInc,
		MaxGasFeeCap: maxGasFeeCap,
	})
	if err != nil {
		return err
	}
	rawTx := types.NewTx(txData)
	txOpts := m.ethChain.TxOpts(context.Background())
	newTx, err := txOpts.Signer(txOpts.From, rawTx)
	if err != nil {
		return err
	}
	ctx := context.Background()
	if err = client.Client.SendTransaction(ctx, newTx); err != nil {
		return err
	}
	logger := ethereum.GetModuleLogger()
	logger.Info("wait for transaction receipt", "txHash", newTx.Hash())
	if receipt, revertReason, err := client.WaitForReceiptAndGet(ctx, newTx.Hash(), m.ethChain.Config().EnableDebugTrace); err != nil {
		return err
	} else if receipt.Status == types.ReceiptStatusFailed {
		return fmt.Errorf("tx execution failed: revertReason=%s", revertReason)
	}
	logger.Info("transaction success", "txHash", newTx.Hash())
	return nil
}

func (m *pendingModel) copyTxData(src *types.Transaction, gas gasInfo) (types.TxData, error) {
	switch src.Type() {
	case types.AccessListTxType:
		gasPrice := m.add(src.GasPrice(), gas.GasPriceInc)
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
		gasTipCap := m.add(src.GasTipCap(), gas.GasTipCapInc)
		if gasTipCap.Cmp(gas.MaxGasTipCap) > 0 {
			return nil, fmt.Errorf("gasTipCap > max : DynamicFeeTx value=%v,max=%v", gasTipCap, gas.MaxGasTipCap)
		}
		gasFeeCap := m.add(src.GasFeeCap(), gas.GasFeeCapInc)
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
		gasTipCap := m.add(src.GasTipCap(), gas.GasTipCapInc)
		if gasTipCap.Cmp(gas.MaxGasTipCap) > 0 {
			return nil, fmt.Errorf("gasTipCap > max : BlobTx value=%v,max=%v", gasTipCap, gas.MaxGasTipCap)
		}
		gasFeeCap := m.add(src.GasFeeCap(), gas.GasFeeCapInc)
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
		gasPrice := m.add(src.GasPrice(), gas.GasPriceInc)
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

func (m *pendingModel) add(x *big.Int, y *big.Int) *big.Int {
	return new(big.Int).Add(x, y)
}

func (m *pendingModel) strToBig(amount string) (*big.Int, error) {
	value, result := new(big.Int).SetString(amount, 10)
	if !result {
		return nil, fmt.Errorf("invalid amount %s", amount)
	}
	return value, nil
}
