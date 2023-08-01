package client

import (
	"context"
	"fmt"
	"math/big"
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
	rpcClient *rpc.Client
	option    option
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
		rpcClient: rpcClient,
		Client:    ethclient.NewClient(rpcClient),
		option:    *opt,
	}, nil
}

func (cl *ETHClient) GetTransactionReceipt(ctx context.Context, txHash common.Hash, enableDebugTrace bool) (rc *gethtypes.Receipt, recoverable bool, err error) {
	var r *Receipt
	if err := cl.rpcClient.CallContext(ctx, &r, "eth_getTransactionReceipt", txHash); err != nil {
		return &r.Receipt, true, err
	}
	if r == nil {
		return nil, true, ethereum.NotFound
	} else if r.Status == gethtypes.ReceiptStatusSuccessful {
		return &r.Receipt, false, nil
	} else if r.HasRevertReason() {
		reason, err := r.GetRevertReason()
		return &r.Receipt, false, fmt.Errorf("revert: %v(parse-err=%v)", reason, err)
	} else {
		errPrefix := "failed to execute a transaction"
		if enableDebugTrace {
			to, revertReason, err := cl.DebugTraceTransaction(ctx, txHash)
			if err != nil {
				return &r.Receipt, false, fmt.Errorf("%s: %v, txHash: %s, debug_transaction error: %v", errPrefix, r, txHash, err)
			}
			return &r.Receipt, false, fmt.Errorf("%s: %v, contract: %s, txHash: %s, revertReason: %v", errPrefix, r, to, txHash, revertReason)
		} else {
			return &r.Receipt, false, fmt.Errorf("%s: %v", errPrefix, r)
		}
	}
}

func (cl *ETHClient) WaitForReceiptAndGet(ctx context.Context, tx *gethtypes.Transaction, enableDebugTrace bool) (*gethtypes.Receipt, error) {
	var receipt *gethtypes.Receipt
	err := retry.Do(
		func() error {
			rc, recoverable, err := cl.GetTransactionReceipt(ctx, tx.Hash(), enableDebugTrace)
			if err != nil {
				if recoverable {
					return err
				} else {
					return retry.Unrecoverable(err)
				}
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

func (cl *ETHClient) DebugTraceTransaction(ctx context.Context, txHash common.Hash) (string, string, error) {
	var result *Result
	if err := cl.rpcClient.CallContext(ctx, &result, "debug_traceTransaction", txHash, map[string]string{"tracer": "callTracer"}); err != nil {
		return "", "", err
	}
	to, revertReason, err := searchToAndReason(result)
	if err != nil {
		return "", "", err
	}
	return to, revertReason, nil
}

type Receipt struct {
	gethtypes.Receipt
	RevertReason []byte `json:"revertReason,omitempty"`
}

func (rc Receipt) HasRevertReason() bool {
	return len(rc.RevertReason) > 0
}

func (rc Receipt) GetRevertReason() (string, error) {
	return parseRevertReason(rc.RevertReason)
}

// A format of revertReason is:
// 4byte: Function selector for Error(string)
// 32byte: Data offset
// 32byte: String length
// Remains: String Data
func parseRevertReason(bz []byte) (string, error) {
	if l := len(bz); l == 0 {
		return "", nil
	} else if l < 68 {
		return "", fmt.Errorf("invalid length")
	}

	size := &big.Int{}
	size.SetBytes(bz[36:68])
	return string(bz[68 : 68+size.Int64()]), nil
}

type Result struct {
	Type         *string  `json:"type"`
	From         *string  `json:"from"`
	To           *string  `json:"to"`
	Value        *string  `json:"value"`
	Gas          *string  `json:"gas"`
	GasUsed      *string  `json:"gasUsed"`
	Input        *string  `json:"input"`
	Output       *string  `json:"output"`
	Error        *string  `json:"error"`
	RevertReason *string  `json:"revertReason"`
	Calls        []Result `json:"calls"`
}

func searchToAndReason(result *Result) (string, string, error) {
	if result.RevertReason != nil {
		return *result.To, *result.RevertReason, nil
	}
	for _, call := range result.Calls {
		to, reason, err := searchToAndReason(&call)
		if err == nil {
			return to, reason, nil
		}
	}
	return "", "", fmt.Errorf("revert reason not found")
}
