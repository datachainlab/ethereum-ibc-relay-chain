package ethereum

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/avast/retry-go"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	conntypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	chantypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	committypes "github.com/cosmos/ibc-go/v8/modules/core/23-commitment/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hyperledger-labs/yui-relayer/core"
	"github.com/hyperledger-labs/yui-relayer/signer"
	"github.com/hyperledger-labs/yui-relayer/log"

	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/client"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/contract/ibchandler"
)

const (
	PACKET_RECEIPT_NONE       uint8 = 0
	PACKET_RECEIPT_SUCCESSFUL uint8 = 1
)

type Chain struct {
	config ChainConfig

	pathEnd          *core.PathEnd
	homePath         string
	chainID          *big.Int
	codec            codec.ProtoCodecMarshaler
	msgEventListener core.MsgEventListener

	client     *client.ETHClient
	ibcHandler *ibchandler.Ibchandler

	ethereumSigner EthereumSigner

	errorRepository ErrorRepository

	// cache
	connectionOpenedConfirmed bool
	allowLCFunctions          *AllowLCFunctions
}

var _ core.Chain = (*Chain)(nil)

func NewChain(config ChainConfig) (*Chain, error) {
	id := big.NewInt(int64(config.EthChainId))
	client, err := client.NewETHClient(
		config.RpcAddr,
		client.WithRetryOption(
			retry.Attempts(uint(config.MaxRetryForInclusion)),
			retry.Delay(time.Duration(config.AverageBlockTimeMsec)*time.Millisecond),
		),
	)
	if err != nil {
		return nil, err
	}
	ibcHandler, err := ibchandler.NewIbchandler(config.IBCAddress(), client)
	if err != nil {
		return nil, err
	}
	signer, err := config.Signer.GetCachedValue().(signer.SignerConfig).Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build signer: %v", err)
	}
	ethereumSigner, err := NewEthereumSigner(signer, big.NewInt(int64(config.EthChainId)))
	if err != nil {
		return nil, fmt.Errorf("failed to build ethreum signer: %v", err)
	}

	var alfs *AllowLCFunctions
	if config.AllowLcFunctions != nil {
		alfs, err = config.AllowLcFunctions.ToAllowLCFunctions()
		if err != nil {
			return nil, fmt.Errorf("failed to build allowLcFunctions: %v", err)
		}
	}
	errorRepository, err := CreateErrorRepository(config.AbiPaths)
	if err != nil {
		return nil, fmt.Errorf("failed to create error repository: %v", err)
	}

	return &Chain{
		config:  config,
		client:  client,
		chainID: id,

		ibcHandler: ibcHandler,

		ethereumSigner: *ethereumSigner,

		errorRepository: errorRepository,

		allowLCFunctions: alfs,
	}, nil
}

// Config returns ChainConfig
func (c *Chain) Config() ChainConfig {
	return c.config
}

// Init ...
func (c *Chain) Init(homePath string, timeout time.Duration, codec codec.ProtoCodecMarshaler, debug bool) error {
	c.homePath = homePath
	c.codec = codec
	return nil
}

// SetupForRelay ...
func (c *Chain) SetupForRelay(ctx context.Context) error {
	return nil
}

// ChainID returns ID of the chain
func (c *Chain) ChainID() string {
	return c.config.ChainId
}

// GetLatestHeight gets the chain for the latest height and returns it
func (c *Chain) LatestHeight() (ibcexported.Height, error) {
	logger := c.GetChainLogger()
	bn, err := c.client.BlockNumber(context.TODO())
	if err != nil {
		logger.Error("failed to get block number", err)
		return nil, err
	}
	return clienttypes.NewHeight(0, bn), nil
}

func (c *Chain) Timestamp(height ibcexported.Height) (time.Time, error) {
	ht := big.NewInt(int64(height.GetRevisionHeight()))
	if header, err := c.client.HeaderByNumber(context.TODO(), ht); err != nil {
		return time.Time{}, err
	} else {
		return time.Unix(int64(header.Time), 0), nil
	}
}

func (c *Chain) AverageBlockTime() time.Duration {
	return time.Duration(c.config.AverageBlockTimeMsec) * time.Millisecond
}

// GetAddress returns the address of relayer
func (c *Chain) GetAddress() (sdk.AccAddress, error) {
	addr := make([]byte, 20)
	return addr, nil
}

// Marshaler returns the marshaler
func (c *Chain) Codec() codec.ProtoCodecMarshaler {
	return c.codec
}

// Client returns the RPC client for ethereum
func (c *Chain) Client() *client.ETHClient {
	return c.client
}

