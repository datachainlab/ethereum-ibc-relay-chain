package ethereum

import (
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestRevertReasonParserDefault(t *testing.T) {
	customErrors := []abi.Error{}

	erepo, err := NewErrorRepository(customErrors)
	require.NoError(t, err)

	revertReason, err := erepo.ParseError(
		common.FromHex("0x08c379a00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000001a4e6f7420656e6f7567682045746865722070726f76696465642e000000000000"),
	)
	require.NoError(t, err)
	require.Equal(t, "Not enough Ether provided.", revertReason)
}

func TestRevertReasonParserAddedCustomError(t *testing.T) {
	uintT, err := abi.NewType("uint256", "", nil)
	if err != nil {
		panic(err)
	}
	customErrors := []abi.Error{
		abi.NewError("AppError", abi.Arguments{{Name: "x", Type: uintT}}),
	}

	erepo, err := NewErrorRepository(customErrors)
	require.NoError(t, err)

	revertReason, err := erepo.ParseError(
		common.FromHex("0x08c379a00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000001a4e6f7420656e6f7567682045746865722070726f76696465642e000000000000"),
	)
	require.NoError(t, err)
	require.Equal(t, "Not enough Ether provided.", revertReason)
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
		abi.NewError("AppError", abi.Arguments{{Name: "x", Type: strT}}),
		abi.NewError("InsufficientBalance", abi.Arguments{{Name: "y", Type: uintT}, {Name: "z", Type: uintT}}),
	}

	erepo, err := NewErrorRepository(customErrors)
	require.NoError(t, err)

	revertReason, err := erepo.ParseError(
		common.FromHex("0xcf47918100000000000000000000000000000000000000000000000000000000000000070000000000000000000000000000000000000000000000000000000000000009"),
	)
	require.NoError(t, err)
	require.Equal(t, `InsufficientBalance{"y":7,"z":9}`, revertReason)
}
