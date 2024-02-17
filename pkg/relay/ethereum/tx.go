package ethereum

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	math "math"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	conntypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	chantypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/hyperledger-labs/yui-relayer/core"
	"github.com/hyperledger-labs/yui-relayer/log"

	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/client"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/contract/ibchandler"
)

// SendMsgs sends msgs to the chain
func (c *Chain) SendMsgs(msgs []sdk.Msg) ([]core.MsgID, error) {
	ctx := context.TODO()
	// if src's connection is OPEN, dst's connection is OPEN or TRYOPEN, so we can skip to update client commitments
	skipUpdateClientCommitment, err := c.confirmConnectionOpened(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to confirm connection opened: %w", err)
	}
	logger := c.GetChainLogger()
	var msgIDs []core.MsgID
	for i, msg := range msgs {
		var (
			tx  *gethtypes.Transaction
			err error
		)
		opts, err := c.TxOpts(ctx)
		if err != nil {
			return nil, err
		}
		opts.GasLimit = math.MaxUint64
		opts.NoSend = true
		tx, err = c.SendTx(opts, msg, skipUpdateClientCommitment)
		if err != nil {
			logger.Error("failed to send msg / NoSend: true", err)
			return nil, err
		}
		estimatedGas, err := c.client.EstimateGasFromTx(ctx, tx)
		if err != nil {
			logger.Error("failed to estimate gas", err)
			if c.config.EnableDebugTrace {
				revertReason, err := c.DebugTraceTransaction(ctx, tx.Hash())
				if err != nil {
					return nil, err
				}
				logger.Error("failed to call debug_traceTransaction", err, "revert_reason", revertReason)
			}
			return nil, err
		}
		txGasLimit := estimatedGas * c.Config().GasEstimateRate.Numerator / c.Config().GasEstimateRate.Denominator
		if txGasLimit > c.Config().MaxGasLimit {
			logger.Warn("estimated gas exceeds max gas limit", "estimated_gas", txGasLimit, "max_gas_limit", c.Config().MaxGasLimit)
			txGasLimit = c.Config().MaxGasLimit
		}
		opts.GasLimit = txGasLimit
		opts.NoSend = false
		tx, err = c.SendTx(opts, msg, skipUpdateClientCommitment)
		if err != nil {
			logger.Error("failed to send msg / NoSend: false", err)
			return nil, err
		}
		receipt, err := c.client.WaitForReceiptAndGet(ctx, tx.Hash())
		if err != nil {
			logger.Error("failed to get receipt", err)
			return nil, err
		}
		if receipt.Status == gethtypes.ReceiptStatusFailed {
			var revertReason string
			if receipt.HasRevertReason() {
				revertReason = c.GetRevertReason(receipt)
			} else if c.config.EnableDebugTrace {
				revertReason, err = c.DebugTraceTransaction(ctx, tx.Hash())
				if err != nil {
					logger.Error("failed to call debug_traceTransaction", err)
					return nil, err
				}
			}
			err = fmt.Errorf("revert_reason=%s", revertReason)
			logger.Error("tx execution failed", err, "msg_index", i)
			return nil, err
		}
		if c.msgEventListener != nil {
			if err := c.msgEventListener.OnSentMsg([]sdk.Msg{msg}); err != nil {
				logger.Error("failed to OnSendMsg call", err)
			}
		}
		msgIDs = append(msgIDs, NewMsgID(tx.Hash()))
	}
	return msgIDs, nil
}

func (c *Chain) GetMsgResult(id core.MsgID) (core.MsgResult, error) {
	msgID, ok := id.(*MsgID)
	if !ok {
		return nil, fmt.Errorf("unexpected message id type: %T", id)
	}
	ctx := context.TODO()
	txHash := msgID.TxHash()
	receipt, err := c.client.WaitForReceiptAndGet(ctx, txHash)
	if err != nil {
		return nil, err
	}
	if receipt.Status == gethtypes.ReceiptStatusSuccessful {
		return c.makeMsgResultFromReceipt(&receipt.Receipt, "")
	}
	var revertReason string
	if receipt.HasRevertReason() {
		revertReason = c.GetRevertReason(receipt)
	} else if c.config.EnableDebugTrace {
		revertReason, err = c.DebugTraceTransaction(ctx, txHash)
		if err != nil {
			return nil, err
		}
	}
	return c.makeMsgResultFromReceipt(&receipt.Receipt, revertReason)
}