// SetRelayInfo sets source's path and counterparty's info to the chain
func (c *Chain) SetRelayInfo(p *core.PathEnd, _ *core.ProvableChain, _ *core.PathEnd) error {
	logger := c.GetChainLogger()
	if err := p.Validate(); err != nil {
		logger.Error("invalid path", err)
		return fmt.Errorf("path on chain %s failed to set: %w", c.ChainID(), err)
	}
	c.pathEnd = p
	return nil
}

func (c *Chain) Path() *core.PathEnd {
	return c.pathEnd
}

// RegisterMsgEventListener registers a given EventListener to the chain
func (c *Chain) RegisterMsgEventListener(listener core.MsgEventListener) {
	c.msgEventListener = listener
}

// QueryClientConsensusState retrevies the latest consensus state for a client in state at a given height
func (c *Chain) QueryClientConsensusState(ctx core.QueryContext, dstClientConsHeight ibcexported.Height) (*clienttypes.QueryConsensusStateResponse, error) {
	logger := c.GetChainLogger()
	defer logger.TimeTrack(time.Now(), "QueryClientConsensusState")
	s, found, err := c.ibcHandler.GetConsensusState(c.callOptsFromQueryContext(ctx), c.pathEnd.ClientID, pbToHostHeight(dstClientConsHeight))
	if err != nil {
		revertReason, data := c.parseRpcError(err)
		logger.Error("failed to get consensus state", err, logAttrRevertReason, revertReason, logAttrRawErrorData, data)
		return nil, err
	} else if !found {
		logger.Error("client consensus not found", errors.New("client consensus not found"))
		return nil, fmt.Errorf("client consensus not found: %v", c.pathEnd.ClientID)
	}
	var consensusState ibcexported.ConsensusState
	if err := c.Codec().UnmarshalInterface(s, &consensusState); err != nil {
		logger.Error("failed to unmarshal consensus state", err)
		return nil, err
	}
	any, err := clienttypes.PackConsensusState(consensusState)
	if err != nil {
		logger.Error("failed to pack consensus state", err)
		return nil, err
	}
	return clienttypes.NewQueryConsensusStateResponse(any, nil, ctx.Height().(clienttypes.Height)), nil
}

// QueryClientState returns the client state of dst chain
// height represents the height of dst chain
func (c *Chain) QueryClientState(ctx core.QueryContext) (*clienttypes.QueryClientStateResponse, error) {
	logger := c.GetChainLogger()
	defer logger.TimeTrack(time.Now(), "QueryClientState")
	s, found, err := c.ibcHandler.GetClientState(c.callOptsFromQueryContext(ctx), c.pathEnd.ClientID)
	if err != nil {
		revertReason, data := c.parseRpcError(err)
		logger.Error("failed to get client state", err, logAttrRevertReason, revertReason, logAttrRawErrorData, data)
		return nil, err
	} else if !found {
		logger.Error("client not found", errors.New("client not found"))
		return nil, fmt.Errorf("client not found: %v", c.pathEnd.ClientID)
	}
	var clientState ibcexported.ClientState
	if err := c.Codec().UnmarshalInterface(s, &clientState); err != nil {
		logger.Error("failed to unmarshal client state", err)
		return nil, err
	}
	any, err := clienttypes.PackClientState(clientState)
	if err != nil {
		logger.Error("failed to pack client state", err)
		return nil, err
	}
	return clienttypes.NewQueryClientStateResponse(any, nil, ctx.Height().(clienttypes.Height)), nil
}

var emptyConnRes = conntypes.NewQueryConnectionResponse(
	conntypes.NewConnectionEnd(
		conntypes.UNINITIALIZED,
		"client",
		conntypes.NewCounterparty(
			"client",
			"connection",
			committypes.NewMerklePrefix([]byte{}),
		),
		[]*conntypes.Version{},
		0,
	),
	[]byte{},
	clienttypes.NewHeight(0, 0),
)

// QueryConnection returns the remote end of a given connection
func (c *Chain) QueryConnection(ctx core.QueryContext) (*conntypes.QueryConnectionResponse, error) {
	logger := c.GetChainLogger()
	defer logger.TimeTrack(time.Now(), "QueryConnection")
	conn, found, err := c.ibcHandler.GetConnection(c.callOptsFromQueryContext(ctx), c.pathEnd.ConnectionID)
	if err != nil {
		revertReason, data := c.parseRpcError(err)
		logger.Error("failed to get connection", err, logAttrRevertReason, revertReason, logAttrRawErrorData, data)
		return nil, err
	} else if !found {
		return emptyConnRes, nil
	}
	return conntypes.NewQueryConnectionResponse(connectionEndToPB(conn), nil, ctx.Height().(clienttypes.Height)), nil
}

