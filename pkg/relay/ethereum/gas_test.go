package ethereum

import (
	"context"
	"math/big"
	"testing"

	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/client"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/client/txpool"
	"github.com/ethereum/go-ethereum"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

func Test_TxOpts_LegacyTx(t *testing.T) {
	ethClient, err := client.NewETHClient("https://bsc-dataseed1.binance.org/")
	if err != nil {
		t.Fatal(err)
	}
	config := createConfig()
	config.TxType = "legacy"
	calculator := NewGasFeeCalculator(ethClient, config)
	txOpts := &bind.TransactOpts{}
	if err = calculator.Apply(context.Background(), txOpts); err != nil {
		t.Fatal(err)
	}
	if txOpts.GasTipCap != nil {
		t.Error("gasTipCap must be nil")
	}
	if txOpts.GasFeeCap != nil {
		t.Error("gasFeeCap must be nil")
	}
	if txOpts.GasPrice == nil || txOpts.GasPrice.Cmp(big.NewInt(0)) == 0 {
		t.Error("gasPrice must be suggested")
	}
}

func Test_TxOpts_DynamicTx(t *testing.T) {
	ethClient, err := client.NewETHClient("https://ethereum.publicnode.com")
	if err != nil {
		t.Fatal(err)
	}
	config := createConfig()
	calculator := NewGasFeeCalculator(ethClient, config)
	txOpts := &bind.TransactOpts{}
	if err = calculator.Apply(context.Background(), txOpts); err != nil {
		t.Fatal(err)
	}
	if txOpts.GasTipCap == nil || txOpts.GasTipCap.Cmp(big.NewInt(0)) == 0 {
		t.Error("gasTipCap must be suggested")
	}
	if txOpts.GasFeeCap == nil || txOpts.GasFeeCap.Cmp(big.NewInt(0)) == 0 {
		t.Error("gasFeeCap must be suggested")
	}
	if txOpts.GasPrice != nil {
		t.Error("gasPrice must be nil")
	}
}

func Test_TxOpts_AutoTx(t *testing.T) {
	ethClient, err := client.NewETHClient("https://ethereum.publicnode.com")
	if err != nil {
		t.Fatal(err)
	}
	config := createConfig()
	config.TxType = "auto"
	calculator := NewGasFeeCalculator(ethClient, config)
	txOpts := &bind.TransactOpts{}
	if err = calculator.Apply(context.Background(), txOpts); err != nil {
		t.Fatal(err)
	}
	if txOpts.GasTipCap != nil {
		t.Error("gasTipCap must be nil")
	}
	if txOpts.GasFeeCap != nil {
		t.Error("gasFeeCap must be nil")
	}
	if txOpts.GasPrice != nil {
		t.Error("gasPrice must be nil")
	}
}

func createConfig() *ChainConfig {
	return &ChainConfig{
		TxType: "dynamic",
		DynamicTxGasConfig: &DynamicTxGasConfig{
			LimitPriorityFeePerGas: "1ether",
			PriorityFeeRate: &Fraction{
				Numerator:   1,
				Denominator: 1,
			},
			LimitFeePerGas: "1ether",
			// https://github.com/ethereum/go-ethereum/blob/0b471c312a82adf172bf6efdc7e3fdf285c62fba/accounts/abi/bind/base.go#L35
			BaseFeeRate: &Fraction{
				Numerator:   2,
				Denominator: 1,
			},
			//https://github.com/NomicFoundation/hardhat/blob/197118fb9f92034d250e7e7d12f69e28f960d3b1/packages/hardhat-core/src/internal/core/providers/gas-providers.ts#L248
			FeeHistoryRewardPercentile: 50,
			MaxRetryForFeeHistory:      1,
		},
	}
}

func Test_getFeeInfo(t *testing.T) {
	feeHistory := &ethereum.FeeHistory{}
	if _, _, ok := getFeeInfo(feeHistory); ok {
		t.Fatal("must be error")
	}
	feeHistory.Reward = append(feeHistory.Reward, []*big.Int{})
	if _, _, ok := getFeeInfo(feeHistory); ok {
		t.Fatal("must be error")
	}
	feeHistory.Reward[0] = append(feeHistory.Reward[0], big.NewInt(0))
	if _, _, ok := getFeeInfo(feeHistory); ok {
		t.Fatal("must be error")
	}
	feeHistory.Reward[0][0] = big.NewInt(1)
	if _, _, ok := getFeeInfo(feeHistory); ok {
		t.Fatal("must be error")
	}
	feeHistory.BaseFee = append(feeHistory.BaseFee, big.NewInt(2))
	if gasTip, baseFee, ok := getFeeInfo(feeHistory); ok {
		if gasTip.Int64() != int64(1) {
			t.Error("invalid gasTip", gasTip.Int64())
		}
		if baseFee.Int64() != int64(2) {
			t.Error("invalid baseFee", baseFee.Int64())
		}
	} else {
		t.Fatal("unexpected")
	}

}

type MockETHClient struct {
	client.IETHClient
	MockSuggestGasPrice big.Int
	MockPendingTransaction *txpool.RPCTransaction
	MockLatestHeaderNumber big.Int
	MockHistoryGasTipCap big.Int
	MockHistoryGasFeeCap big.Int
}
func (cl *MockETHClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return &cl.MockSuggestGasPrice, nil
}

func inclByPercent(n *big.Int, percent uint64) {
	n.Mul(n, big.NewInt(int64(100+percent)))
	n.Div(n, big.NewInt(100))
}

func (cl *MockETHClient) GetMinimumRequiredFee(ctx context.Context, address common.Address, nonce uint64, priceBump uint64) (*txpool.RPCTransaction, *big.Int, *big.Int, error) {
	gasFeeCap := new(big.Int).Set(cl.MockPendingTransaction.GasFeeCap.ToInt())
	gasTipCap := new(big.Int).Set(cl.MockPendingTransaction.GasTipCap.ToInt())

	inclByPercent(gasFeeCap, priceBump)
	inclByPercent(gasTipCap, priceBump)

	return cl.MockPendingTransaction, gasFeeCap, gasTipCap, nil
}

func (cl *MockETHClient) HeaderByNumber(ctx context.Context, number *big.Int) (*gethtypes.Header, error) {
	if number != nil {
		return &gethtypes.Header{
			Number: big.NewInt(0).Set(number),
		}, nil
	} else {
		return &gethtypes.Header{
			Number: &cl.MockLatestHeaderNumber,
		}, nil
	}
}
func (cl *MockETHClient) FeeHistory(ctx context.Context, blockCount uint64, lastBlock *big.Int, rewardPercentiles []float64) (*ethereum.FeeHistory, error) {
	return &ethereum.FeeHistory{
		Reward: [][]*big.Int{ // gasTipCap
			{ &cl.MockHistoryGasTipCap, },
		},
		BaseFee: []*big.Int{ // baseFee. This is used as gasFeeCap
			&cl.MockHistoryGasFeeCap,
		},
	}, nil
}

func TestPriceBumpLegacy(t *testing.T) {
	cli := MockETHClient{}
	config := createConfig()
	config.TxType = TxTypeLegacy
	config.PriceBump = 10
	calculator := NewGasFeeCalculator(&cli, config)

	txOpts := &bind.TransactOpts{}
	txOpts.Nonce = big.NewInt(1)

	cli.MockPendingTransaction = &txpool.RPCTransaction{
		GasPrice: (*hexutil.Big)(big.NewInt(100)),
		GasTipCap: (*hexutil.Big)(big.NewInt(200)),
		GasFeeCap: (*hexutil.Big)(big.NewInt(300)),
		Nonce: (hexutil.Uint64)(txOpts.Nonce.Uint64()),
	}

	// test that gasPrice is bumped from old tx's gasFeeCap
	{
		cli.MockSuggestGasPrice.SetUint64(100)
		if err := calculator.Apply(context.Background(), txOpts); err != nil {
			t.Fatal(err)
		}
		if txOpts.GasPrice.Uint64() != 330 { //gasFeeCap * 1.1
			t.Errorf("gasPrice should be 330 but %v", txOpts.GasPrice)
		}
	}

	// test that old tx's gasPrice is already exceeds suggestion
	{
		cli.MockSuggestGasPrice.SetUint64(99)
		err := calculator.Apply(context.Background(), txOpts)
		if err == nil || err.Error() != "old tx's gasPrice(100) is higher than suggestion(99)" {
			t.Fatal(err)
		}
	}
}

func TestPriceBumpDynamic(t *testing.T) {
	cli := MockETHClient{}
	config := createConfig()
	config.TxType = TxTypeDynamic
	config.PriceBump = 100 //double
	config.DynamicTxGasConfig.LimitFeePerGas = "1ether"
	config.DynamicTxGasConfig.LimitPriorityFeePerGas = "1ether"
	config.DynamicTxGasConfig.BaseFeeRate = &Fraction{
		Numerator: 1,
		Denominator: 1,
	}
	config.DynamicTxGasConfig.PriorityFeeRate = &Fraction{
		Numerator: 1,
		Denominator: 1,
	}

	calculator := NewGasFeeCalculator(&cli, config)

	txOpts := &bind.TransactOpts{}
	txOpts.Nonce = big.NewInt(1)

	cli.MockLatestHeaderNumber.SetUint64(1000)
	cli.MockPendingTransaction = &txpool.RPCTransaction{
		GasPrice: (*hexutil.Big)(big.NewInt(100)),
		GasTipCap: (*hexutil.Big)(big.NewInt(200)),
		GasFeeCap: (*hexutil.Big)(big.NewInt(300)),
		Nonce: (hexutil.Uint64)(txOpts.Nonce.Uint64()),
	}

	// test that gasTipCap and gasFeeCap are bumped from old tx's one
	{
		// set suggenstion between old value and bump value to apply bump value
		cli.MockHistoryGasTipCap.SetUint64(201)
		cli.MockHistoryGasFeeCap.SetUint64(301 - 201) // note that gasTipCap is added to gasFeeCap
		if err := calculator.Apply(context.Background(), txOpts); err != nil {
			t.Fatal(err)
		}
		if txOpts.GasTipCap.Uint64() != 400 {
			t.Errorf("gasTipCap should be 400 but %v", txOpts.GasTipCap)
		}
		if txOpts.GasFeeCap.Uint64() != 600 {
			t.Errorf("gasFeeCap should be 600 but %v", txOpts.GasFeeCap)
		}
	}

	// test that old tx's gasTipCap is already exceeds suggestion
	{
		cli.MockHistoryGasTipCap.SetUint64(199)
		err := calculator.Apply(context.Background(), txOpts)
		if err == nil || err.Error() != "old tx's gasTipCap(200) is higher than suggestion(199)" {
			t.Fatal(err)
		}
	}

	// test that old tx's gasFeeCap is already exceeds suggestion
	{
		// Because gasTipCap suggestion is added to gasFeeCap suggenstion,
		// gasTipCap suggestion should be lower than old tx's gasFeeCap
		cli.MockPendingTransaction.GasTipCap = (*hexutil.Big)(big.NewInt(100))
		cli.MockPendingTransaction.GasFeeCap = (*hexutil.Big)(big.NewInt(300))
		cli.MockHistoryGasTipCap.SetUint64(199)
		cli.MockHistoryGasFeeCap.SetUint64(299 - 200)
		err := calculator.Apply(context.Background(), txOpts)
		if err == nil || err.Error() != "old tx's gasFeeCap(300) is higher than suggestion(299)" {
			t.Fatal(err)
		}
	}
}
