package ethereum

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	math "math"
	"slices"
	"sort"
	"strings"

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
	ethereumSignerLogger := c.ethereumSigner.GetLogger()
	defer c.ethereumSigner.SetLogger(ethereumSignerLogger)

	var msgIDs []core.MsgID

	iter := NewCallIter(msgs, skipUpdateClientCommitment)
	for !iter.End() {
		from := iter.Cursor()
		logger := &log.RelayLogger{Logger: logger.With(logAttrMsgIndexFrom, from)}
		c.ethereumSigner.SetLogger(logger)

		built, err := iter.BuildTx(ctx, c)

		if err != nil {
			logger.Error("failed to build msg tx", err)
			return nil, err
		} else if built == nil {
			break
		} else {
			logger = iter.updateLoggerMessageInfo(logger, from, built.count)
			logger = &log.RelayLogger{Logger: logger.With(logAttrTxHash, built.tx.Hash())}
		}

		if rawTxData, err := built.tx.MarshalBinary(); err != nil {
			logger.Error("failed to encode tx", err)
		} else {
			logger = &log.RelayLogger{Logger: logger.With(logAttrRawTxData, hex.EncodeToString(rawTxData))}
		}

		err = c.client.SendTransaction(ctx, built.tx)
		if err != nil {
			logger.Error("failed to send tx", err)
			return nil, err
		}

		receipt, err := c.client.WaitForReceiptAndGet(ctx, built.tx.Hash())
		if err != nil {
			logger.Error("failed to get receipt", err)
			return nil, err
		} else {
			logger = &log.RelayLogger{Logger: logger.With(
				logAttrBlockHash, receipt.BlockHash,
				logAttrBlockNumber, receipt.BlockNumber.Uint64(),
				logAttrTxIndex, receipt.TransactionIndex,
			)}
		}

		if receipt.Status == gethtypes.ReceiptStatusFailed {
			if revertReason, rawErrorData, err := c.getRevertReasonFromReceipt(ctx, receipt); err != nil {
				// Raw error data may be available even if revert reason isn't available.
				logger = &log.RelayLogger{Logger: logger.With(
					logAttrRawErrorData, hex.EncodeToString(rawErrorData),
				)}
				logger.Error("failed to get revert reason", err)
			} else {
				logger = &log.RelayLogger{Logger: logger.With(
					logAttrRawErrorData, hex.EncodeToString(rawErrorData),
					logAttrRevertReason, revertReason,
				)}
			}

			err := errors.New("tx execution reverted")
			logger.Error("tx execution reverted", err)
			return nil, err
		}
		logger.Info("successfully sent tx")
		if c.msgEventListener != nil {
			for i := from; i < iter.Cursor(); i++ {
				if err := c.msgEventListener.OnSentMsg([]sdk.Msg{msgs[i]}); err != nil {
					logger.Error("failed to OnSendMsg call", err, "index", i)
				}
			}
		}
		for i := 0; i < built.count; i++ {
			msgIDs = append(msgIDs, NewMsgID(built.tx.Hash()))
		}
		iter.Next(built.count)
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
		logger.Error("failed to get revert reason", err,
			logAttrRawErrorData, hex.EncodeToString(rawErrorData),
			logAttrTxHash, msgID.TxHashHex,
			logAttrBlockHash, receipt.BlockHash.Hex(),
			logAttrBlockNumber, receipt.BlockNumber.Uint64(),
			logAttrTxIndex, receipt.TransactionIndex,
		)
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

func (c *Chain) TxChannelUpgradeInit(opts *bind.TransactOpts, msg *chantypes.MsgChannelUpgradeInit) (*gethtypes.Transaction, error) {
	return c.ibcHandler.ChannelUpgradeInit(opts, ibchandler.IIBCChannelUpgradeBaseMsgChannelUpgradeInit{
		PortId:                c.pathEnd.PortID,
		ChannelId:             c.pathEnd.ChannelID,
		ProposedUpgradeFields: pbToUpgradeFields(msg.Fields),
	})
}

func (c *Chain) TxChannelUpgradeTry(opts *bind.TransactOpts, msg *chantypes.MsgChannelUpgradeTry) (*gethtypes.Transaction, error) {
	return c.ibcHandler.ChannelUpgradeTry(opts, ibchandler.IIBCChannelUpgradeBaseMsgChannelUpgradeTry{
		PortId:                      c.pathEnd.PortID,
		ChannelId:                   c.pathEnd.ChannelID,
		CounterpartyUpgradeSequence: msg.CounterpartyUpgradeSequence,
		CounterpartyUpgradeFields:   pbToUpgradeFields(msg.CounterpartyUpgradeFields),
		ProposedConnectionHops:      slices.Clone(msg.ProposedUpgradeConnectionHops),
		Proofs: ibchandler.IIBCChannelUpgradeBaseChannelUpgradeProofs{
			ProofChannel: msg.ProofChannel,
			ProofUpgrade: msg.ProofUpgrade,
			ProofHeight:  pbToHandlerHeight(msg.ProofHeight),
		},
	})
}

func (c *Chain) TxChannelUpgradeAck(opts *bind.TransactOpts, msg *chantypes.MsgChannelUpgradeAck) (*gethtypes.Transaction, error) {
	return c.ibcHandler.ChannelUpgradeAck(opts, ibchandler.IIBCChannelUpgradeBaseMsgChannelUpgradeAck{
		PortId:              c.pathEnd.PortID,
		ChannelId:           c.pathEnd.ChannelID,
		CounterpartyUpgrade: pbToUpgrade(msg.CounterpartyUpgrade),
		Proofs: ibchandler.IIBCChannelUpgradeBaseChannelUpgradeProofs{
			ProofChannel: msg.ProofChannel,
			ProofUpgrade: msg.ProofUpgrade,
			ProofHeight:  pbToHandlerHeight(msg.ProofHeight),
		},
	})
}

func (c *Chain) TxChannelUpgradeConfirm(opts *bind.TransactOpts, msg *chantypes.MsgChannelUpgradeConfirm) (*gethtypes.Transaction, error) {
	return c.ibcHandler.ChannelUpgradeConfirm(opts, ibchandler.IIBCChannelUpgradeBaseMsgChannelUpgradeConfirm{
		PortId:                   c.pathEnd.PortID,
		ChannelId:                c.pathEnd.ChannelID,
		CounterpartyChannelState: uint8(msg.CounterpartyChannelState),
		CounterpartyUpgrade:      pbToUpgrade(msg.CounterpartyUpgrade),
		Proofs: ibchandler.IIBCChannelUpgradeBaseChannelUpgradeProofs{
			ProofChannel: msg.ProofChannel,
			ProofUpgrade: msg.ProofUpgrade,
			ProofHeight:  pbToHandlerHeight(msg.ProofHeight),
		},
	})
}

func (c *Chain) TxChannelUpgradeOpen(opts *bind.TransactOpts, msg *chantypes.MsgChannelUpgradeOpen) (*gethtypes.Transaction, error) {
	return c.ibcHandler.ChannelUpgradeOpen(opts, ibchandler.IIBCChannelUpgradeBaseMsgChannelUpgradeOpen{
		PortId:                      c.pathEnd.PortID,
		ChannelId:                   c.pathEnd.ChannelID,
		CounterpartyChannelState:    uint8(msg.CounterpartyChannelState),
		CounterpartyUpgradeSequence: msg.CounterpartyUpgradeSequence,
		ProofChannel:                msg.ProofChannel,
		ProofHeight:                 pbToHandlerHeight(msg.ProofHeight),
	})
}

func (c *Chain) TxChannelUpgradeCancel(opts *bind.TransactOpts, msg *chantypes.MsgChannelUpgradeCancel) (*gethtypes.Transaction, error) {
	return c.ibcHandler.CancelChannelUpgrade(opts, ibchandler.IIBCChannelUpgradeBaseMsgCancelChannelUpgrade{
		PortId:            c.pathEnd.PortID,
		ChannelId:         c.pathEnd.ChannelID,
		ErrorReceipt:      ibchandler.ErrorReceiptData(msg.ErrorReceipt),
		ProofUpgradeError: msg.ProofErrorReceipt,
		ProofHeight:       pbToHandlerHeight(msg.ProofHeight),
	})
}

func (c *Chain) TxChannelUpgradeTimeout(opts *bind.TransactOpts, msg *chantypes.MsgChannelUpgradeTimeout) (*gethtypes.Transaction, error) {
	return c.ibcHandler.TimeoutChannelUpgrade(opts, ibchandler.IIBCChannelUpgradeBaseMsgTimeoutChannelUpgrade{
		PortId:              c.pathEnd.PortID,
		ChannelId:           c.pathEnd.ChannelID,
		CounterpartyChannel: pbToChannel(msg.CounterpartyChannel),
		ProofChannel:        msg.ProofChannel,
		ProofHeight:         pbToHandlerHeight(msg.ProofHeight),
	})
}

func (c *Chain) BuildMessageTx(opts *bind.TransactOpts, msg sdk.Msg, skipUpdateClientCommitment bool) (*gethtypes.Transaction, error) {
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
	case *chantypes.MsgChannelUpgradeInit:
		tx, err = c.TxChannelUpgradeInit(opts, msg)
	case *chantypes.MsgChannelUpgradeTry:
		tx, err = c.TxChannelUpgradeTry(opts, msg)
	case *chantypes.MsgChannelUpgradeAck:
		tx, err = c.TxChannelUpgradeAck(opts, msg)
	case *chantypes.MsgChannelUpgradeConfirm:
		tx, err = c.TxChannelUpgradeConfirm(opts, msg)
	case *chantypes.MsgChannelUpgradeOpen:
		tx, err = c.TxChannelUpgradeOpen(opts, msg)
	case *chantypes.MsgChannelUpgradeCancel:
		tx, err = c.TxChannelUpgradeCancel(opts, msg)
	case *chantypes.MsgChannelUpgradeTimeout:
		tx, err = c.TxChannelUpgradeTimeout(opts, msg)
	// case *transfertypes.MsgTransfer:
	// 	err = c.client.transfer(msg)
	default:
		logger.Error("failed to send msg", errors.New("illegal msg type"), "msg", msg)
		panic("illegal msg type")
	}
	return tx, err
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

func (c *Chain) getRevertReasonFromRpcError(err error) (string, []byte, error) {
	if de, ok := err.(rpc.DataError); !ok {
		return fmt.Sprintf("not an rpc error: errorType=%T: msg=%s", err, err.Error()), nil, nil
	} else if de.ErrorData() == nil {
		return "", nil, fmt.Errorf("failed without error data")
	} else if errorData, ok := de.ErrorData().(string); !ok {
		return "", nil, fmt.Errorf("failed with unexpected error data type: errorDataType=%T", de.ErrorData())
	} else {
		errorData := common.FromHex(errorData)
		revertReason, err := c.errorRepository.ParseError(errorData)
		if err != nil {
			return "", errorData, fmt.Errorf("failed to parse error: %v", err)
		}
		return revertReason, errorData, nil
	}
}

func (c *Chain) parseRpcError(err error) (string, string) {
	revertReason, rawErrorData, err := c.getRevertReasonFromRpcError(err)
	if err != nil {
		revertReason = fmt.Sprintf("failed to get revert reason: %s", err.Error())
	}
	// Note that Raw error data may be available even if revert reason isn't available.
	return revertReason, hex.EncodeToString(rawErrorData)
}

func estimateGas(
	ctx context.Context,
	c *Chain,
	tx *gethtypes.Transaction,
	doRound bool, // return rounded gas limit when gas limit is over
	logger *log.RelayLogger,
) (uint64, error) {
	if rawTxData, err := tx.MarshalBinary(); err != nil {
		logger.Error("failed to encode tx", err)
	} else {
		logger = &log.RelayLogger{Logger: logger.With(logAttrRawTxData, hex.EncodeToString(rawTxData))}
	}

	estimatedGas, err := c.client.EstimateGasFromTx(ctx, tx)
	if err != nil {
		if revertReason, rawErrorData, err := c.getRevertReasonFromRpcError(err); err != nil {
			// Raw error data may be available even if revert reason isn't available.
			logger = &log.RelayLogger{Logger: logger.With(logAttrRawErrorData, hex.EncodeToString(rawErrorData))}
			logger.Error("failed to get revert reason", err)
		} else {
			logger = &log.RelayLogger{Logger: logger.With(
				logAttrRawErrorData, hex.EncodeToString(rawErrorData),
				logAttrRevertReason, revertReason,
			)}
		}

		logger.Error("failed to estimate gas", err)
		return 0, err
	}

	txGasLimit := estimatedGas * c.Config().GasEstimateRate.Numerator / c.Config().GasEstimateRate.Denominator
	if txGasLimit > c.Config().MaxGasLimit {
		if !doRound {
			return 0, fmt.Errorf("estimated gas exceeds max gas limit")
		}

		logger.Warn("estimated gas exceeds max gas limit",
			logAttrEstimatedGas, txGasLimit,
			logAttrMaxGasLimit, c.Config().MaxGasLimit,
		)
		return c.Config().MaxGasLimit, nil
	}

	return txGasLimit, nil
}

type CallIter struct {
	msgs                       []sdk.Msg
	cursor                     int
	skipUpdateClientCommitment bool
	// for multicall
	txs          []gethtypes.Transaction
	msgTypeNames []string
}

func NewCallIter(msgs []sdk.Msg, skipUpdateClientCommitment bool) CallIter {
	msgTypeNames := make([]string, 0, len(msgs))
	for i := 0; i < len(msgs); i++ {
		msgTypeName := fmt.Sprintf("%T", msgs[i])
		msgTypeNames = append(msgTypeNames, msgTypeName)
	}

	return CallIter{
		msgs:                       msgs,
		cursor:                     0,
		skipUpdateClientCommitment: skipUpdateClientCommitment,
		msgTypeNames:               msgTypeNames,
	}
}
func (iter *CallIter) Cursor() int {
	return iter.cursor
}
func (iter *CallIter) Current() sdk.Msg {
	return iter.msgs[iter.cursor]
}
func (iter *CallIter) End() bool {
	return len(iter.msgs) <= iter.cursor
}
func (iter *CallIter) Next(n int) {
	iter.cursor = min(len(iter.msgs), iter.cursor+n)
}

func (iter *CallIter) updateLoggerMessageInfo(logger *log.RelayLogger, from int, count int) *log.RelayLogger {
	if from < 0 || count < 0 || len(iter.msgs) <= from || len(iter.msgs) < from+count {
		panic(fmt.Errorf("out of index: len(msgs)=%d, from=%d, count=%d", len(iter.msgs), from, count))
	}

	return &log.RelayLogger{
		Logger: logger.With(
			logAttrMsgIndexFrom, from,
			logAttrMsgCount, count,
			logAttrMsgType, strings.Join(iter.msgTypeNames[from:from+count], ","),
		),
	}
}

type CallIterBuildResult struct {
	tx    *gethtypes.Transaction
	count int
}

func (iter *CallIter) BuildTx(ctx context.Context, c *Chain) (*CallIterBuildResult, error) {
	if c.multicall3 == nil || len(iter.msgs) == 1 {
		return iter.buildSingleTx(ctx, c)
	} else {
		return iter.buildMultiTx(ctx, c)
	}
}

func (iter *CallIter) buildSingleTx(ctx context.Context, c *Chain) (*CallIterBuildResult, error) {
	if iter.End() {
		return nil, nil
	}

	logger := c.GetChainLogger()
	logger = iter.updateLoggerMessageInfo(logger, iter.Cursor(), 1)

	opts, err := c.TxOpts(ctx, true)
	if err != nil {
		return nil, err
	}
	opts.NoSend = true

	// gas estimation
	{
		opts.GasLimit = math.MaxUint64
		tx, err := c.BuildMessageTx(opts, iter.Current(), iter.skipUpdateClientCommitment)
		if err != nil {
			logger.Error("failed to build tx for gas estimation", err)
			return nil, err
		}

		txGasLimit, err := estimateGas(ctx, c, tx, true, logger)
		if err != nil {
			return nil, err
		}
		opts.GasLimit = txGasLimit
	}

	tx, err := c.BuildMessageTx(opts, iter.Current(), iter.skipUpdateClientCommitment)
	if err != nil {
		logger.Error("failed to build tx", err)
		return nil, err
	}

	return &CallIterBuildResult{tx, 1}, nil
}

func (iter *CallIter) buildMultiTx(ctx context.Context, c *Chain) (*CallIterBuildResult, error) {
	if iter.End() {
		return nil, nil
	}
	// now iter.cursor < len(iter.msgs)

	logger := c.GetChainLogger()

	opts, err := c.TxOpts(ctx, true)
	if err != nil {
		return nil, err
	}
	opts.NoSend = true
	opts.GasLimit = math.MaxUint64

	if iter.txs == nil { // create txs at first multicall call
		txs := make([]gethtypes.Transaction, 0, len(iter.msgs))

		for i := 0; i < len(iter.msgs); i++ {
			logger := iter.updateLoggerMessageInfo(logger, i, 1)

			// note that its nonce is not checked
			tx, err := c.BuildMessageTx(opts, iter.msgs[i], iter.skipUpdateClientCommitment)
			if err != nil {
				logger.Error("failed to build tx for gas estimation", err)
				return nil, err
			}
			if tx.To() == nil {
				err2 := fmt.Errorf("no target address")
				logger.Error("failed to construct Multicall3Call", err2)
				return nil, err2
			}
			txs = append(txs, *tx)
		}
		iter.txs = txs
	}

	var (
		lastOkCalls    []multicall3.Multicall3Call = nil
		lastOkGasLimit uint64                      = 0
	)
	count, err := findItems(
		len(iter.msgs)-iter.Cursor(),
		func(count int) error {
			from := iter.Cursor()
			to := from + count

			logger := iter.updateLoggerMessageInfo(logger, from, count)

			calls := make([]multicall3.Multicall3Call, 0, count)
			for i := from; i < to; i++ {
				calls = append(calls, multicall3.Multicall3Call{
					Target:   *iter.txs[i].To(),
					CallData: iter.txs[i].Data(),
				})
			}

			multiTx, err := c.multicall3.Aggregate(opts, calls)
			if err != nil {
				return err
			}

			txGasLimit, err := estimateGas(ctx, c, multiTx, 1 == count, logger)
			if err != nil {
				return err
			}

			lastOkGasLimit = txGasLimit
			lastOkCalls = calls
			return nil
		})

	logger = iter.updateLoggerMessageInfo(logger, iter.Cursor(), count)

	if err != nil {
		logger.Error("failed to prepare multicall tx", err)
		return nil, err
	}

	opts.GasLimit = lastOkGasLimit

	// add raw tx to log attribute
	tx, err := c.multicall3.Aggregate(opts, lastOkCalls)
	if err != nil {
		logger.Error("failed to build multicall tx with real send parameters", err)
		return nil, err
	}
	return &CallIterBuildResult{tx, count}, nil
}

func findItems(
	size int,
	f func(int) error,
) (int, error) {
	if size <= 0 {
		return 0, fmt.Errorf("empty items")
	}

	// hot path
	if err := f(size); err == nil {
		return size, nil
	}
	// binary search
	if i := sort.Search(size-1, func(n int) bool { return f(n+1) != nil }); i == 0 {
		return 0, fmt.Errorf("not found")
	} else {
		return i, nil
	}
}
