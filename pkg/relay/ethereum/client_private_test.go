package ethereum

import (
	"github.com/ethereum/go-ethereum"
	"math/big"
	"testing"
)

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
