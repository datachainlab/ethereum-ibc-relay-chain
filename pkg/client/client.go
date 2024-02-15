package client

import (
	"context"
	"time"

	"github.com/avast/retry-go"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type ETHClient struct {
	*ethclient.Client
	option option
}

type Option func(*option)

type option struct {
	retryOpts []retry.Option
}

func DefaultOption() *option {
	return &option{
		retryOpts: []retry.Option{
			retry.Delay(1 * time.Second),
			retry.Attempts(10),
		},
	}
}

func WithRetryOption(rops ...retry.Option) Option {
	return func(opt *option) {
		opt.retryOpts = rops
	}
}

func NewETHClient(endpoint string, opts ...Option) (*ETHClient, error) {
	rpcClient, err := rpc.DialHTTP(endpoint)
	if err != nil {
		return nil, err
	}
	opt := DefaultOption()
	for _, o := range opts {
		o(opt)
	}
	return &ETHClient{
		Client: ethclient.NewClient(rpcClient),
		option: *opt,
	}, nil
}

func (cl *ETHClient) Raw() *rpc.Client {
	return cl.Client.Client()
}

func (cl *ETHClient) GetTransactionReceipt(ctx context.Context, txHash common.Hash) (rc *Receipt, err error) {
	var r *Receipt

	if err := cl.Raw().CallContext(ctx, &r, "eth_getTransactionReceipt", txHash); err != nil {
		return nil, err
	} else if r == nil {
		return nil, ethereum.NotFound
	} else {
		return r, nil
	}
}

func (cl *ETHClient) WaitForReceiptAndGet(ctx context.Context, txHash common.Hash) (*Receipt, error) {
	var receipt *Receipt
	err := retry.Do(
		func() error {
			rc, err := cl.GetTransactionReceipt(ctx, txHash)
			if err != nil {
				return err
			}
			receipt = rc
			return nil
		},
		cl.option.retryOpts...,
	)
	if err != nil {
		return nil, err
	}
	return receipt, nil
}

type Receipt struct {
	gethtypes.Receipt
	RevertReason []byte `json:"revertReason,omitempty"`
}

func (rc Receipt) HasRevertReason() bool {
	return len(rc.RevertReason) > 0
}

func (cl *ETHClient) EstimateGasFromTx(ctx context.Context, tx *gethtypes.Transaction) (uint64, error) {
	from, err := gethtypes.LatestSignerForChainID(tx.ChainId()).Sender(tx)
	if err != nil {
		return 0, err
	}
	to := tx.To()
	value := tx.Value()
	gasTipCap := tx.GasTipCap()
	gasFeeCap := tx.GasFeeCap()
	gasPrice := tx.GasPrice()
	data := tx.Data()
	accessList := tx.AccessList()
	callMsg := ethereum.CallMsg{
		From:       from,
		To:         to,
		GasPrice:   gasPrice,
		GasTipCap:  gasTipCap,
		GasFeeCap:  gasFeeCap,
		Value:      value,
		Data:       data,
		AccessList: accessList,
	}
	estimatedGas, err := cl.EstimateGas(ctx, callMsg)
	if err != nil {
		return 0, err
	}
	return estimatedGas, nil
}
