package ethereum_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/relay/ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hyperledger-labs/yui-relayer/core"
)

func TestCodec(t *testing.T) {
	codec := codec.NewProtoCodec(types.NewInterfaceRegistry())
	ethereum.RegisterInterfaces(codec.InterfaceRegistry())

	orig := ethereum.NewMsgID(common.HexToHash("0x123456789"))

	bz, err := codec.MarshalInterface(core.MsgID(orig))
	if err != nil {
		t.Fatalf("failed to marshal ethereum.MsgID into Any: %v", err)
	}

	var msgID core.MsgID
	if err := codec.UnmarshalInterface(bz, &msgID); err != nil {
		t.Fatalf("failed to unmarshal core.MsgID from Any: %v", err)
	}

	ethMsgID, ok := msgID.(*ethereum.MsgID)
	if !ok {
		t.Fatalf("unexpected type of core.MsgID instance: %T", msgID)
	}

	if *orig != *ethMsgID {
		t.Fatalf("unmatched ethereum.MsgID values: %v != %v", orig, *ethMsgID)
	}
}
