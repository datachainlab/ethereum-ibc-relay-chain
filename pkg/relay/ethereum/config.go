package ethereum

import (
	"encoding/hex"
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
	if c.GasEstimateRate.Numerator == 0 {
		errs = append(errs, fmt.Errorf("config attribute \"gas_estimate_rate.numerator\" is zero"))
	}
	if c.GasEstimateRate.Denominator == 0 {
		errs = append(errs, fmt.Errorf("config attribute \"gas_estimate_rate.denominator\" is zero"))
	}
	if c.MaxGasLimit == 0 {
		errs = append(errs, fmt.Errorf("config attribute \"max_gas_limit\" is zero"))
	}
	if c.Signer == nil {
		errs = append(errs, fmt.Errorf("config attribute \"signer\" is empty"))
	} else if err := c.Signer.GetCachedValue().(SignerConfig).Validate(); err != nil {
		errs = append(errs, fmt.Errorf("config attribute \"signer\" is invalid: %v", err))
	}
	if c.AllowLcFunctions != nil {
		if err := c.AllowLcFunctions.ValidateBasic(); err != nil {
			errs = append(errs, fmt.Errorf("config attribute \"allow_lc_functions\" is invalid: %v", err))
		}
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

func (alf AllowLCFunctionsConfig) ValidateBasic() error {
	if !common.IsHexAddress(alf.LcAddress) {
		return fmt.Errorf("invalid contract address: %s", alf.LcAddress)
	} else if alf.AllowAll && len(alf.Selectors) > 0 {
		return fmt.Errorf("allowAll is true and selectors is not empty")
	} else if !alf.AllowAll && len(alf.Selectors) == 0 {
		return fmt.Errorf("allowAll is false and selectors is empty")
	}
	return nil
}

// CONTRACT: alf.ValidateBasic() must be called before calling this method.
func (alf AllowLCFunctionsConfig) ToAllowLCFunctions() (*AllowLCFunctions, error) {
	if alf.AllowAll {
		return &AllowLCFunctions{
			LCAddress: common.HexToAddress(alf.LcAddress),
			AllowALL:  true,
		}, nil
	}
	selectors := make([][4]byte, len(alf.Selectors))
	for i, s := range alf.Selectors {
		bz, err := hex.DecodeString(strings.TrimPrefix(s, "0x"))
		if err != nil {
			return nil, fmt.Errorf("failed to decode selector: selector=%v err=%v", s, err)
		}
		if len(bz) != 4 {
			return nil, fmt.Errorf("invalid selector: %s", s)
		}
		copy(selectors[i][:], bz)
	}
	return &AllowLCFunctions{
		LCAddress: common.HexToAddress(alf.LcAddress),
		AllowALL:  false,
		Selectors: selectors,
	}, nil
}

type AllowLCFunctions struct {
	LCAddress common.Address
	AllowALL  bool
	Selectors [][4]byte
}

func (lcf AllowLCFunctions) IsAllowed(address common.Address, selector [4]byte) bool {
	if lcf.LCAddress != address {
		return false
	}
	if lcf.AllowALL {
		return true
	}
	for _, s := range lcf.Selectors {
		if s == selector {
			return true
		}
	}
	return false
}
