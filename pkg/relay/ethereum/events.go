package ethereum

import (
	"fmt"
	"math/big"
	"time"

	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hyperledger-labs/yui-relayer/core"
	"github.com/hyperledger-labs/yui-relayer/log"

	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/contract/ibchandler"
)

const (
	BlocksPerEventQueryDefault = 1000
)

var (
	abiGeneratedClientIdentifier,
	abiGeneratedConnectionIdentifier,
	abiGeneratedChannelIdentifier,
	abiSendPacket,
	abiRecvPacket,
	abiWriteAcknowledgement,
	abiAcknowledgePacket abi.Event
)

func init() {
	abiIBCHandler, err := ibchandler.IbchandlerMetaData.GetAbi()
	if err != nil {
		panic(err)
	}
	abiGeneratedClientIdentifier = abiIBCHandler.Events["GeneratedClientIdentifier"]
	abiGeneratedConnectionIdentifier = abiIBCHandler.Events["GeneratedConnectionIdentifier"]
	abiGeneratedChannelIdentifier = abiIBCHandler.Events["GeneratedChannelIdentifier"]
	abiSendPacket = abiIBCHandler.Events["SendPacket"]
	abiRecvPacket = abiIBCHandler.Events["RecvPacket"]
	abiWriteAcknowledgement = abiIBCHandler.Events["WriteAcknowledgement"]
	abiAcknowledgePacket = abiIBCHandler.Events["AcknowledgePacket"]

}

// filterPacketsWithActiveCommitment filters packets with non-zero packet commitments
func (chain *Chain) filterPacketsWithActiveCommitment(ctx core.QueryContext, packets core.PacketInfoList) (core.PacketInfoList, error) {

	var activePackets core.PacketInfoList

	for _, packet := range packets {
		commitmentPath := host.PacketCommitmentPath(packet.SourcePort, packet.SourceChannel, packet.Sequence)
		commitmentKey := crypto.Keccak256([]byte(commitmentPath))
		var fixedArray [32]byte
		copy(fixedArray[:], commitmentKey) // Copy the slice into the fixed array

		res, err := chain.ibcHandler.GetCommitment(
			chain.callOptsFromQueryContext(ctx),
			fixedArray,
		)

		if err != nil {
			return packets, err
		}

		if res == [32]byte{0} {
			continue
		}

		activePackets = append(activePackets, packet)

	}

	return activePackets, nil

}