func (c *Chain) TxCreateClient(opts *bind.TransactOpts, msg *clienttypes.MsgCreateClient) (*gethtypes.Transaction, error) {
	var clientState exported.ClientState
	if err := c.codec.UnpackAny(msg.ClientState, &clientState); err != nil {
		return nil, err
	}
	clientStateBytes, err := proto.Marshal(msg.ClientState)
	if err != nil {
		return nil, err
	}
	consensusStateBytes, err := proto.Marshal(msg.ConsensusState)
	if err != nil {
		return nil, err
	}
	return c.ibcHandler.CreateClient(opts, ibchandler.IIBCClientMsgCreateClient{
		ClientType:          clientState.ClientType(),
		ProtoClientState:    clientStateBytes,
		ProtoConsensusState: consensusStateBytes,
	})
}

func (c *Chain) TxUpdateClient(opts *bind.TransactOpts, msg *clienttypes.MsgUpdateClient, skipUpdateClientCommitment bool) (*gethtypes.Transaction, error) {
	clientMessageBytes, err := proto.Marshal(msg.ClientMessage)
	if err != nil {
		return nil, err
	}
	m := ibchandler.IIBCClientMsgUpdateClient{
		ClientId:           msg.ClientId,
		ProtoClientMessage: clientMessageBytes,
	}
	// if `skipUpdateClientCommitment` is true and `allowLCFunctions` is not nil,
	// the relayer calls `routeUpdateClient` to constructs an UpdateClient tx to the LC contract.
	// ref. https://github.com/hyperledger-labs/yui-ibc-solidity/blob/main/docs/adr/adr-001.md
	if skipUpdateClientCommitment && c.allowLCFunctions != nil {
		lcAddr, fnSel, args, err := c.ibcHandler.RouteUpdateClient(c.CallOpts(opts.Context, 0), m)
		if err != nil {
			return nil, fmt.Errorf("failed to route update client: %w", err)
		}
		// ensure that the contract and function are allowed
		if c.allowLCFunctions.IsAllowed(lcAddr, fnSel) {
			log.GetLogger().Info("contract function is allowed", "contract", lcAddr.Hex(), "selector", fmt.Sprintf("0x%x", fnSel))
			calldata := append(fnSel[:], args...)
			return bind.NewBoundContract(lcAddr, abi.ABI{}, c.client, c.client, c.client).RawTransact(opts, calldata)
		}
		// fallback to send an UpdateClient to the IBC handler contract
		log.GetLogger().Warn("contract function is not allowed", "contract", lcAddr.Hex(), "selector", fmt.Sprintf("0x%x", fnSel))
	}
	return c.ibcHandler.UpdateClient(opts, m)
}

func (c *Chain) TxConnectionOpenInit(opts *bind.TransactOpts, msg *conntypes.MsgConnectionOpenInit) (*gethtypes.Transaction, error) {
	return c.ibcHandler.ConnectionOpenInit(opts, ibchandler.IIBCConnectionMsgConnectionOpenInit{
		ClientId: msg.ClientId,
		Counterparty: ibchandler.CounterpartyData{
			ClientId:     msg.Counterparty.ClientId,
			ConnectionId: msg.Counterparty.ConnectionId,
			Prefix:       ibchandler.MerklePrefixData(msg.Counterparty.Prefix),
		},
		DelayPeriod: msg.DelayPeriod,
	})
}

