package ethereum

import (
	"fmt"
	"math/big"
	"strings"

	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/contract/ibchandler"
	"github.com/hyperledger-labs/yui-relayer/core"
)

var (
	abiIBCHandler abi.ABI

	abiSendPacket,
	abiRecvPacket,
	abiWriteAcknowledgement abi.Event
)

func init() {
	var err error
	abiIBCHandler, err = abi.JSON(strings.NewReader(ibchandler.IbchandlerABI))
	if err != nil {
		panic(err)
	}
	abiSendPacket = abiIBCHandler.Events["SendPacket"]
	abiRecvPacket = abiIBCHandler.Events["RecvPacket"]
	abiWriteAcknowledgement = abiIBCHandler.Events["WriteAcknowledgement"]
}

func (chain *Chain) findPacket(
	ctx core.QueryContext,
	sourcePortID string,
	sourceChannel string,
	sequence uint64,
) (*channeltypes.Packet, error) {
	channel, found, err := chain.ibcHandler.GetChannel(
		chain.callOptsFromQueryContext(ctx),
		sourcePortID, sourceChannel,
	)
	if err != nil {
		return nil, err
	} else if !found {
		return nil, fmt.Errorf("channel not found: sourcePortID=%v sourceChannel=%v", sourcePortID, sourceChannel)
	}

	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(0),
		ToBlock:   new(big.Int).SetUint64(ctx.Height().GetRevisionHeight()),
		Addresses: []common.Address{
			chain.config.IBCAddress(),
		},
		Topics: [][]common.Hash{{
			abiSendPacket.ID,
		}},
	}
	logs, err := chain.client.FilterLogs(ctx.Context(), query)
	if err != nil {
		return nil, err
	}

	for _, log := range logs {
		if values, err := abiSendPacket.Inputs.Unpack(log.Data); err != nil {
			return nil, err
		} else {
			if l := len(values); l != 6 {
				return nil, fmt.Errorf("unexpected values length: expected=%v actual=%v", 6, l)
			}
			pSequence := values[0].(uint64)
			pSourcePortID := values[1].(string)
			pSourceChannel := values[2].(string)
			pTimeoutHeight := values[3].(struct {
				RevisionNumber uint64 "json:\"revision_number\""
				RevisionHeight uint64 "json:\"revision_height\""
			})
			pTimeoutTimestamp := values[4].(uint64)
			pData := values[5].([]uint8)

			if pSequence == sequence && pSourcePortID == sourcePortID && pSourceChannel == sourceChannel {
				return &channeltypes.Packet{
					Sequence:           pSequence,
					SourcePort:         pSourcePortID,
					SourceChannel:      pSourceChannel,
					DestinationPort:    channel.Counterparty.PortId,
					DestinationChannel: channel.Counterparty.ChannelId,
					Data:               pData,
					TimeoutHeight:      clienttypes.Height{RevisionNumber: pTimeoutHeight.RevisionNumber, RevisionHeight: pTimeoutHeight.RevisionHeight},
					TimeoutTimestamp:   pTimeoutTimestamp,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("packet not found: sourcePortID=%v sourceChannel=%v sequence=%v", sourcePortID, sourceChannel, sequence)
}

// getAllPackets returns all packets from events
func (chain *Chain) getAllPackets(
	ctx core.QueryContext,
	sourcePortID string,
	sourceChannel string,
) ([]*channeltypes.Packet, error) {
	channel, found, err := chain.ibcHandler.GetChannel(
		chain.callOptsFromQueryContext(ctx),
		sourcePortID, sourceChannel,
	)
	if err != nil {
		return nil, err
	} else if !found {
		return nil, fmt.Errorf("channel not found: sourcePortID=%v sourceChannel=%v", sourcePortID, sourceChannel)
	}

	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(0),
		ToBlock:   new(big.Int).SetUint64(ctx.Height().GetRevisionHeight()),
		Addresses: []common.Address{
			chain.config.IBCAddress(),
		},
		Topics: [][]common.Hash{{
			abiSendPacket.ID,
		}},
	}
	logs, err := chain.client.FilterLogs(ctx.Context(), query)
	if err != nil {
		return nil, err
	}

	var packets []*channeltypes.Packet
	for _, log := range logs {
		if values, err := abiSendPacket.Inputs.Unpack(log.Data); err != nil {
			return nil, err
		} else {
			if l := len(values); l != 6 {
				return nil, fmt.Errorf("unexpected values length: expected=%v actual=%v", 6, l)
			}
			pSequence := values[0].(uint64)
			pSourcePortID := values[1].(string)
			pSourceChannel := values[2].(string)
			pTimeoutHeight := values[3].(struct {
				RevisionNumber uint64 "json:\"revision_number\""
				RevisionHeight uint64 "json:\"revision_height\""
			})
			pTimeoutTimestamp := values[4].(uint64)
			pData := values[5].([]uint8)

			if pSourcePortID == sourcePortID && pSourceChannel == sourceChannel {
				packets = append(packets, &channeltypes.Packet{
					Sequence:           pSequence,
					SourcePort:         pSourcePortID,
					SourceChannel:      pSourceChannel,
					DestinationPort:    channel.Counterparty.PortId,
					DestinationChannel: channel.Counterparty.ChannelId,
					Data:               pData,
					TimeoutHeight:      clienttypes.Height{RevisionNumber: pTimeoutHeight.RevisionNumber, RevisionHeight: pTimeoutHeight.RevisionHeight},
					TimeoutTimestamp:   pTimeoutTimestamp,
				})
			}
		}
	}
	return packets, nil
}

func (chain *Chain) findAcknowledgement(
	ctx core.QueryContext,
	dstPortID string,
	dstChannel string,
	sequence uint64,
) ([]byte, error) {
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(0),
		ToBlock:   new(big.Int).SetUint64(ctx.Height().GetRevisionHeight()),
		Addresses: []common.Address{
			chain.config.IBCAddress(),
		},
		Topics: [][]common.Hash{{
			abiWriteAcknowledgement.ID,
		}},
	}
	logs, err := chain.client.FilterLogs(ctx.Context(), query)
	if err != nil {
		return nil, err
	}

	for _, log := range logs {
		if values, err := abiWriteAcknowledgement.Inputs.Unpack(log.Data); err != nil {
			return nil, err
		} else {
			if len(values) != 4 {
				return nil, fmt.Errorf("unexpected values: %v", values)
			}
			if dstPortID == values[0].(string) && dstChannel == values[1].(string) && sequence == values[2].(uint64) {
				return values[3].([]byte), nil
			}
		}
	}

	return nil, fmt.Errorf("ack not found: dstPortID=%v dstChannel=%v sequence=%v", dstPortID, dstChannel, sequence)
}

type PacketAcknowledgement struct {
	Sequence uint64
	Data     []byte
}

func (chain *Chain) getAllAcknowledgements(
	ctx core.QueryContext,
	dstPortID string,
	dstChannel string,
) ([]PacketAcknowledgement, error) {
	var acks []PacketAcknowledgement
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(0),
		ToBlock:   new(big.Int).SetUint64(ctx.Height().GetRevisionHeight()),
		Addresses: []common.Address{
			chain.config.IBCAddress(),
		},
		Topics: [][]common.Hash{{
			abiWriteAcknowledgement.ID,
		}},
	}
	logs, err := chain.client.FilterLogs(ctx.Context(), query)
	if err != nil {
		return nil, err
	}
	for _, log := range logs {
		if values, err := abiWriteAcknowledgement.Inputs.Unpack(log.Data); err != nil {
			return nil, err
		} else {
			if len(values) != 4 {
				return nil, fmt.Errorf("unexpected values: %v", values)
			}
			if dstPortID == values[0].(string) && dstChannel == values[1].(string) {
				acks = append(acks, PacketAcknowledgement{
					Sequence: values[2].(uint64),
					Data:     values[3].([]byte),
				})
			}
		}
	}
	return acks, nil
}

