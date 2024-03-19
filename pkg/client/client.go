package client

import (
	"context"
	"encoding/json"
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

func (cl *ETHClient) DebugTraceTransaction(ctx context.Context, txHash common.Hash) (string, error) {
	var result *callFrame
	if err := cl.Raw().CallContext(ctx, &result, "debug_traceTransaction", txHash, map[string]string{"tracer": "callTracer"}); err != nil {
		return "", err
	}
	revertReason, err := searchRevertReason(result)
	if err != nil {
		return "", err
	}
	return revertReason, nil
}

// see: https://github.com/ethereum/go-ethereum/blob/v1.12.0/eth/tracers/native/call.go#L44-L59
type callFrame struct {
	Type         vm.OpCode       `json:"-"`
	From         common.Address  `json:"from"`
	Gas          uint64          `json:"gas"`
	GasUsed      uint64          `json:"gasUsed"`
	To           *common.Address `json:"to,omitempty" rlp:"optional"`
	Input        []byte          `json:"input" rlp:"optional"`
	Output       []byte          `json:"output,omitempty" rlp:"optional"`
	Error        string          `json:"error,omitempty" rlp:"optional"`
	RevertReason string          `json:"revertReason,omitempty"`
	Calls        []callFrame     `json:"calls,omitempty" rlp:"optional"`
	Logs         []callLog       `json:"logs,omitempty" rlp:"optional"`
	// Placed at end on purpose. The RLP will be decoded to 0 instead of
	// nil if there are non-empty elements after in the struct.
	Value *big.Int `json:"value,omitempty" rlp:"optional"`
}

type callLog struct {
	Address common.Address `json:"address"`
	Topics  []common.Hash  `json:"topics"`
	Data    hexutil.Bytes  `json:"data"`
}

// UnmarshalJSON unmarshals from JSON.
func (c *callFrame) UnmarshalJSON(input []byte) error {
	type callFrame0 struct {
		Type         *vm.OpCode      `json:"-"`
		From         *common.Address `json:"from"`
		Gas          *hexutil.Uint64 `json:"gas"`
		GasUsed      *hexutil.Uint64 `json:"gasUsed"`
		To           *common.Address `json:"to,omitempty" rlp:"optional"`
		Input        *hexutil.Bytes  `json:"input" rlp:"optional"`
		Output       *hexutil.Bytes  `json:"output,omitempty" rlp:"optional"`
		Error        *string         `json:"error,omitempty" rlp:"optional"`
		RevertReason *string         `json:"revertReason,omitempty"`
		Calls        []callFrame     `json:"calls,omitempty" rlp:"optional"`
		Logs         []callLog       `json:"logs,omitempty" rlp:"optional"`
		Value        *hexutil.Big    `json:"value,omitempty" rlp:"optional"`
	}
	var dec callFrame0
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.Type != nil {
		c.Type = *dec.Type
	}
	if dec.From != nil {
		c.From = *dec.From
	}
	if dec.Gas != nil {
		c.Gas = uint64(*dec.Gas)
	}
	if dec.GasUsed != nil {
		c.GasUsed = uint64(*dec.GasUsed)
	}
	if dec.To != nil {
		c.To = dec.To
	}
	if dec.Input != nil {
		c.Input = *dec.Input
	}
	if dec.Output != nil {
		c.Output = *dec.Output
	}
	if dec.Error != nil {
		c.Error = *dec.Error
	}
	if dec.RevertReason != nil {
		c.RevertReason = *dec.RevertReason
	}
	if dec.Calls != nil {
		c.Calls = dec.Calls
	}
	if dec.Logs != nil {
		c.Logs = dec.Logs
	}
	if dec.Value != nil {
		c.Value = (*big.Int)(dec.Value)
	}
	return nil
}

func searchRevertReason(result *callFrame) (string, error) {
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
