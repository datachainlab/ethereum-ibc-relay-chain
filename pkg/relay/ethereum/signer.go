package ethereum

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	gethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/hyperledger-labs/yui-relayer/log"
	"github.com/hyperledger-labs/yui-relayer/signer"
)

type EthereumSigner struct {
	bytesSigner  signer.Signer
	gethSigner   gethtypes.Signer
	addressCache common.Address
	logger       *log.RelayLogger
	NoSign       bool
}

func NewEthereumSigner(ctx context.Context, bytesSigner signer.Signer, chainID *big.Int) (*EthereumSigner, error) {
	pkbytes, err := bytesSigner.GetPublicKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("fail to get public key")
	}

	pk, err := gethcrypto.DecompressPubkey(pkbytes)
	if err != nil {
		return nil, fmt.Errorf("fail to decompress public key")
	}

	addr := gethcrypto.PubkeyToAddress(*pk)

	gethSigner := gethtypes.LatestSignerForChainID(chainID)

	return &EthereumSigner{
		bytesSigner:  bytesSigner,
		gethSigner:   gethSigner,
		addressCache: addr,
		logger:       nil,
		NoSign:       false,
	}, nil
}

func (s *EthereumSigner) GetLogger() *log.RelayLogger {
	return s.logger
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

	if s.NoSign {
		return tx, nil
	}

	txHash := s.gethSigner.Hash(tx)

	if s.logger != nil {
		s.logger.Info("try to sign", "address", address, "txHash", txHash.Hex())
	}

	// NOTE: This method is called from methods in the go-ethereum package so we cannot pass a context to this method,
	//   which means that we cannot cancel Sign method even if the process receives a signal to stop.
	//   Although we can set a context in a similar way to SetLogger does, leave context.TODO() for now
	//   because it seems rare to receive a signal while signing a transaction.
	sig, err := s.bytesSigner.Sign(context.TODO(), txHash.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to sign tx: %v", err)
	}

	return tx.WithSignature(s.gethSigner, sig)
}