func (chain *Chain) findSentPackets(ctx core.QueryContext, fromHeight uint64) (core.PacketInfoList, error) {
	var dstPortID, dstChannelID string
	if channel, found, err := chain.ibcHandler.GetChannel(
		chain.callOptsFromQueryContext(ctx),
		chain.Path().PortID,
		chain.Path().ChannelID,
	); err != nil {
		return nil, err
	} else if !found {
		return nil, fmt.Errorf("channel not found: sourcePortID=%v sourceChannel=%v", chain.Path().PortID, chain.Path().ChannelID)
	} else {
		dstPortID = channel.Counterparty.PortId
		dstChannelID = channel.Counterparty.ChannelId
	}

	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromHeight)),
		ToBlock:   big.NewInt(int64(ctx.Height().GetRevisionHeight())),
		Addresses: []common.Address{
			chain.config.IBCAddress(),
		},
		Topics: [][]common.Hash{{
			abiSendPacket.ID,
		}},
	}

	logs, err := chain.client.FilterLogs(ctx.Context(), query)
	if err != nil {
		return nil, err
	}

	var packets core.PacketInfoList
	for _, log := range logs {
		height := clienttypes.NewHeight(0, log.BlockNumber)

		var sendPacket ibchandler.IbchandlerSendPacket
		if err := abiIBCHandler.UnpackIntoInterface(&sendPacket, "SendPacket", log.Data); err != nil {
			return nil, err
		}

		packet := &core.PacketInfo{
			Packet: channeltypes.NewPacket(
				sendPacket.Data,
				sendPacket.Sequence,
				sendPacket.SourcePort,
				sendPacket.SourceChannel,
				dstPortID,
				dstChannelID,
				clienttypes.Height(sendPacket.TimeoutHeight),
				sendPacket.TimeoutTimestamp,
			),
			Height: height,
		}
		packets = append(packets, packet)
	}

	return packets, nil
}