func (c *Chain) TxConnectionOpenTry(opts *bind.TransactOpts, msg *conntypes.MsgConnectionOpenTry) (*gethtypes.Transaction, error) {
	clientStateBytes, err := proto.Marshal(msg.ClientState)
	if err != nil {
		return nil, err
	}
	var versions []ibchandler.VersionData
	for _, v := range msg.CounterpartyVersions {
		versions = append(versions, ibchandler.VersionData(*v))
	}
	return c.ibcHandler.ConnectionOpenTry(opts, ibchandler.IIBCConnectionMsgConnectionOpenTry{
		Counterparty: ibchandler.CounterpartyData{
			ClientId:     msg.Counterparty.ClientId,
			ConnectionId: msg.Counterparty.ConnectionId,
			Prefix:       ibchandler.MerklePrefixData(msg.Counterparty.Prefix),
		},
		DelayPeriod:             msg.DelayPeriod,
		ClientId:                msg.ClientId,
		ClientStateBytes:        clientStateBytes,
		CounterpartyVersions:    versions,
		ProofInit:               msg.ProofInit,
		ProofClient:             msg.ProofClient,
		ProofConsensus:          msg.ProofConsensus,
		ProofHeight:             pbToHandlerHeight(msg.ProofHeight),
		ConsensusHeight:         pbToHandlerHeight(msg.ConsensusHeight),
		HostConsensusStateProof: msg.HostConsensusStateProof,
	})
}

func (c *Chain) TxConnectionOpenAck(opts *bind.TransactOpts, msg *conntypes.MsgConnectionOpenAck) (*gethtypes.Transaction, error) {
	clientStateBytes, err := proto.Marshal(msg.ClientState)
	if err != nil {
		return nil, err
	}
	return c.ibcHandler.ConnectionOpenAck(opts, ibchandler.IIBCConnectionMsgConnectionOpenAck{
		ConnectionId:     msg.ConnectionId,
		ClientStateBytes: clientStateBytes,
		Version: ibchandler.VersionData{
			Identifier: msg.Version.Identifier,
			Features:   msg.Version.Features,
		},
		CounterpartyConnectionId: msg.CounterpartyConnectionId,
		ProofTry:                 msg.ProofTry,
		ProofClient:              msg.ProofClient,
		ProofConsensus:           msg.ProofConsensus,
		ProofHeight:              pbToHandlerHeight(msg.ProofHeight),
		ConsensusHeight:          pbToHandlerHeight(msg.ConsensusHeight),
		HostConsensusStateProof:  msg.HostConsensusStateProof,
	})
}

func (c *Chain) TxConnectionOpenConfirm(opts *bind.TransactOpts, msg *conntypes.MsgConnectionOpenConfirm) (*gethtypes.Transaction, error) {
	return c.ibcHandler.ConnectionOpenConfirm(opts, ibchandler.IIBCConnectionMsgConnectionOpenConfirm{
		ConnectionId: msg.ConnectionId,
		ProofAck:     msg.ProofAck,
		ProofHeight:  pbToHandlerHeight(msg.ProofHeight),
	})
}

func (c *Chain) TxChannelOpenInit(opts *bind.TransactOpts, msg *chantypes.MsgChannelOpenInit) (*gethtypes.Transaction, error) {
	return c.ibcHandler.ChannelOpenInit(opts, ibchandler.IIBCChannelHandshakeMsgChannelOpenInit{
		PortId: msg.PortId,
		Channel: ibchandler.ChannelData{
			State:          uint8(msg.Channel.State),
			Ordering:       uint8(msg.Channel.Ordering),
			Counterparty:   ibchandler.ChannelCounterpartyData(msg.Channel.Counterparty),
			ConnectionHops: msg.Channel.ConnectionHops,
			Version:        msg.Channel.Version,
		},
	})
}

func (c *Chain) TxChannelOpenTry(opts *bind.TransactOpts, msg *chantypes.MsgChannelOpenTry) (*gethtypes.Transaction, error) {
	return c.ibcHandler.ChannelOpenTry(opts, ibchandler.IIBCChannelHandshakeMsgChannelOpenTry{
		PortId: msg.PortId,
		Channel: ibchandler.ChannelData{
			State:          uint8(msg.Channel.State),
			Ordering:       uint8(msg.Channel.Ordering),
			Counterparty:   ibchandler.ChannelCounterpartyData(msg.Channel.Counterparty),
			ConnectionHops: msg.Channel.ConnectionHops,
			Version:        msg.Channel.Version,
		},
		CounterpartyVersion: msg.CounterpartyVersion,
		ProofInit:           msg.ProofInit,
		ProofHeight:         pbToHandlerHeight(msg.ProofHeight),
	})
}

