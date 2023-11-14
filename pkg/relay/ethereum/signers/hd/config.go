package hd

import (
	fmt "fmt"
	"math/big"

	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/relay/ethereum"
)

var _ ethereum.SignerConfig = (*SignerConfig)(nil)

func (c *SignerConfig) Validate() error {
	if len(c.Mnemonic) == 0 {
		return fmt.Errorf("SignerConfig attribute \"mnemonic\" is empty")
	}
	if len(c.Path) == 0 {
		return fmt.Errorf("SignerConfig attribute \"path\" is empty")
	}
	return nil
}

func (c *SignerConfig) Build(chainID *big.Int) (ethereum.Signer, error) {
	return NewSigner(chainID, c.Mnemonic, c.Path)
}
