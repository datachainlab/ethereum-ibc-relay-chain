package ethereum

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	math "math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	conntypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	chantypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/hyperledger-labs/yui-relayer/core"
	"github.com/hyperledger-labs/yui-relayer/log"

	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/client"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/contract/ibchandler"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/contract/multicall3"
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

	iter := NewMulticallIter(msgs, skipUpdateClientCommitment)
	for !iter.IsEnd() {
		from := iter.Cursor()
		tx, err := iter.SendTx(ctx, c)
		if err != nil {
			logger.Error("failed to send msg", err)
			return nil, err
		}
		if tx == nil {
			break
		}

		logger.Logger = logger.With(logAttrTxHash, tx.Hash())
		if rawTxData, err := tx.MarshalBinary(); err != nil {
			logger.Error("failed to encode tx", err)
		} else {
			logger.Logger = logger.With(logAttrRawTxData, hex.EncodeToString(rawTxData))
		}

		receipt, err := c.client.WaitForReceiptAndGet(ctx, tx.Hash())
		if err != nil {
			logger.Error("failed to get receipt", err, "msg_index.from", from, "msg_index.to", iter.Cursor(), "tx_hash", tx.Hash())
			return nil, err
		}
		if receipt.Status == gethtypes.ReceiptStatusFailed {
			revertReason, rawErrorData, err2 := c.getRevertReasonFromReceipt(ctx, receipt)
			if err2 != nil {
				logger.Error("failed to get revert reason", err2, "msg_index.from", from, "msg_index.to", iter.Cursor(), "tx_hash", tx.Hash())
			}

			err := fmt.Errorf("tx execution reverted: revertReason=%s, rawErrorData=%x, msgIndex.from=%d, msgIndex.to=%d, txHash=%s", revertReason, rawErrorData, from, iter.Cursor(), tx.Hash())
			logger.Error("tx execution reverted", err, "revert_reason", revertReason, "raw_error_data", hex.EncodeToString(rawErrorData), "msg_index.from", from, "msg_index.to", iter.Cursor(), "tx_hash", tx.Hash())
			return nil, err
		}
		if c.msgEventListener != nil {
			for i := from; i < iter.Cursor(); i++ {
				if err := c.msgEventListener.OnSentMsg([]sdk.Msg{msgs[i]}); err != nil {
					logger.Error("failed to OnSendMsg call", err, "msg_index", i, "tx_hash", tx.Hash())
				}
			}
		}
		for i := from; i < iter.Cursor(); i++ {
			msgIDs = append(msgIDs, NewMsgID(tx.Hash()))
		}
	}
	return msgIDs, nil
}

