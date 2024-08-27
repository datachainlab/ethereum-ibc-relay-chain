package ethereum

import (
	"fmt"
	"time"

	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hyperledger-labs/yui-relayer/core"
)

var (
	_ core.MsgID     = (*MsgID)(nil)
	_ core.MsgResult = (*MsgResult)(nil)
)

func NewMsgID(txHash common.Hash) *MsgID {
	return &MsgID{
		TxHashHex: txHash.Hex(),
	}
}

func (*MsgID) Is_MsgID() {}
func (id *MsgID) TxHash() common.Hash {
	return common.HexToHash(id.TxHashHex)
}

type MsgResult struct {
	height       clienttypes.Height
	status       bool
	revertReason string
	events       []core.MsgEventLog
}

func (r *MsgResult) BlockHeight() clienttypes.Height {
	return r.height
}

func (r *MsgResult) Status() (bool, string) {
	return r.status, r.revertReason
}

func (r *MsgResult) Events() []core.MsgEventLog {
	return r.events
}

func (c *Chain) makeMsgResultFromReceipt(receipt *types.Receipt, revertReason string) (*MsgResult, error) {
	events, err := c.parseMsgEventLogs(receipt.Logs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse logs: %v", err)
	}
	return &MsgResult{
		height:       clienttypes.NewHeight(0, receipt.BlockNumber.Uint64()),
		status:       receipt.Status == types.ReceiptStatusSuccessful,
		revertReason: revertReason,
		events:       events,
	}, nil
}

func (c *Chain) parseMsgEventLogs(logs []*types.Log) ([]core.MsgEventLog, error) {
	var events []core.MsgEventLog
	for i, log := range logs {
		if len(log.Topics) == 0 {
			return nil, fmt.Errorf("log has no topic: logIndex=%d, log=%v", i, log)
		}

		var event core.MsgEventLog
		switch log.Topics[0] {
		case abiGeneratedClientIdentifier.ID:
			ev, err := c.ibcHandler.ParseGeneratedClientIdentifier(*log)
			if err != nil {
				return nil, fmt.Errorf("failed to parse GeneratedClientIdentifier event: logIndex=%d, log=%v", i, log)
			}
			event = &core.EventGenerateClientIdentifier{ID: ev.ClientId}
		case abiGeneratedConnectionIdentifier.ID:
			ev, err := c.ibcHandler.ParseGeneratedConnectionIdentifier(*log)
			if err != nil {
				return nil, fmt.Errorf("failed to parse GeneratedConnectionIdentifier event: logIndex=%d, log=%v", i, log)
			}
			event = &core.EventGenerateConnectionIdentifier{ID: ev.ConnectionId}
		case abiGeneratedChannelIdentifier.ID:
			ev, err := c.ibcHandler.ParseGeneratedChannelIdentifier(*log)
			if err != nil {
				return nil, fmt.Errorf("failed to parse GeneratedChannelIdentifier event: logIndex=%d, log=%v", i, log)
			}
			event = &core.EventGenerateChannelIdentifier{ID: ev.ChannelId}
		case abiSendPacket.ID:
			ev, err := c.ibcHandler.ParseSendPacket(*log)
			if err != nil {
				return nil, fmt.Errorf("failed to parse SendPacket event: logIndex=%d, log=%v", i, log)
			}
			event = &core.EventSendPacket{
				Sequence:         ev.Sequence,
				SrcPort:          ev.SourcePort,
				SrcChannel:       ev.SourceChannel,
				TimeoutHeight:    clienttypes.Height(ev.TimeoutHeight),
				TimeoutTimestamp: time.Unix(0, int64(ev.TimeoutTimestamp)),
				Data:             ev.Data,
			}
		case abiRecvPacket.ID:
			ev, err := c.ibcHandler.ParseRecvPacket(*log)
			if err != nil {
				return nil, fmt.Errorf("failed to parse RecvPacket event: logIndex=%d, log=%v", i, log)
			}
			event = &core.EventRecvPacket{
				Sequence:         ev.Packet.Sequence,
				DstPort:          ev.Packet.DestinationPort,
				DstChannel:       ev.Packet.DestinationChannel,
				TimeoutHeight:    clienttypes.Height(ev.Packet.TimeoutHeight),
				TimeoutTimestamp: time.Unix(0, int64(ev.Packet.TimeoutTimestamp)),
				Data:             ev.Packet.Data,
			}
		case abiWriteAcknowledgement.ID:
			ev, err := c.ibcHandler.ParseWriteAcknowledgement(*log)
			if err != nil {
				return nil, fmt.Errorf("failed to parse WriteAcknowledgement event: logIndex=%d, log=%v", i, log)
			}
			event = &core.EventWriteAcknowledgement{
				Sequence:        ev.Sequence,
				DstPort:         ev.DestinationPortId,
				DstChannel:      ev.DestinationChannel,
				Acknowledgement: ev.Acknowledgement,
			}
		case abiAcknowledgePacket.ID:
			ev, err := c.ibcHandler.ParseAcknowledgePacket(*log)
			if err != nil {
				return nil, fmt.Errorf("failed to parse AcknowledgePacket event: logIndex=%d, log=%v", i, log)
			}
			event = &core.EventAcknowledgePacket{
				Sequence:         ev.Packet.Sequence,
				SrcPort:          ev.Packet.SourcePort,
				SrcChannel:       ev.Packet.SourceChannel,
				TimeoutHeight:    clienttypes.Height(ev.Packet.TimeoutHeight),
				TimeoutTimestamp: time.Unix(0, int64(ev.Packet.TimeoutTimestamp)),
			}
		case abiChannelUpgradeOpen.ID:
			ev, err := c.ibcHandler.ParseChannelUpgradeOpen(*log)
			if err != nil {
				return nil, fmt.Errorf("failed to parse ChannelUpgradeOpen event: logIndex=%d, log=%v", i, log)
			}
			event = &core.EventUpgradeChannel{
				PortID:          ev.PortId,
				ChannelID:       ev.ChannelId,
				UpgradeSequence: ev.UpgradeSequence,
			}
		default:
			event = &core.EventUnknown{Value: log}
		}
		events = append(events, event)
	}
	return events, nil
}
