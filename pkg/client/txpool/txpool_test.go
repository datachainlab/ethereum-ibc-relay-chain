package txpool

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"
	"unsafe"

	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/ethereum/go-ethereum/node"
)

func makeGenesisAlloc(t *testing.T, n int) ([]common.Address, types.GenesisAlloc) {
	var addrs []common.Address
	alloc := make(types.GenesisAlloc)
	for i := 0; i < n; i++ {
		key, err := crypto.GenerateKey()
		if err != nil {
			t.Fatalf("failed to generate a private key: err=%v", err)
		}

		addr := crypto.PubkeyToAddress(key.PublicKey)

		addrs = append(addrs, addr)
		alloc[addr] = types.Account{
			Balance:    new(big.Int).Lsh(common.Big1, 250),
			PrivateKey: crypto.FromECDSA(key),
		}
	}
	return addrs, alloc
}

func getPrivateKey(t *testing.T, account types.Account) *ecdsa.PrivateKey {
	key, err := crypto.ToECDSA(account.PrivateKey)
	if err != nil {
		t.Fatalf("failed to unmarshal the private key: err=%v", err)
	}
	return key
}

func getEthClient(t *testing.T, sim *simulated.Backend) client.IETHClient {
	type simClient struct {
		*ethclient.Client
	}
	ifceCli := sim.Client()
	ptrCli := unsafe.Add(unsafe.Pointer(&ifceCli), unsafe.Sizeof(uintptr(0)))
	cli, err := client.NewETHClientWith((*simClient)(ptrCli).Client)
	if err != nil {
		t.Fatalf("failed to create client: err=%v", err)
	}
	return cli
}

func transfer(t *testing.T, ctx context.Context, cl client.IETHClient, signer types.Signer, key *ecdsa.PrivateKey, nonce uint64, gasTipCap, gasFeeCap *big.Int, to common.Address, amount *big.Int) {
	tx := types.NewTx(&types.DynamicFeeTx{
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       21000,
		To:        &to,
		Value:     amount,
	})

	tx, err := types.SignTx(tx, signer, key)
	if err != nil {
		t.Fatalf("failed to sign tx: err=%v", err)
	}

	if err := cl.Inner().SendTransaction(ctx, tx); err != nil {
		t.Fatalf("failed to send tx: err=%v", err)
	}
}

func replace(t *testing.T, ctx context.Context, cl client.IETHClient, signer types.Signer, key *ecdsa.PrivateKey, priceBump uint64, nonce uint64, gasTipCap, gasFeeCap *big.Int, to common.Address, amount *big.Int) {
	addr := crypto.PubkeyToAddress(key.PublicKey)

	if _, minFeeCap, minTipCap, err := GetMinimumRequiredFee(ctx, cl, addr, nonce, priceBump); err != nil {
		t.Fatalf("failed to get the minimum fee required to replace tx: err=%v", err)
	} else if minFeeCap.Cmp(common.Big0) == 0 {
		t.Fatalf("tx to replace not found")
	} else {
		t.Logf("minimum required fees: feeCap=%v, tipCap=%v", minFeeCap, minTipCap)
		if gasFeeCap.Cmp(minFeeCap) < 0 {
			t.Logf("gasFeeCap updated: %v => %v", gasFeeCap, minFeeCap)
			gasFeeCap = minFeeCap
		}
		if gasTipCap.Cmp(minTipCap) < 0 {
			t.Logf("gasTipCap updated: %v => %v", gasTipCap, minTipCap)
			gasTipCap = minTipCap
		}
	}

	transfer(t, ctx, cl, signer, key, nonce, gasTipCap, gasFeeCap, to, amount)
}

func TestContentFrom(t *testing.T) {
	ctx := context.Background()

	addrs, alloc := makeGenesisAlloc(t, 2)
	sender, receiver := addrs[0], addrs[1]
	senderKey := getPrivateKey(t, alloc[sender])

	priceBump := uint64(25)
	sim := simulated.NewBackend(alloc, func(nodeConf *node.Config, ethConf *ethconfig.Config) {
		t.Logf("original price bump is %v", ethConf.TxPool.PriceBump)
		ethConf.TxPool.PriceBump = priceBump
	})
	defer sim.Close()

	cli := getEthClient(t, sim)

	// make signer
	chainID, err := cli.Inner().ChainID(ctx)
	if err != nil {
		t.Fatalf("failed to get chain ID: err=%v", err)
	}
	signer := types.LatestSignerForChainID(chainID)

	// we use small fee cap to prepend txs from being included
	feeCap := big.NewInt(1_000_000)
	tipCap := big.NewInt(1_000_000)
	amount := big.NewInt(1_000_000_000) // 1 GWei

	transfer(t, ctx, cli, signer, senderKey, 0, tipCap, feeCap, receiver, amount)
	for i := 0; i < 10; i++ {
		replace(t, ctx, cli, signer, senderKey, priceBump, 0, tipCap, feeCap, receiver, amount)
	}

	// check block info
	block, err := cli.Inner().BlockByNumber(ctx, nil)
	if err != nil {
		t.Fatalf("failed to get block by number: err=%v", err)
	} else {
		t.Logf("bn=%v, basefee=%v", block.Number(), block.BaseFee())
	}

	// commit block
	sim.Commit()

	// check block info
	block, err = cli.Inner().BlockByNumber(ctx, nil)
	if err != nil {
		t.Fatalf("failed to get block by number: err=%v", err)
	} else {
		t.Logf("bn=%v, basefee=%v", block.Number(), block.BaseFee())
	}

	// check that len(pendingTxs) == 1
	if pendingTxs, err := PendingTransactions(ctx, cli, sender); err != nil {
		t.Fatalf("failed to get pending transactions: err=%v", err)
	} else if len(pendingTxs) != 1 {
		t.Fatalf("unexpected pending txs: pendingTxs=%v", pendingTxs)
	}

	// update fee cap to allow the tx to be included
	feeCap = block.BaseFee()

	replace(t, ctx, cli, signer, senderKey, priceBump, 0, tipCap, feeCap, receiver, amount)

	// commit block
	sim.Commit()

	// check block info
	block, err = cli.Inner().BlockByNumber(ctx, nil)
	if err != nil {
		t.Fatalf("failed to get block by number: err=%v", err)
	} else {
		t.Logf("bn=%v, basefee=%v", block.Number(), block.BaseFee())
	}

	// check that len(pendingTxs) == 0
	if pendingTxs, err := PendingTransactions(ctx, cli, sender); err != nil {
		t.Fatalf("failed to get pending transactions: err=%v", err)
	} else if len(pendingTxs) != 0 {
		t.Fatalf("unexpected pending txs: pendingTxs=%v", pendingTxs)
	}
}
