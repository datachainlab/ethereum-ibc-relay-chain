package hd

import (
	fmt "fmt"
	"math/big"

	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/relay/ethereum"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/wallet"
)

var _ ethereum.SignerConfig = (*SignerConfig)(nil)

func (c *SignerConfig) Validate() error {
	if _, err := wallet.GetPrvKeyFromMnemonicAndHDWPath(c.Mnemonic, c.Path); err != nil {
		return fmt.Errorf("invalid mnemonic and/or path for HD wallet: %v", err)
	}
	return nil
}

func (c *SignerConfig) Build(chainID *big.Int) (ethereum.Signer, error) {
	return NewSigner(chainID, c.Mnemonic, c.Path)
}