func (chain *Chain) findSentPackets(ctx core.QueryContext, fromHeight uint64) (core.PacketInfoList, error) {
	logger := chain.GetChannelLogger()
	now := time.Now()

	var dstPortID, dstChannelID string
	if channel, found, err := chain.ibcHandler.GetChannel(
		chain.callOptsFromQueryContext(ctx),
		chain.Path().PortID,
		chain.Path().ChannelID,
	); err != nil {
		logger.Error("failed to get channel", err)
		return nil, err
	} else if !found {
		err := fmt.Errorf("channel not found")
		logger.Error("failed to get channel", err, "port_id", chain.Path().PortID, "channel_id", chain.Path().ChannelID)
		return nil, err
	} else {
		dstPortID = channel.Counterparty.PortId
		dstChannelID = channel.Counterparty.ChannelId
	}
	logs, err := chain.filterLogs(ctx, fromHeight, abiSendPacket)
	if err != nil {
		logger.Error("failed to filter logs", err)
		return nil, err
	}
	defer logger.TimeTrack(now, "findSentPackets", "num_logs", len(logs))

	var packets core.PacketInfoList
	for _, log := range logs {
		height := clienttypes.NewHeight(0, log.BlockNumber)

		sendPacket, err := chain.ibcHandler.ParseSendPacket(log)
		if err != nil {
			return nil, fmt.Errorf("failed to parse SendPacket event: err=%v, log=%v", err, log)
		}
		if sendPacket.SourceChannel != chain.Path().ChannelID || sendPacket.SourcePort != chain.Path().PortID {
			continue
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
	logger := chain.GetChannelLogger()
	now := time.Now()

	recvPacketEvents, err := chain.findRecvPacketEvents(ctx, fromHeight)
	if err != nil {
		logger.Error("failed to find recv packet events", err)
		return nil, err
	} else if len(recvPacketEvents) == 0 {
		return nil, nil
	}

	writeAckEvents, err := chain.findWriteAckEvents(ctx, recvPacketEvents[0].Raw.BlockNumber)
	if err != nil {
		logger.Error("failed to find write ack events", err)
		return nil, err
	} else if len(writeAckEvents) == 0 {
		return nil, nil
	}

	defer logger.TimeTrack(now, "findReceivedPackets", "num_recv_packet_events", len(recvPacketEvents), "num_write_ack_events", len(writeAckEvents))

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

func (chain *Chain) findRecvPacketEvents(ctx core.QueryContext, fromHeight uint64) ([]*ibchandler.IbchandlerRecvPacket, error) {
	logs, err := chain.filterLogs(ctx, fromHeight, abiRecvPacket)
	if err != nil {
		return nil, err
	}
	var events []*ibchandler.IbchandlerRecvPacket
	for _, log := range logs {
		event, err := chain.ibcHandler.ParseRecvPacket(log)
		if err != nil {
			return nil, fmt.Errorf("failed to parse RecvPacket event: err=%v, log=%v", err, log)
		}
		if event.Packet.DestinationChannel != chain.Path().ChannelID || event.Packet.DestinationPort != chain.Path().PortID {
			continue
		}
		events = append(events, event)
	}

	return events, nil
}

func (chain *Chain) findWriteAckEvents(ctx core.QueryContext, fromHeight uint64) ([]*ibchandler.IbchandlerWriteAcknowledgement, error) {
	logs, err := chain.filterLogs(ctx, fromHeight, abiWriteAcknowledgement)
	if err != nil {
		return nil, err
	}
	var events []*ibchandler.IbchandlerWriteAcknowledgement
	for _, log := range logs {
		event, err := chain.ibcHandler.ParseWriteAcknowledgement(log)
		if err != nil {
			return nil, fmt.Errorf("failed to parse WriteAcknowledgement event: err=%v, log=%v", err, log)
		}
		if event.DestinationChannel != chain.Path().ChannelID || event.DestinationPortId != chain.Path().PortID {
			continue
		}
		events = append(events, event)
	}

	return events, nil
}

func (chain *Chain) GetChannelLogger() *log.RelayLogger {
	logger := GetModuleLogger()
	if chain.Path() == nil {
		return logger
	}
	chainID := chain.Path().ChainID
	portID := chain.Path().PortID
	channelID := chain.Path().ChannelID
	return logger.WithChannel(chainID, portID, channelID)
}

func (chain *Chain) filterLogs(ctx core.QueryContext, fromHeight uint64, event abi.Event) ([]types.Log, error) {
	blocksPerEventQuery := chain.config.BlocksPerEventQuery
	if blocksPerEventQuery == 0 {
		blocksPerEventQuery = BlocksPerEventQueryDefault
	}

	toHeight := ctx.Height().GetRevisionHeight()
	totalBlocks := toHeight - fromHeight + 1
	loopCount := totalBlocks / blocksPerEventQuery
	if totalBlocks%blocksPerEventQuery != 0 {
		loopCount++
	}
	var logs []types.Log
	for i := uint64(0); i < loopCount; i++ {
		var endBlockNum uint64
		if i == loopCount-1 {
			endBlockNum = toHeight
		} else {
			endBlockNum = fromHeight + (i+1)*blocksPerEventQuery - 1
		}
		startBlock := big.NewInt(int64(fromHeight + i*blocksPerEventQuery))
		endBlock := big.NewInt(int64(endBlockNum))
		query := ethereum.FilterQuery{
			FromBlock: startBlock,
			ToBlock:   endBlock,
			Addresses: []common.Address{
				chain.config.IBCAddress(),
			},
			Topics: [][]common.Hash{{
				event.ID,
			}},
		}
		filterLogs, err := chain.client.FilterLogs(ctx.Context(), query)
		if err != nil {
			return nil, err
		}
		logs = append(logs, filterLogs...)
	}
	return logs, nil
}
