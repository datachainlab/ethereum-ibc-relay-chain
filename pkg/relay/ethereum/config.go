package ethereum

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/hyperledger-labs/yui-relayer/core"
)

var _ core.ChainConfig = (*ChainConfig)(nil)

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
	if c.EthChainId <= 0 {
		errs = append(errs, fmt.Errorf("config attribute \"eth_chain_id\" must be greater than zero: %v", c.EthChainId))
	}
	if isEmpty(c.RpcAddr) {
		errs = append(errs, fmt.Errorf("config attribute \"rpc_addr\" is empty"))
	}
	if isEmpty(c.HdwMnemonic) {
		errs = append(errs, fmt.Errorf("config attribute \"hdw_mnemonic\" is empty"))
	}
	if isEmpty(c.HdwPath) {
		errs = append(errs, fmt.Errorf("config attribute \"hdw_path\" is empty"))
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

	return errors.Join(errs...)
}

func (c ChainConfig) IBCAddress() common.Address {
	return common.HexToAddress(c.IbcAddress)
}
