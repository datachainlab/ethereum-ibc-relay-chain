package ethereum

import (
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/stretchr/testify/require"
)

func TestRevertReasonParserDefault(t *testing.T) {
	customErrors := []abi.Error{}

	erepo, err := NewErrorsRepository(customErrors)
	require.NoError(t, err)
	s, args, err := erepo.ParseError(
		hexToBytes("0x08c379a00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000001a4e6f7420656e6f7567682045746865722070726f76696465642e000000000000"),
	)
	require.NoError(t, err)
	require.Equal(t, "Error(string)", s)
	require.Equal(t, []interface{}{"Not enough Ether provided."}, args)
}

func TestRevertReasonParserAddedCustomError(t *testing.T) {
	uintT, err := abi.NewType("uint256", "", nil)
	if err != nil {
		panic(err)
	}
	customErrors := []abi.Error{
		{
			Name: "AppError",
			Inputs: []abi.Argument{
				{
					Type: uintT,
				},
			},
		},
	}

	erepo, err := NewErrorsRepository(customErrors)
	require.NoError(t, err)
	s, args, err := erepo.ParseError(
		hexToBytes("0x08c379a00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000001a4e6f7420656e6f7567682045746865722070726f76696465642e000000000000"),
	)
	require.NoError(t, err)
	require.Equal(t, "Error(string)", s)
	require.Equal(t, []interface{}{"Not enough Ether provided."}, args)
}

func TestRevertReasonParserCustomError(t *testing.T) {
	strT, err := abi.NewType("string", "", nil)
	if err != nil {
		panic(err)
	}
	uintT, err := abi.NewType("uint256", "", nil)
	if err != nil {
		panic(err)
	}
	customErrors := []abi.Error{
		{
			Name: "AppError",
			Inputs: []abi.Argument{
				{
					Type: strT,
				},
			},
		},
		{
			Name: "InsufficientBalance",
			Inputs: []abi.Argument{
				{
					Type: uintT,
				},
				{
					Type: uintT,
				},
			},
		},
	}

	erepo, err := NewErrorsRepository(customErrors)
	require.NoError(t, err)
	s, args, err := erepo.ParseError(
		hexToBytes("0xcf47918100000000000000000000000000000000000000000000000000000000000000070000000000000000000000000000000000000000000000000000000000000009"),
	)
	require.NoError(t, err)
	require.Equal(t, "InsufficientBalance(uint256,uint256)", s)
	require.Equal(t, []interface{}{big.NewInt(7), big.NewInt(9)}, args)
}

func hexToBytes(s string) []byte {
	reason, err := hex.DecodeString(strings.TrimPrefix(s, "0x"))
	if err != nil {
		panic(err)
	}
	return reason
}