func (c *Chain) TxChannelOpenAck(opts *bind.TransactOpts, msg *chantypes.MsgChannelOpenAck) (*gethtypes.Transaction, error) {
	return c.ibcHandler.ChannelOpenAck(opts, ibchandler.IIBCChannelHandshakeMsgChannelOpenAck{
		PortId:                msg.PortId,
		ChannelId:             msg.ChannelId,
		CounterpartyVersion:   msg.CounterpartyVersion,
		CounterpartyChannelId: msg.CounterpartyChannelId,
		ProofTry:              msg.ProofTry,
		ProofHeight:           pbToHandlerHeight(msg.ProofHeight),
	})
}

func (c *Chain) TxChannelOpenConfirm(opts *bind.TransactOpts, msg *chantypes.MsgChannelOpenConfirm) (*gethtypes.Transaction, error) {
	return c.ibcHandler.ChannelOpenConfirm(opts, ibchandler.IIBCChannelHandshakeMsgChannelOpenConfirm{
		PortId:      msg.PortId,
		ChannelId:   msg.ChannelId,
		ProofAck:    msg.ProofAck,
		ProofHeight: pbToHandlerHeight(msg.ProofHeight),
	})
}

func (c *Chain) TxRecvPacket(opts *bind.TransactOpts, msg *chantypes.MsgRecvPacket) (*gethtypes.Transaction, error) {
	return c.ibcHandler.RecvPacket(opts, ibchandler.IIBCChannelRecvPacketMsgPacketRecv{
		Packet: ibchandler.PacketData{
			Sequence:           msg.Packet.Sequence,
			SourcePort:         msg.Packet.SourcePort,
			SourceChannel:      msg.Packet.SourceChannel,
			DestinationPort:    msg.Packet.DestinationPort,
			DestinationChannel: msg.Packet.DestinationChannel,
			Data:               msg.Packet.Data,
			TimeoutHeight:      ibchandler.HeightData(msg.Packet.TimeoutHeight),
			TimeoutTimestamp:   msg.Packet.TimeoutTimestamp,
		},
		Proof:       msg.ProofCommitment,
		ProofHeight: pbToHandlerHeight(msg.ProofHeight),
	})
}

func (c *Chain) TxAcknowledgement(opts *bind.TransactOpts, msg *chantypes.MsgAcknowledgement) (*gethtypes.Transaction, error) {
	return c.ibcHandler.AcknowledgePacket(opts, ibchandler.IIBCChannelAcknowledgePacketMsgPacketAcknowledgement{
		Packet: ibchandler.PacketData{
			Sequence:           msg.Packet.Sequence,
			SourcePort:         msg.Packet.SourcePort,
			SourceChannel:      msg.Packet.SourceChannel,
			DestinationPort:    msg.Packet.DestinationPort,
			DestinationChannel: msg.Packet.DestinationChannel,
			Data:               msg.Packet.Data,
			TimeoutHeight:      ibchandler.HeightData(msg.Packet.TimeoutHeight),
			TimeoutTimestamp:   msg.Packet.TimeoutTimestamp,
		},
		Acknowledgement: msg.Acknowledgement,
		Proof:           msg.ProofAcked,
		ProofHeight:     pbToHandlerHeight(msg.ProofHeight),
	})
}