var emptyChannelRes = chantypes.NewQueryChannelResponse(
	chantypes.NewChannel(
		chantypes.UNINITIALIZED,
		chantypes.UNORDERED,
		chantypes.NewCounterparty(
			"port",
			"channel",
		),
		[]string{},
		"version",
	),
	[]byte{},
	clienttypes.NewHeight(0, 0),
)

// QueryChannel returns the channel associated with a channelID
func (c *Chain) QueryChannel(ctx core.QueryContext) (chanRes *chantypes.QueryChannelResponse, err error) {
	logger := c.GetChainLogger()
	defer logger.TimeTrack(time.Now(), "QueryChannel")
	chann, found, err := c.ibcHandler.GetChannel(c.callOptsFromQueryContext(ctx), c.pathEnd.PortID, c.pathEnd.ChannelID)
	if err != nil {
		revertReason, data := c.parseRpcError(err)
		logger.Error("failed to get channel", err, logAttrRevertReason, revertReason, logAttrRawErrorData, data)
		return nil, err
	} else if !found {
		return emptyChannelRes, nil
	}
	return chantypes.NewQueryChannelResponse(channelToPB(chann), nil, ctx.Height().(clienttypes.Height)), nil
}

// QueryUnreceivedPackets returns a list of unrelayed packet commitments
func (c *Chain) QueryUnreceivedPackets(ctx core.QueryContext, seqs []uint64) ([]uint64, error) {
	logger := c.GetChannelLogger()
	var ret []uint64
	var nextSequenceRecv uint64
	for _, seq := range seqs {
		var received bool
		var err error

		// With UNORDERED channel, we can use receipts to check if packets are already received.
		// With ORDERED channel, since IBC impls don't record receipts, we need to check nextSequenceRecv.
		switch c.Path().GetOrder() {
		case chantypes.UNORDERED:
			if rc, err := c.ibcHandler.GetPacketReceipt(c.callOptsFromQueryContext(ctx), c.pathEnd.PortID, c.pathEnd.ChannelID, seq); err != nil {
				revertReason, data := c.parseRpcError(err)
				logger.Error("failed to get packet receipt", err, logAttrRevertReason, revertReason, logAttrRawErrorData, data)
				return nil, err
			} else if rc == PACKET_RECEIPT_SUCCESSFUL {
				received = true
			} else if rc == PACKET_RECEIPT_NONE {
				received = false
			} else {
				return nil, fmt.Errorf("unknown receipt: %d", rc)
			}
		case chantypes.ORDERED:
			if nextSequenceRecv == 0 {
				// queried only once
				nextSequenceRecv, err = c.ibcHandler.GetNextSequenceRecv(c.callOptsFromQueryContext(ctx), c.pathEnd.PortID, c.pathEnd.ChannelID)
				if err != nil {
					revertReason, data := c.parseRpcError(err)
					logger.Error("failed to get nextSequenceRecv", err, logAttrRevertReason, revertReason, logAttrRawErrorData, data)
					return nil, err
				}
			}
			received = seq < nextSequenceRecv
		default:
			panic(fmt.Sprintf("unexpected order type: %d", c.Path().GetOrder()))
		}

		if !received {
			ret = append(ret, seq)
		}
	}
	return ret, nil
}

// QueryUnfinalizedRelayedPackets returns packets and heights that are sent but not received at the latest finalized block on the counterparty chain
func (c *Chain) QueryUnfinalizedRelayPackets(ctx core.QueryContext, counterparty core.LightClientICS04Querier) (core.PacketInfoList, error) {
	logger := c.GetChannelLogger()
	checkpoint, err := c.loadCheckpoint(sendCheckpoint)
	if err != nil {
		logger.Error("failed to load checkpoint", err)
		return nil, err
	}

	if checkpoint > ctx.Height().GetRevisionHeight() {
		logger.Info("`send` checkpoint is greater than target height", "checkpoint", checkpoint, "height", ctx.Height().GetRevisionHeight())
		return core.PacketInfoList{}, nil
	}
	packets, err := c.findSentPackets(ctx, checkpoint)
	if err != nil {
		logger.Error("failed to find sent packets", err)
		return nil, err
	}

	packets, err = c.filterPacketsWithActiveCommitment(ctx, packets)
	if err != nil {
		logger.Error("failed to filter packets with active commitment", err)
		return nil, err
	}

	counterpartyHeader, err := counterparty.GetLatestFinalizedHeader()
	if err != nil {
		logger.Error("failed to get latest finalized header", err)
		return nil, err
	}

	counterpartyCtx := core.NewQueryContext(context.TODO(), counterpartyHeader.GetHeight())
	seqs, err := counterparty.QueryUnreceivedPackets(counterpartyCtx, packets.ExtractSequenceList())
	if err != nil {
		logger.Error("failed to query unreceived packets", err)
		return nil, err
	}

	packets = packets.Filter(seqs)
	if len(packets) == 0 {
		checkpoint = ctx.Height().GetRevisionHeight() + 1
	} else {
		checkpoint = packets[0].EventHeight.GetRevisionHeight()
	}
	if err := c.saveCheckpoint(checkpoint, sendCheckpoint); err != nil {
		logger.Error("failed to save checkpoint", err)
		return nil, err
	}

	return packets, nil
}