func (chain *Chain) findReceivedPackets(ctx core.QueryContext, fromHeight uint64) (core.PacketInfoList, error) {
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromHeight)),
		ToBlock:   big.NewInt(int64(ctx.Height().GetRevisionHeight())),
		Addresses: []common.Address{
			chain.config.IBCAddress(),
		},
		Topics: [][]common.Hash{{
			abiRecvPacket.ID,
			abiWriteAcknowledgement.ID,
		}},
	}

	logs, err := chain.client.FilterLogs(ctx.Context(), query)
	if err != nil {
		return nil, err
	}

	nLogs := len(logs)
	if nLogs%2 != 0 {
		return nil, fmt.Errorf("the number of found logs must be even number, but actually %d", nLogs)
	}
	var packets core.PacketInfoList
	for i := 0; i < nLogs; i += 2 {
		if rpBN, waBN := logs[i].BlockNumber, logs[i+1].BlockNumber; rpBN != waBN {
			return nil, fmt.Errorf("block number unmatch: %v != %v", rpBN, waBN)
		}
		if rpTI, waTI := logs[i].TxIndex, logs[i+1].TxIndex; rpTI != waTI {
			return nil, fmt.Errorf("tx index unmatch: %v != %v", rpTI, waTI)
		}
		height := clienttypes.NewHeight(0, logs[i].BlockNumber)

		var recvPacket ibchandler.IbchandlerRecvPacket
		if err := abiIBCHandler.UnpackIntoInterface(&recvPacket, "RecvPacket", logs[i].Data); err != nil {
			return nil, err
		}

		var writeAck ibchandler.IbchandlerWriteAcknowledgement
		if err := abiIBCHandler.UnpackIntoInterface(&writeAck, "WriteAcknowledgement", logs[i+1].Data); err != nil {
			return nil, err
		}

		packet := &core.PacketInfo{
			Packet: channeltypes.NewPacket(
				recvPacket.Packet.Data,
				recvPacket.Packet.Sequence,
				recvPacket.Packet.SourcePort,
				recvPacket.Packet.SourceChannel,
				recvPacket.Packet.DestinationPort,
				recvPacket.Packet.DestinationChannel,
				clienttypes.Height(recvPacket.Packet.TimeoutHeight),
				recvPacket.Packet.TimeoutTimestamp,
			),
			Acknowledgement: writeAck.Acknowledgement,
			Height:          height,
		}
		packets = append(packets, packet)
	}

	return packets, nil
}