func (c *Chain) SendTx(opts *bind.TransactOpts, msg sdk.Msg, skipUpdateClientCommitment bool) (*gethtypes.Transaction, error) {
	logger := c.GetChainLogger()
	var (
		tx  *gethtypes.Transaction
		err error
	)
	switch msg := msg.(type) {
	case *clienttypes.MsgCreateClient:
		tx, err = c.TxCreateClient(opts, msg)
	case *clienttypes.MsgUpdateClient:
		tx, err = c.TxUpdateClient(opts, msg, skipUpdateClientCommitment)
	case *conntypes.MsgConnectionOpenInit:
		tx, err = c.TxConnectionOpenInit(opts, msg)
	case *conntypes.MsgConnectionOpenTry:
		tx, err = c.TxConnectionOpenTry(opts, msg)
	case *conntypes.MsgConnectionOpenAck:
		tx, err = c.TxConnectionOpenAck(opts, msg)
	case *conntypes.MsgConnectionOpenConfirm:
		tx, err = c.TxConnectionOpenConfirm(opts, msg)
	case *chantypes.MsgChannelOpenInit:
		tx, err = c.TxChannelOpenInit(opts, msg)
	case *chantypes.MsgChannelOpenTry:
		tx, err = c.TxChannelOpenTry(opts, msg)
	case *chantypes.MsgChannelOpenAck:
		tx, err = c.TxChannelOpenAck(opts, msg)
	case *chantypes.MsgChannelOpenConfirm:
		tx, err = c.TxChannelOpenConfirm(opts, msg)
	case *chantypes.MsgRecvPacket:
		tx, err = c.TxRecvPacket(opts, msg)
	case *chantypes.MsgAcknowledgement:
		tx, err = c.TxAcknowledgement(opts, msg)
	// case *transfertypes.MsgTransfer:
	// 	err = c.client.transfer(msg)
	default:
		logger.Error("failed to send msg", errors.New("illegal msg type"), "msg", msg)
		panic("illegal msg type")
	}
	return tx, err
}

func (c *Chain) GetRevertReason(receipt *client.Receipt) string {
	erepo := NewErrorsRepository(c.config.CustomErrors)
	sig, args, err := erepo.ParseError(receipt.RevertReason)
	if err != nil {
		return fmt.Sprintf("raw-revert-reason=\"%x\" parse-err=\"%v\"", receipt.RevertReason, err)
	}
	return fmt.Sprintf("revert-reason=\"%v\" args=\"%v\"", sig, args)
}

func (c *Chain) DebugTraceTransaction(ctx context.Context, txHash common.Hash) (string, error) {
	var result *callFrame
	if err := c.client.Raw().CallContext(ctx, &result, "debug_traceTransaction", txHash, map[string]string{"tracer": "callTracer"}); err != nil {
		return "", err
	}
	revertReason, err := searchRevertReason(result)
	if err != nil {
		return "", err
	}
	return revertReason, nil
}

type callLog struct {
	Address common.Address `json:"address"`
	Topics  []common.Hash  `json:"topics"`
	Data    hexutil.Bytes  `json:"data"`
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

type ErrorsRepository struct {
	errs map[[4]byte]abi.Error
}

func NewErrorsRepository(customErrors []*CustomError) ErrorsRepository {
	customErrors = append(customErrors, defaultErrors()...)
	er := ErrorsRepository{
		errs: make(map[[4]byte]abi.Error),
	}
	for _, e := range customErrors {
		var arguments []abi.Argument
		for _, arg := range e.Arguments {
			strT, err := abi.NewType(arg.Type, "", nil)
			if err != nil {
				panic(err)
			}
			arguments = append(arguments, abi.Argument{Type: strT})

		}
		e := abi.NewError(e.FunctionName, arguments)
		if err := er.Add(e); err != nil {
			panic(err)
		}
	}
	return er
}

func (r *ErrorsRepository) Add(e abi.Error) error {
	var sel [4]byte
	copy(sel[:], e.ID[:4])
	if _, ok := r.errs[sel]; ok {
		return fmt.Errorf("duplicate error selector: error=%v sel=%x", e.String(), sel)
	}
	r.errs[sel] = e
	return nil
}

func (r *ErrorsRepository) GetError(sel [4]byte) (abi.Error, bool) {
	e, ok := r.errs[sel]
	return e, ok
}

func (r *ErrorsRepository) ParseError(bz []byte) (string, interface{}, error) {
	if len(bz) < 4 {
		return "", nil, fmt.Errorf("invalid error data: %v", bz)
	}
	var sel [4]byte
	copy(sel[:], bz[:4])
	e, ok := r.GetError(sel)
	if !ok {
		return "", nil, fmt.Errorf("unknown error: sel=%x", sel)
	}
	v, err := e.Unpack(bz)
	return e.Sig, v, err
}

func defaultErrors() []*CustomError {
	customErrors := []*CustomError{
		{
			FunctionName: "Error",
			Arguments: []*Argument{
				{
					Type: "string",
				},
			},
		},
		{
			FunctionName: "Panic",
			Arguments: []*Argument{
				{
					Type: "uint256",
				},
			},
		},
	}
	return customErrors
}
