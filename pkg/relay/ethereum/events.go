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
			EventHeight:     height,
		}
		packets = append(packets, packet)
	}

	return packets, nil
}
