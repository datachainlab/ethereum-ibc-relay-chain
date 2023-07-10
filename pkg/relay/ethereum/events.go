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
			EventHeight: height,
		}
		packets = append(packets, packet)
	}

	return packets, nil
}

func (chain *Chain) findReceivedPackets(ctx core.QueryContext, fromHeight uint64) (core.PacketInfoList, error) {
	recvPacketEvents, err := chain.findRecvPacketEvents(ctx, fromHeight)
	if err != nil {
		return nil, err
	} else if len(recvPacketEvents) == 0 {
		return nil, nil
	}

	writeAckEvents, err := chain.findWriteAckEvents(ctx, fromHeight)
	if err != nil {
		return nil, err
	} else if len(writeAckEvents) == 0 {
		return nil, nil
	}

	var packets core.PacketInfoList
	for _, rp := range recvPacketEvents {
		for _, wa := range writeAckEvents {
			if rp.Packet.Sequence == wa.Sequence {
				packets = append(packets, &core.PacketInfo{
					Packet: channeltypes.Packet{
						Sequence:           rp.Packet.Sequence,
						SourcePort:         rp.Packet.SourcePort,
						SourceChannel:      rp.Packet.SourceChannel,
						DestinationPort:    rp.Packet.DestinationPort,
						DestinationChannel: rp.Packet.DestinationChannel,
						Data:               rp.Packet.Data,
						TimeoutHeight:      clienttypes.Height(rp.Packet.TimeoutHeight),
						TimeoutTimestamp:   rp.Packet.TimeoutTimestamp,
					},
					Acknowledgement: wa.Acknowledgement,
					EventHeight:     clienttypes.NewHeight(0, rp.Raw.BlockNumber),
				})
				break
			}
		}
	}

	return packets, nil
}

func (chain *Chain) findRecvPacketEvents(ctx core.QueryContext, fromHeight uint64) ([]ibchandler.IbchandlerRecvPacket, error) {
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromHeight)),
		ToBlock:   big.NewInt(int64(ctx.Height().GetRevisionHeight())),
		Addresses: []common.Address{
			chain.config.IBCAddress(),
		},
		Topics: [][]common.Hash{{
			abiRecvPacket.ID,
		}},
	}

	logs, err := chain.client.FilterLogs(ctx.Context(), query)
	if err != nil {
		return nil, err
	}

	var events []ibchandler.IbchandlerRecvPacket
	for _, log := range logs {
		var event ibchandler.IbchandlerRecvPacket
		if err := abiIBCHandler.UnpackIntoInterface(&event, "RecvPacket", log.Data); err != nil {
			return nil, err
		}
		event.Raw = log
		events = append(events, event)
	}

	return events, nil
}

func (chain *Chain) findWriteAckEvents(ctx core.QueryContext, fromHeight uint64) ([]ibchandler.IbchandlerWriteAcknowledgement, error) {
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromHeight)),
		ToBlock:   big.NewInt(int64(ctx.Height().GetRevisionHeight())),
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

	var events []ibchandler.IbchandlerWriteAcknowledgement
	for _, log := range logs {
		var event ibchandler.IbchandlerWriteAcknowledgement
		if err := abiIBCHandler.UnpackIntoInterface(&event, "WriteAcknowledgement", log.Data); err != nil {
			return nil, err
		}
		event.Raw = log
		events = append(events, event)
	}

	return events, nil
}
