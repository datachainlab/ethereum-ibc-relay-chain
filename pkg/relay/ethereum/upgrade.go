package ethereum

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/contract/iibcchannelupgradablemodule"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/contract/iibccontractupgradablemodule"
	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/hyperledger-labs/yui-relayer/log"
)

func (c *Chain) ProposeUpgrade(
	ctx context.Context,
	portID string,
	channelID string,
	upgradeFields iibcchannelupgradablemodule.UpgradeFieldsData,
	timeout iibcchannelupgradablemodule.TimeoutData,
) error {
	logger := c.GetChainLogger()
	logger = &log.RelayLogger{Logger: logger.With(
		logAttrPortID, portID,
		logAttrChannelID, channelID,
		logAttrUpgradeFields, upgradeFields,
		logAttrTimeout, timeout,
	)}

	appAddr, err := c.ibcHandler.GetIBCModuleByChannel(c.CallOpts(ctx, 0), portID, channelID)
	if err != nil {
		return err
	}

	mockApp, err := iibcchannelupgradablemodule.NewIibcchannelupgradablemodule(
		appAddr,
		c.client.Client,
	)
	if err != nil {
		return nil
	}

	txOpts, err := c.TxOpts(ctx, true)
	if err != nil {
		return err
	}

	tx, err := mockApp.ProposeUpgrade(
		txOpts,
		portID,
		channelID,
		upgradeFields,
		timeout,
	)

	return processSendTxResult(ctx, logger, c, tx, err)
}

func (c *Chain) AllowTransitionToFlushComplete(
	ctx context.Context,
	portID string,
	channelID string,
	upgradeSequence uint64,
) error {
	logger := c.GetChainLogger()
	logger = &log.RelayLogger{Logger: logger.With(
		logAttrPortID, portID,
		logAttrChannelID, channelID,
		logAttrUpgradeSequence, upgradeSequence,
	)}

	appAddr, err := c.ibcHandler.GetIBCModuleByChannel(c.CallOpts(ctx, 0), portID, channelID)
	if err != nil {
		return err
	}

	mockApp, err := iibcchannelupgradablemodule.NewIibcchannelupgradablemodule(
		appAddr,
		c.client.Client,
	)
	if err != nil {
		return nil
	}

	txOpts, err := c.TxOpts(ctx, true)
	if err != nil {
		return err
	}

	tx, err := mockApp.AllowTransitionToFlushComplete(
		txOpts,
		portID,
		channelID,
		upgradeSequence,
	)
	if err != nil {
		return err
	}

	return processSendTxResult(ctx, logger, c, tx, err)
}

func (c *Chain) ProposeAppVersion(
	ctx context.Context,
	portID string,
	channelID string,
	version string,
	implementation common.Address,
	initialCalldata []byte,
) error {
	logger := c.GetChainLogger()
	logger = &log.RelayLogger{Logger: logger.With(
		logAttrPortID, portID,
		logAttrChannelID, channelID,
		logAttrVersion, version,
		logAttrImplementation, implementation.Hex(),
		logAttrInitialCalldata, hex.EncodeToString(initialCalldata),
	)}

	appAddr, err := c.ibcHandler.GetIBCModuleByChannel(c.CallOpts(ctx, 0), portID, channelID)
	if err != nil {
		return err
	}

	mockApp, err := iibccontractupgradablemodule.NewIibccontractupgradablemodule(
		appAddr,
		c.client.Client,
	)
	if err != nil {
		return nil
	}

	txOpts, err := c.TxOpts(ctx, true)
	if err != nil {
		return err
	}

	tx, err := mockApp.ProposeAppVersion(
		txOpts,
		version,
		iibccontractupgradablemodule.IIBCContractUpgradableModuleAppInfo{
			Implementation:  implementation,
			InitialCalldata: initialCalldata,
		},
	)

	return processSendTxResult(ctx, logger, c, tx, err)
}

func processSendTxResult(ctx context.Context, logger *log.RelayLogger, c *Chain, tx *gethtypes.Transaction, err error) error {
	if err != nil {
		if revertReason, returnData, err := c.getRevertReasonFromRpcError(err); err != nil {
			logger = &log.RelayLogger{Logger: logger.With(
				logAttrRawErrorData, hex.EncodeToString(returnData),
			)}
			logger.Error("failed to get revert reason from RPC error", err)
		} else {
			logger = &log.RelayLogger{Logger: logger.With(
				logAttrRevertReason, revertReason,
				logAttrRawErrorData, hex.EncodeToString(returnData),
			)}
		}
		logger.Error("failed to send tx", err)
		return err
	}

	if rawTxData, err := tx.MarshalBinary(); err != nil {
		logger.Error("failed to encode tx", err)
	} else {
		logger = &log.RelayLogger{Logger: logger.With(
			logAttrRawTxData, hex.EncodeToString(rawTxData),
		)}
	}

	if receipt, err := c.client.WaitForReceiptAndGet(ctx, tx.Hash()); err != nil {
		logger.Error("failed to wait for tx receipt", err)
		return err
	} else if receipt.Status == gethtypes.ReceiptStatusFailed {
		if revertReason, returnData, err := c.getRevertReasonFromReceipt(ctx, receipt); err != nil {
			logger = &log.RelayLogger{Logger: logger.With(
				logAttrRawErrorData, hex.EncodeToString(returnData),
			)}
			logger.Error("failed to get revert reason from receipt", err)
		} else {
			logger = &log.RelayLogger{Logger: logger.With(
				logAttrRevertReason, revertReason,
				logAttrRawErrorData, hex.EncodeToString(returnData),
			)}
		}
		return fmt.Errorf("tx execution reverted")
	}

	return nil
}
