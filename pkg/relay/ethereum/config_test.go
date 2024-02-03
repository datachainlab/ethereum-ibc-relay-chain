package ethereum

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestAllowLCFunction(t *testing.T) {
	type Input struct {
		addr     common.Address
		selector [4]byte
	}

	var cases = []struct {
		in      Input
		alf     AllowLCFunctions
		allowed bool
	}{
		{
			in: Input{addr: genAddr("A"), selector: [4]byte{0x01, 0x02, 0x03, 0x04}},
			alf: AllowLCFunctions{
				LCAddress: genAddr("A"),
				AllowALL:  false,
			},
			allowed: false,
		},
		{
			in: Input{addr: genAddr("A"), selector: [4]byte{0x01, 0x02, 0x03, 0x04}},
			alf: AllowLCFunctions{
				LCAddress: genAddr("A"),
				AllowALL:  true,
			},
			allowed: true,
		},
		{
			in: Input{addr: genAddr("A"), selector: [4]byte{0x01, 0x02, 0x03, 0x04}},
			alf: AllowLCFunctions{
				LCAddress: genAddr("B"),
				AllowALL:  true,
			},
			allowed: false,
		},
		{
			in: Input{addr: genAddr("A"), selector: [4]byte{0x01, 0x02, 0x03, 0x04}},
			alf: AllowLCFunctions{
				LCAddress: genAddr("B"),
				AllowALL:  false,
				Selectors: [][4]byte{{0x01, 0x02, 0x03, 0x04}},
			},
		},
		{
			in: Input{addr: genAddr("A"), selector: [4]byte{0x01, 0x02, 0x03, 0x04}},
			alf: AllowLCFunctions{
				LCAddress: genAddr("A"),
				AllowALL:  false,
				Selectors: [][4]byte{{0x01, 0x02, 0x03, 0x04}},
			},
			allowed: true,
		},
		{
			in: Input{addr: genAddr("A"), selector: [4]byte{0x01, 0x02, 0x03, 0x04}},
			alf: AllowLCFunctions{
				LCAddress: genAddr("A"),
				AllowALL:  false,
				Selectors: [][4]byte{{0x01, 0x02, 0x03, 0x05}},
			},
			allowed: false,
		},
		{
			in: Input{addr: genAddr("A"), selector: [4]byte{0x01, 0x02, 0x03, 0x04}},
			alf: AllowLCFunctions{
				LCAddress: genAddr("A"),
				AllowALL:  false,
				Selectors: [][4]byte{{0x01, 0x02, 0x03, 0x04}, {0x01, 0x02, 0x03, 0x05}},
			},
			allowed: true,
		},
	}
	for i, c := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			require := require.New(t)
			require.Equal(c.allowed, c.alf.IsAllowed(c.in.addr, c.in.selector))
		})
	}
}

func genAddr(preimage string) common.Address {
	bz := make([]byte, 20)
	h := sha256.Sum256([]byte(preimage))
	copy(bz, h[:20])
	return common.BytesToAddress(bz)
}
