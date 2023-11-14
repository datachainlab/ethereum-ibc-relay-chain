package hd

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/relay/ethereum"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/wallet"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

var _ ethereum.Signer = (*Signer)(nil)

type Signer struct {
	signer types.Signer
	key    *ecdsa.PrivateKey
}

func NewSigner(chainID *big.Int, mnemonic, path string) (*Signer, error) {
	signer := types.LatestSignerForChainID(chainID)
	key, err := wallet.GetPrvKeyFromMnemonicAndHDWPath(mnemonic, path)
	if err != nil {
		return nil, fmt.Errorf("failed to extract a private key from the HD wallet")
	}
	return &Signer{signer, key}, nil
}

func (s *Signer) Address() common.Address {
	return crypto.PubkeyToAddress(s.key.PublicKey)
}

func (s *Signer) Sign(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
	if address != s.Address() {
		return nil, fmt.Errorf("unauthorized address: authorized=%v, given=%v", s.Address(), address)
	}

	sig, err := crypto.Sign(s.signer.Hash(tx).Bytes(), s.key)
	if err != nil {
		return nil, fmt.Errorf("failed to sign tx: %v", err)
	}

	return tx.WithSignature(s.signer, sig)
}
