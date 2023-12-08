package ethereum

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/utils"
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
	if c.LimitPriorityFeePerGas != "" {
		if _, err := utils.ParseEtherAmount(c.LimitPriorityFeePerGas); err != nil {
			errs = append(errs, fmt.Errorf("config attribute \"limit_priority_fee_per_gas\" is invalid: %v", err))
		}
	}
	if err := c.PriorityFeeRate.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("config attribute \"priority_fee_rate\" is invalid: %v", err))
	}
	if c.LimitFeePerGas != "" {
		if _, err := utils.ParseEtherAmount(c.LimitFeePerGas); err != nil {
			errs = append(errs, fmt.Errorf("config attribute \"limit_fee_per_gas\" is invalid: %v", err))
		}
	}
	if err := c.BaseFeeRate.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("config attribute \"base_fee_rate\" is invalid: %v", err))
	}

	return errors.Join(errs...)
}

func (f Fraction) Validate() error {
	if f.Denominator == 0 {
		return errors.New("zero is invalid fraction denominator")
	}
	return nil
}

// Mul multiplies `n` by `f` (this function mutates `n`)
func (f Fraction) Mul(n *big.Int) {
	n.Mul(n, new(big.Int).SetUint64(f.Numerator))
	n.Div(n, new(big.Int).SetUint64(f.Denominator))
}

func (c ChainConfig) GetLimitPriorityFeePerGas() *big.Int {
	if c.LimitPriorityFeePerGas == "" {
		return new(big.Int)
	} else if limit, err := utils.ParseEtherAmount(c.LimitPriorityFeePerGas); err != nil {
		panic(err)
	} else {
		return limit
	}
}

func (c ChainConfig) GetLimitFeePerGas() *big.Int {
	if c.LimitFeePerGas == "" {
		return new(big.Int)
	} else if limit, err := utils.ParseEtherAmount(c.LimitFeePerGas); err != nil {
		panic(err)
	} else {
		return limit
	}
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
