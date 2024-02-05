package utils

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/params"
)

func ParseEtherAmount(amount string) (*big.Int, error) {
	var unit *big.Int
	switch {
	case strings.HasSuffix(amount, "ether"):
		amount = amount[:len(amount)-5]
		unit = big.NewInt(params.Ether)
	case strings.HasSuffix(amount, "gwei"): // Since "gwei" also includes "wei" as suffix, the "gwei" case must precedes the "wei" case.
		amount = amount[:len(amount)-4]
		unit = big.NewInt(params.GWei)
	case strings.HasSuffix(amount, "wei"):
		amount = amount[:len(amount)-3]
		unit = big.NewInt(params.Wei)
	default:
		return nil, fmt.Errorf("amount(%s) has invalid unit (acceptable: wei/gwei/ether)", amount)
	}

	am, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return nil, fmt.Errorf("amount(%s) has invalid quantity", amount)
	}

	return am.Mul(am, unit), nil
}
