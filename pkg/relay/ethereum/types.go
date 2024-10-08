package ethereum

import (
	"slices"

	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	commitmenttypes "github.com/cosmos/ibc-go/v8/modules/core/23-commitment/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/contract/ibchandler"
)

func connectionEndToPB(conn ibchandler.ConnectionEndData) connectiontypes.ConnectionEnd {
	connpb := connectiontypes.ConnectionEnd{
		ClientId:    conn.ClientId,
		Versions:    []*connectiontypes.Version{},
		State:       connectiontypes.State(conn.State),
		DelayPeriod: conn.DelayPeriod,
		Counterparty: connectiontypes.Counterparty{
			ClientId:     conn.Counterparty.ClientId,
			ConnectionId: conn.Counterparty.ConnectionId,
			Prefix:       commitmenttypes.MerklePrefix(conn.Counterparty.Prefix),
		},
	}
	for _, v := range conn.Versions {
		ver := connectiontypes.Version(v)
		connpb.Versions = append(connpb.Versions, &ver)
	}
	return connpb
}

func channelToPB(chann ibchandler.ChannelData) channeltypes.Channel {
	return channeltypes.Channel{
		State:           channeltypes.State(chann.State),
		Ordering:        channeltypes.Order(chann.Ordering),
		Counterparty:    channeltypes.Counterparty(chann.Counterparty),
		ConnectionHops:  chann.ConnectionHops,
		Version:         chann.Version,
		UpgradeSequence: chann.UpgradeSequence,
	}
}

func upgradeToPB(upg ibchandler.UpgradeData) channeltypes.Upgrade {
	return channeltypes.NewUpgrade(
		channeltypes.NewUpgradeFields(
			channeltypes.Order(upg.Fields.Ordering),
			upg.Fields.ConnectionHops,
			upg.Fields.Version,
		),
		channeltypes.NewTimeout(
			clienttypes.Height(upg.Timeout.Height),
			upg.Timeout.Timestamp,
		),
		upg.NextSequenceSend,
	)
}

func pbToHandlerHeight(height exported.Height) ibchandler.HeightData {
	return ibchandler.HeightData{
		RevisionNumber: height.GetRevisionNumber(),
		RevisionHeight: height.GetRevisionHeight(),
	}
}

func pbToHostHeight(height exported.Height) ibchandler.HeightData {
	return ibchandler.HeightData{
		RevisionNumber: height.GetRevisionNumber(),
		RevisionHeight: height.GetRevisionHeight(),
	}
}

func pbToUpgradeFields(f channeltypes.UpgradeFields) ibchandler.UpgradeFieldsData {
	return ibchandler.UpgradeFieldsData{
		Ordering:       uint8(f.Ordering),
		ConnectionHops: slices.Clone(f.ConnectionHops),
		Version:        f.Version,
	}
}

func pbToTimeout(t channeltypes.Timeout) ibchandler.TimeoutData {
	return ibchandler.TimeoutData{
		Height:    pbToHandlerHeight(t.Height),
		Timestamp: t.Timestamp,
	}
}

func pbToUpgrade(u channeltypes.Upgrade) ibchandler.UpgradeData {
	return ibchandler.UpgradeData{
		Fields:           pbToUpgradeFields(u.Fields),
		Timeout:          pbToTimeout(u.Timeout),
		NextSequenceSend: u.NextSequenceSend,
	}
}

func pbToChannel(c channeltypes.Channel) ibchandler.ChannelData {
	return ibchandler.ChannelData{
		State:           uint8(c.State),
		Ordering:        uint8(c.Ordering),
		Counterparty:    ibchandler.ChannelCounterpartyData(c.Counterparty),
		ConnectionHops:  slices.Clone(c.ConnectionHops),
		Version:         c.Version,
		UpgradeSequence: c.UpgradeSequence,
	}
}
