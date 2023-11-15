package ethereum

import (
	"math/big"

	"github.com/cosmos/gogoproto/proto"
	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
)

type SignerConfig interface {
	proto.Message
	Build(chainID *big.Int) (Signer, error)
	Validate() error
}

type Signer interface {
	Sign(common.Address, *gethtypes.Transaction) (*gethtypes.Transaction, error)
	Address() common.Address
}
