package ethereum

import (
	"fmt"
	"math/big"

	"github.com/hyperledger-labs/yui-relayer/log"
	"github.com/hyperledger-labs/yui-relayer/core"
	"github.com/ethereum/go-ethereum/common"
	gethcrypto "github.com/ethereum/go-ethereum/crypto"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
)

type EthereumSigner struct {
	bytesSigner core.Signer
	gethSigner gethtypes.Signer
	addressCache common.Address
	logger *log.RelayLogger
}

func NewEthereumSigner(bytesSigner core.Signer, chainID *big.Int) (*EthereumSigner, error) {
	pkbytes, err := bytesSigner.GetPublicKey()
	if err != nil {
		return nil, fmt.Errorf("fail to get public key")
	}

	pk, err := gethcrypto.DecompressPubkey(pkbytes)
	if err != nil {
		return nil, fmt.Errorf("fail to decompress public key")
	}

	addr := gethcrypto.PubkeyToAddress(*pk)

	gethSigner := gethtypes.LatestSignerForChainID(chainID)

	return &EthereumSigner {
		bytesSigner: bytesSigner,
		gethSigner: gethSigner,
		addressCache: addr,
		logger: nil,
	}, nil
}

func (s *EthereumSigner) SetLogger(logger *log.RelayLogger) {
	s.logger = logger
}

func (s *EthereumSigner) Address() common.Address {
	return s.addressCache
}

func (s *EthereumSigner) Sign(address common.Address, tx *gethtypes.Transaction) (*gethtypes.Transaction, error) {
	if address != s.Address() {
		return nil, fmt.Errorf("unauthorized address: authorized=%v, given=%v", s.Address(), address)
	}

	txHash := s.gethSigner.Hash(tx)

	if (s.logger != nil) {
		s.logger.Info("try to sign", "address", address, "txHash", txHash.Hex());
	}

	sig, err := s.bytesSigner.Sign(txHash.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to sign tx: %v", err)
	}

	return tx.WithSignature(s.gethSigner, sig)
}
