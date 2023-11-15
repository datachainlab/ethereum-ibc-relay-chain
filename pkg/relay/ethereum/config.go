package ethereum

import (
	"errors"
	"fmt"
	"strings"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hyperledger-labs/yui-relayer/core"
)

var (
	_ core.ChainConfig                   = (*ChainConfig)(nil)
	_ codectypes.UnpackInterfacesMessage = (*ChainConfig)(nil)
)

func (c ChainConfig) Build() (core.Chain, error) {
	return NewChain(c)
}

func (c ChainConfig) Validate() error {
	isEmpty := func(s string) bool {
		return strings.TrimSpace(s) == ""
	}

	var errs []error
	if isEmpty(c.ChainId) {
		errs = append(errs, fmt.Errorf("config attribute \"chain_id\" is empty"))
	}
	if isEmpty(c.RpcAddr) {
		errs = append(errs, fmt.Errorf("config attribute \"rpc_addr\" is empty"))
	}
	if isEmpty(c.IbcAddress) {
		errs = append(errs, fmt.Errorf("config attribute \"ibc_address\" is empty"))
	}
	if c.AverageBlockTimeMsec == 0 {
		errs = append(errs, fmt.Errorf("config attribute \"average_block_time_msec\" is zero"))
	}
	if c.MaxRetryForInclusion == 0 {
		errs = append(errs, fmt.Errorf("config attribute \"max_retry_for_inclusion\" is zero"))
	}
	if c.Signer == nil {
		errs = append(errs, fmt.Errorf("config attribute \"signer\" is empty"))
	} else if err := c.Signer.GetCachedValue().(SignerConfig).Validate(); err != nil {
		errs = append(errs, fmt.Errorf("config attribute \"signer\" is invalid: %v", err))
	}

	return errors.Join(errs...)
}

func (c ChainConfig) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	if err := unpacker.UnpackAny(c.Signer, new(SignerConfig)); err != nil {
		return fmt.Errorf("failed to unpack ChainConfig attribute \"signer\": %v", err)
	}
	return nil
}

func (c ChainConfig) IBCAddress() common.Address {
	return common.HexToAddress(c.IbcAddress)
}
