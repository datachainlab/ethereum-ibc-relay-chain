package client

import (
	"context"
	"encoding/json"
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

// RPCTransaction represents a transaction that will serialize to the RPC representation of a transaction
type RPCTransaction struct {
	BlockHash           *common.Hash      `json:"blockHash"`
	BlockNumber         *hexutil.Big      `json:"blockNumber"`
	From                common.Address    `json:"from"`
	Gas                 hexutil.Uint64    `json:"gas"`
	GasPrice            *hexutil.Big      `json:"gasPrice"`
	GasFeeCap           *hexutil.Big      `json:"maxFeePerGas,omitempty"`
	GasTipCap           *hexutil.Big      `json:"maxPriorityFeePerGas,omitempty"`
	MaxFeePerBlobGas    *hexutil.Big      `json:"maxFeePerBlobGas,omitempty"`
	Hash                common.Hash       `json:"hash"`
	Input               hexutil.Bytes     `json:"input"`
	Nonce               hexutil.Uint64    `json:"nonce"`
	To                  *common.Address   `json:"to"`
	TransactionIndex    *hexutil.Uint64   `json:"transactionIndex"`
	Value               *hexutil.Big      `json:"value"`
	Type                hexutil.Uint64    `json:"type"`
	Accesses            *gethtypes.AccessList `json:"accessList,omitempty"`
	ChainID             *hexutil.Big      `json:"chainId,omitempty"`
	BlobVersionedHashes []common.Hash     `json:"blobVersionedHashes,omitempty"`
	V                   *hexutil.Big      `json:"v"`
	R                   *hexutil.Big      `json:"r"`
	S                   *hexutil.Big      `json:"s"`
	YParity             *hexutil.Uint64   `json:"yParity,omitempty"`
}

// wrapping interface of ethclient.Client struct
type IETHClient interface {
	Inner() *ethclient.Client
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*gethtypes.Header, error);
	FeeHistory(ctx context.Context, blockCount uint64, lastBlock *big.Int, rewardPercentiles []float64) (*ethereum.FeeHistory, error)
	ContentFrom(ctx context.Context, address common.Address) (map[string]map[string]*RPCTransaction, error);
}

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
	return NewETHClientWith(ethclient.NewClient(rpcClient), opts...)
}

func NewETHClientWith(cli *ethclient.Client, opts ...Option) (*ETHClient, error) {
	opt := DefaultOption()
	for _, o := range opts {
		o(opt)
	}
	return &ETHClient{
		Client: cli,
		option: *opt,
	}, nil
}

func (cl *ETHClient) Raw() *rpc.Client {
	return cl.Client.Client()
}

// implements IETHClient
func (cl *ETHClient) Inner() *ethclient.Client {
	return cl.Client
}
func (cl *ETHClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return cl.Client.SuggestGasPrice(ctx)
}
func (cl *ETHClient) HeaderByNumber(ctx context.Context, number *big.Int) (*gethtypes.Header, error) {
	return cl.Client.HeaderByNumber(ctx, number)
}
func (cl *ETHClient) FeeHistory(ctx context.Context, blockCount uint64, lastBlock *big.Int, rewardPercentiles []float64) (*ethereum.FeeHistory, error) {
	return cl.Client.FeeHistory(ctx, blockCount, lastBlock, rewardPercentiles)
}

// ContentFrom calls `txpool_contentFrom` of the Ethereum RPC
func (cl *ETHClient) ContentFrom(ctx context.Context, address common.Address) (map[string]map[string]*RPCTransaction, error) {
	var res map[string]map[string]*RPCTransaction
	if err := cl.Client.Client().CallContext(ctx, &res, "txpool_contentFrom", address); err != nil {
		return nil, err
	}
	return res, nil
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

func (cl *ETHClient) DebugTraceTransaction(ctx context.Context, txHash common.Hash) (CallFrame, error) {
	var callFrame CallFrame
	err := cl.Raw().CallContext(ctx, &callFrame, "debug_traceTransaction", txHash, map[string]string{"tracer": "callTracer"})
	return callFrame, err
}

type Receipt struct {
	gethtypes.Receipt
	RevertReason []byte `json:"revertReason,omitempty"`
}

func (rc Receipt) HasRevertReason() bool {
	return len(rc.RevertReason) > 0
}

type CallLog struct {
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
	Logs         []CallLog       `json:"logs,omitempty" rlp:"optional"`
	// Placed at end on purpose. The RLP will be decoded to 0 instead of
	// nil if there are non-empty elements after in the struct.
	Value *big.Int `json:"value,omitempty" rlp:"optional"`
}

// UnmarshalJSON unmarshals from JSON.
func (c *CallFrame) UnmarshalJSON(input []byte) error {
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
		Calls        []CallFrame     `json:"calls,omitempty" rlp:"optional"`
		Logs         []CallLog       `json:"logs,omitempty" rlp:"optional"`
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

func (cl *ETHClient) EstimateGasFromTx(ctx context.Context, tx *gethtypes.Transaction) (uint64, error) {
	from, err := gethtypes.LatestSignerForChainID(tx.ChainId()).Sender(tx)
	if err != nil {
		return 0, err
	}

	callMsg := ethereum.CallMsg{
		From:  from,
		To:    tx.To(),
		Value: tx.Value(),
		Data:  tx.Data(),
	}

	switch tx.Type() {
	case gethtypes.BlobTxType:
		callMsg.GasTipCap = tx.GasTipCap()
		callMsg.GasFeeCap = tx.GasFeeCap()
		callMsg.AccessList = tx.AccessList()
		callMsg.BlobGasFeeCap = tx.BlobGasFeeCap()
		callMsg.BlobHashes = tx.BlobHashes()
	case gethtypes.DynamicFeeTxType:
		callMsg.GasTipCap = tx.GasTipCap()
		callMsg.GasFeeCap = tx.GasFeeCap()
		callMsg.AccessList = tx.AccessList()
	case gethtypes.AccessListTxType:
		callMsg.GasPrice = tx.GasPrice()
		callMsg.AccessList = tx.AccessList()
	case gethtypes.LegacyTxType:
		callMsg.GasPrice = tx.GasPrice()
	default:
		panic("unsupported tx type")
	}

	return cl.EstimateGas(ctx, callMsg)
}