func (c *Chain) GetMsgResult(id core.MsgID) (core.MsgResult, error) {
	logger := c.GetChainLogger()

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
	revertReason, rawErrorData, err := c.getRevertReasonFromReceipt(ctx, receipt)
	if err != nil {
		logger.Error("failed to get revert reason", err, "raw_error_data", hex.EncodeToString(rawErrorData), "tx_hash", msgID.TxHashHex)
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
		Packet: ibchandler.Packet{
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
		Packet: ibchandler.Packet{
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

type MulticallIter struct {
	msgs []sdk.Msg
	cursor int
	opts  *bind.TransactOpts
	skipUpdateClientCommitment bool
}
func NewMulticallIter(msgs []sdk.Msg, skipUpdateClientCommitment bool) MulticallIter {
	return MulticallIter {
		msgs: msgs,
		cursor: 0,
		skipUpdateClientCommitment: skipUpdateClientCommitment,
	}
}
func (m *MulticallIter) Cursor() int {
	return m.cursor
}
func (m *MulticallIter) Current() *sdk.Msg {
	return &m.msgs[m.cursor]
}
func (m *MulticallIter) IsEnd() bool {
	return len(m.msgs) <= m.cursor
}
func (m *MulticallIter) Next() bool {
	if m.IsEnd() {
		return false
	}
	m.cursor += 1
	return true
}

func (m *MulticallIter) SendTx(ctx context.Context, c *Chain) (*gethtypes.Transaction, error) {
	if c.multicall3 == nil {
		return m.sendSinglecallTx(ctx, c)
	} else {
		return m.sendMultilcallTx(ctx, c)
	}
}

func (m *MulticallIter) sendMulticallTx(ctx context.Context, c *Chain) (*gethtypes.Transaction, error) {
	from := m.Cursor()
	if (m.IsEnd()) {
		return nil, nil
	}

	logger := c.GetChainLogger()

	var (
		saveTxGasLimit uint64
		opts *bind.TransactOpts
	)

	calls := make([]multicall3.Multicall3Call, 0, len(m.msgs))
	logger.Info("SentMulticallTx", "msgs", len(m.msgs), "len", len(calls), "cap", cap(calls))

	opts, err := c.TxOpts(ctx);
	if err != nil {
		if err != nil {
			return nil, err
		}
	}

	for !m.IsEnd() {
		logger := &log.RelayLogger{Logger: logger.With(
			logAttrMsgIndex, m.Cursor(),
			logAttrMsgType, fmt.Sprintf("%T", m.Current()),
		)}

		opts.GasLimit = math.MaxUint64
		opts.NoSend = true
		tx, err := c.SendTx(opts, m.Current(), m.skipUpdateClientCommitment)
		if err != nil {
			if len(calls) > 0 {
				break
			}
			return nil, err
		}

		to := tx.To(); if to == nil {
			if len(calls) > 0 {
				break
			}
			err2 := fmt.Errorf("no target address")
			logger.Error("failed to construct Multicall3Call", err2, "msg_index", m.Cursor())
			return nil, err2
		}
		newCalls := append(calls, multicall3.Multicall3Call{
			Target: *tx.To(),
			CallData: tx.Data(),
		})
		multiTx, err := c.multicall3.Aggregate(opts, newCalls)
		if err != nil {
			if len(calls) > 0 {
				break
			}
			logger.Error("failed to call Multicall3.Aggregate", err, "msg_index", m.Cursor())
			return nil, err
		}

		estimatedGas, err := c.client.EstimateGasFromTx(ctx, multiTx)
		if err != nil {
			if len(calls) > 0 {
				break
			}
			revertReason, rawErrorData, err2 := c.getRevertReasonFromEstimateGas(err)
			if err2 != nil {
				logger.Error("failed to get revert reason", err2, "msg_index", m.Cursor())
			}

			logger.Error("failed to estimate gas", err, "revert_reason", revertReason, "raw_error_data", hex.EncodeToString(rawErrorData), "msg_index", m.Cursor())
			return nil, err
		}

		txGasLimit := estimatedGas * c.Config().GasEstimateRate.Numerator / c.Config().GasEstimateRate.Denominator
		if txGasLimit > c.Config().MaxGasLimit {
			if len(calls) > 0 {
				break
			}
			logger.Warn("estimated gas exceeds max gas limit", "estimated_gas", txGasLimit, "max_gas_limit", c.Config().MaxGasLimit, "msg_index", m.Cursor())
			txGasLimit = c.Config().MaxGasLimit
			// pass
		}

		saveTxGasLimit = txGasLimit
		calls = newCalls
		m.Next()
		logger.Info("next", "cursor", m.Cursor(), "len", len(calls), "cap", cap(calls))
	}

	if len(calls) == 0 {
		return nil, nil
	}

	opts.GasLimit = saveTxGasLimit
	opts.NoSend = false
	logger.Info("multicall", "count", len(calls))
	tx, err := c.multicall3.Aggregate(opts, calls)
	if err != nil {
		logger.Error("failed to send msg / NoSend: true", err, "msg_index.from", from, "msg_index.to", m.Cursor())
		return nil, err
	}
	return tx, nil
}

func (m *MulticallIter) sendMulticallTx(ctx context.Context, c *Chain) (*gethtypes.Transaction, int, error) {
	if m.IsEnd() {
		return nil, m.Cursor(), nil
	}

	logger := &log.RelayLogger{Logger: logger.With(
		logAttrMsgIndex, m.Cursor(),
		logAttrMsgType, fmt.Sprintf("%T", m.Current()),
	)}

	// gas estimation
	{
		logger := &log.RelayLogger{Logger: logger.Logger}

		opts.GasLimit = math.MaxUint64
		opts.NoSend = true
		tx, err := c.SendTx(opts, m.Current(), skipUpdateClientCommitment)
		if err != nil {
			logger.Error("failed to build tx for gas estimation", err)
			return nil, err
		}

		if rawTxData, err := tx.MarshalBinary(); err != nil {
			logger.Error("failed to encode tx", err)
		} else {
			logger.Logger = logger.With(logAttrRawTxData, hex.EncodeToString(rawTxData))
		}

		estimatedGas, err := c.client.EstimateGasFromTx(ctx, tx)
		if err != nil {
			if revertReason, rawErrorData, err := c.getRevertReasonFromEstimateGas(err); err != nil {
				// Raw error data may be available even if revert reason isn't available.
				logger.Logger = logger.With(logAttrRawErrorData, hex.EncodeToString(rawErrorData))
				logger.Error("failed to get revert reason", err)
			} else {
				logger.Logger = logger.With(
					logAttrRawErrorData, hex.EncodeToString(rawErrorData),
					logAttrRevertReason, revertReason,
				)
			}

			logger.Error("failed to estimate gas", err)
			return nil, err
		}

		txGasLimit := estimatedGas * c.Config().GasEstimateRate.Numerator / c.Config().GasEstimateRate.Denominator
		if txGasLimit > c.Config().MaxGasLimit {
			logger.Warn("estimated gas exceeds max gas limit",
				logAttrEstimatedGas, txGasLimit,
				logAttrMaxGasLimit, c.Config().MaxGasLimit,
			)
			txGasLimit = c.Config().MaxGasLimit
		}
		opts.GasLimit = txGasLimit
	}

	opts.NoSend = false
	tx, err := c.SendTx(opts, m.Current(), skipUpdateClientCommitment)
	if err != nil {
		logger.Error("failed to send msg", err)
		return nil, m.Cursor(), err
	}
	m.Next()
	return tx, nil
}


func (c *Chain) getRevertReasonFromReceipt(ctx context.Context, receipt *client.Receipt) (string, []byte, error) {
	var errorData []byte
	if receipt.HasRevertReason() {
		errorData = receipt.RevertReason
	} else if c.config.EnableDebugTrace {
		callFrame, err := c.client.DebugTraceTransaction(ctx, receipt.TxHash)
		if err != nil {
			return "", nil, err
		} else if len(callFrame.Output) == 0 {
			return "", nil, fmt.Errorf("execution reverted without error data")
		}
		errorData = callFrame.Output
	} else {
		return "", nil, fmt.Errorf("no way to get revert reason")
	}

	revertReason, err := c.errorRepository.ParseError(errorData)
	if err != nil {
		return "", errorData, fmt.Errorf("failed to parse error: %v", err)
	}
	return revertReason, errorData, nil
}

func (c *Chain) getRevertReasonFromEstimateGas(err error) (string, []byte, error) {
	if de, ok := err.(rpc.DataError); !ok {
		return "", nil, fmt.Errorf("eth_estimateGas failed with unexpected error type: errorType=%T", err)
	} else if de.ErrorData() == nil {
		return "", nil, fmt.Errorf("eth_estimateGas failed without error data")
	} else if errorData, ok := de.ErrorData().(string); !ok {
		return "", nil, fmt.Errorf("eth_estimateGas failed with unexpected error data type: errorDataType=%T", de.ErrorData())
	} else {
		errorData := common.FromHex(errorData)
		revertReason, err := c.errorRepository.ParseError(errorData)
		if err != nil {
			return "", errorData, fmt.Errorf("failed to parse error: %v", err)
		}
		return revertReason, errorData, nil
	}
}
