package hd

import (
	fmt "fmt"

	"github.com/hyperledger-labs/yui-relayer/signer"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/wallet"
)

var _ signer.SignerConfig = (*SignerConfig)(nil)

func (c *SignerConfig) Validate() error {
	if _, err := wallet.GetPrvKeyFromMnemonicAndHDWPath(c.Mnemonic, c.Path); err != nil {
		return fmt.Errorf("invalid mnemonic and/or path for HD wallet: %v", err)
	}
	return nil
}

func (c *SignerConfig) Build() (signer.Signer, error) {
	return NewSigner(c.Mnemonic, c.Path)
}