// QueryUnreceivedAcknowledgements returns a list of unrelayed packet acks
func (c *Chain) QueryUnreceivedAcknowledgements(ctx core.QueryContext, seqs []uint64) ([]uint64, error) {
	logger := c.GetChannelLogger()
	var ret []uint64
	for _, seq := range seqs {
		key := crypto.Keccak256Hash(host.PacketCommitmentKey(c.pathEnd.PortID, c.pathEnd.ChannelID, seq))
		commitment, err := c.ibcHandler.GetCommitment(c.callOptsFromQueryContext(ctx), key)
		if err != nil {
			revertReason, data := c.parseRpcError(err)
			logger.Error("failed to get hashed packet commitment", err, logAttrRevertReason, revertReason, logAttrRawErrorData, data)
			return nil, err
		} else if commitment != [32]byte{} {
			ret = append(ret, seq)
		}
	}
	return ret, nil
}

// QueryUnfinalizedRelayedAcknowledgements returns acks and heights that are sent but not received at the latest finalized block on the counterpartychain
func (c *Chain) QueryUnfinalizedRelayAcknowledgements(ctx core.QueryContext, counterparty core.LightClientICS04Querier) (core.PacketInfoList, error) {
	logger := c.GetChannelLogger()
	checkpoint, err := c.loadCheckpoint(recvCheckpoint)
	if err != nil {
		logger.Error("failed to load checkpoint", err)
		return nil, err
	}

	if checkpoint > ctx.Height().GetRevisionHeight() {
		logger.Info("`recv` checkpoint is greater than target height", "checkpoint", checkpoint, "height", ctx.Height().GetRevisionHeight())
		return core.PacketInfoList{}, nil
	}
	packets, err := c.findReceivedPackets(ctx, checkpoint)
	if err != nil {
		logger.Error("failed to find received packets", err)
		return nil, err
	}

	counterpartyHeader, err := counterparty.GetLatestFinalizedHeader()
	if err != nil {
		logger.Error("failed to get latest finalized header", err)
		return nil, err
	}

	counterpartyCtx := core.NewQueryContext(context.TODO(), counterpartyHeader.GetHeight())
	seqs, err := counterparty.QueryUnreceivedAcknowledgements(counterpartyCtx, packets.ExtractSequenceList())
	if err != nil {
		logger.Error("failed to query unreceived acknowledgements", err)
		return nil, err
	}

	packets = packets.Filter(seqs)
	if len(packets) == 0 {
		checkpoint = ctx.Height().GetRevisionHeight() + 1
	} else {
		checkpoint = packets[0].EventHeight.GetRevisionHeight()
	}
	if err := c.saveCheckpoint(checkpoint, recvCheckpoint); err != nil {
		logger.Error("failed to save checkpoint", err)
		return nil, err
	}

	return packets, nil
}

// QueryBalance returns the amount of coins in the relayer account
func (c *Chain) QueryBalance(ctx core.QueryContext, address sdk.AccAddress) (sdk.Coins, error) {
	panic("not supported")
}

// QueryDenomTraces returns all the denom traces from a given chain
func (c *Chain) QueryDenomTraces(ctx core.QueryContext, offset uint64, limit uint64) (*transfertypes.QueryDenomTracesResponse, error) {
	panic("not supported")
}

func (c *Chain) callOptsFromQueryContext(ctx core.QueryContext) *bind.CallOpts {
	return c.CallOpts(ctx.Context(), int64(ctx.Height().GetRevisionHeight()))
}

func (c *Chain) GetChainLogger() *log.RelayLogger {
	logger := GetModuleLogger()
	if c.Path() == nil {
		return logger
	}
	chainID := c.Path().ChainID
	return logger.WithChain(chainID)
}

func (c *Chain) confirmConnectionOpened(ctx context.Context) (bool, error) {
	if c.connectionOpenedConfirmed {
		return true, nil
	}
	if c.pathEnd.ConnectionID == "" {
		return false, nil
	}
	latestHeight, err := c.LatestHeight()
	if err != nil {
		return false, err
	}
	// NOTE: err is nil if the connection not found
	connRes, err := c.QueryConnection(
		core.NewQueryContext(ctx, latestHeight),
	)
	if err != nil {
		return false, err
	}
	if connRes.Connection.State != conntypes.OPEN {
		return false, nil
	}
	c.connectionOpenedConfirmed = true
	return true, nil
}
