package client

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/avast/retry-go"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/hyperledger-labs/yui-relayer/log"
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

func (cl *ETHClient) GetTransactionReceipt(ctx context.Context, txHash common.Hash, enableDebugTrace bool) (rc *gethtypes.Receipt, revertReason string, err error) {
	var r *Receipt

	if err := cl.Raw().CallContext(ctx, &r, "eth_getTransactionReceipt", txHash); err != nil {
		return nil, "", err
	} else if r == nil {
		return nil, "", ethereum.NotFound
	} else if r.Status == gethtypes.ReceiptStatusSuccessful {
		return &r.Receipt, "", nil
	} else if r.HasRevertReason() {
		reason, err := r.GetRevertReason()
		if err != nil {
			// TODO: use more proper logger
			logger := log.GetLogger().WithModule("ethereum.chain")
			logger.Error("failed to get revert reason", err)
		}
		return &r.Receipt, reason, nil
	} else if enableDebugTrace {
		reason, err := cl.DebugTraceTransaction(ctx, txHash)
		if err != nil {
			// TODO: use more proper logger
			logger := log.GetLogger().WithModule("ethereum.chain")
			logger.Error("failed to call debug_traceTransaction", err)
		}
		return &r.Receipt, reason, nil
	} else {
		// TODO: use more proper logger
		logger := log.GetLogger().WithModule("ethereum.chain")
		logger.Info("tx execution failed but the reason couldn't be obtained", "tx_hash", txHash.Hex())
		return &r.Receipt, "", nil
	}
}

func (cl *ETHClient) WaitForReceiptAndGet(ctx context.Context, txHash common.Hash, enableDebugTrace bool) (*gethtypes.Receipt, string, error) {
	var receipt *gethtypes.Receipt
	var revertReason string
	err := retry.Do(
		func() error {
			rc, reason, err := cl.GetTransactionReceipt(ctx, txHash, enableDebugTrace)
			if err != nil {
				return err
			}
			receipt = rc
			revertReason = reason
			return nil
		},
		cl.option.retryOpts...,
	)
	if err != nil {
		return nil, "", err
	}
	return receipt, revertReason, nil
}

func (cl *ETHClient) DebugTraceTransaction(ctx context.Context, txHash common.Hash) (string, error) {
	var result *CallFrame
	if err := cl.Raw().CallContext(ctx, &result, "debug_traceTransaction", txHash, map[string]string{"tracer": "callTracer"}); err != nil {
		return "", err
	}
	revertReason, err := searchRevertReason(result)
	if err != nil {
		return "", err
	}
	return revertReason, nil
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

type callLog struct {
	Address common.Address `json:"address"`
	Topics  []common.Hash  `json:"topics"`
	Data    hexutil.Bytes  `json:"data"`
}

// see: https://github.com/ethereum/go-ethereum/blob/v1.12.0/eth/tracers/native/call.go#L44-L59
type CallFrame struct {
	Type         vm.OpCode       `json:"-"`
	From         common.Address  `json:"from"`
	Gas          uint64          `json:"gas"`
	GasUsed      uint64          `json:"gasUsed"`
	To           *common.Address `json:"to,omitempty" rlp:"optional"`
	Input        []byte          `json:"input" rlp:"optional"`
	Output       []byte          `json:"output,omitempty" rlp:"optional"`
	Error        string          `json:"error,omitempty" rlp:"optional"`
	RevertReason string          `json:"revertReason,omitempty"`
	Calls        []CallFrame     `json:"calls,omitempty" rlp:"optional"`
	Logs         []callLog       `json:"logs,omitempty" rlp:"optional"`
	// Placed at end on purpose. The RLP will be decoded to 0 instead of
	// nil if there are non-empty elements after in the struct.
	Value *big.Int `json:"value,omitempty" rlp:"optional"`
}

func searchRevertReason(result *CallFrame) (string, error) {
	if result.RevertReason != "" {
		return result.RevertReason, nil
	}
	for _, call := range result.Calls {
		reason, err := searchRevertReason(&call)
		if err == nil {
			return reason, nil
		}
	}
	return "", fmt.Errorf("revert reason not found")
}
