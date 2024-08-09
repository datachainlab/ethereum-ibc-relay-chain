package ethereum

import (
	"context"
	"fmt"

	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/contract/iibcchannelupgradablemodule"
	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
)

func (c *Chain) ProposeUpgrade(
	ctx context.Context,
	appAddr common.Address,
	portID string,
	channelID string,
	upgradeFields iibcchannelupgradablemodule.UpgradeFieldsData,
	timeout iibcchannelupgradablemodule.TimeoutData,
) error {
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
	if err != nil {
		return err
	}

	if receipt, err := c.client.WaitForReceiptAndGet(ctx, tx.Hash()); err != nil {
		return err
	} else if receipt.Status == gethtypes.ReceiptStatusFailed {
		return fmt.Errorf("tx execution reverted")
	}

	return nil
}

func (c *Chain) AllowTransitionToFlushComplete(
	ctx context.Context,
	appAddr common.Address,
	portID string,
	channelID string,
	upgradeSequence uint64,
) error {
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

	if receipt, err := c.client.WaitForReceiptAndGet(ctx, tx.Hash()); err != nil {
		return err
	} else if receipt.Status == gethtypes.ReceiptStatusFailed {
		return fmt.Errorf("tx execution reverted")
	}

	return nil
}
